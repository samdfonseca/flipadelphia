package store

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisConnection interface {
	Close() error
	Err() error
	Do(string, ...interface{}) (interface{}, error)
	Send(string, ...interface{}) error
	Flush() error
	Receive() (interface{}, error)
}

type RedisConnectionPool interface {
	Get() redis.Conn
	Close() error
	ActiveCount() int
}

type FlipadelphiaRedisDBV2 struct {
	pool RedisConnectionPool
}

func NewFlipadelphiaRedisDBV2(server, password string, db int) FlipadelphiaRedisDBV2 {
	return FlipadelphiaRedisDBV2{
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server, redis.DialDatabase(db))
				if err != nil {
					return nil, err
				}
				if password != "" {
					if _, err := c.Do("AUTH", password); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (rdb FlipadelphiaRedisDBV2) get(conn RedisConnection, scope, key []byte) (Serializable, error) {
	value, err := redis.String(conn.Do("HGET", string(scope), string(key)))
	if err != nil {
		return nil, err
	}
	return NewFlipadelphiaFeature(key, []byte(value)), nil
}

func (rdb FlipadelphiaRedisDBV2) Get(scope, key []byte) (Serializable, error) {
	return rdb.get(rdb.pool.Get(), scope, key)
}

func (rdb FlipadelphiaRedisDBV2) set(conn RedisConnection, scope, key, value []byte) (Serializable, error) {
	_, err := conn.Do("HSET", string(scope), string(key), string(value))
	return NewFlipadelphiaFeature(key, value), err
}

func (rdb FlipadelphiaRedisDBV2) Set(scope, key, value []byte) (Serializable, error) {
	return rdb.set(rdb.pool.Get(), scope, key, value)
}

func (rdb FlipadelphiaRedisDBV2) checkValueExistsInSet(setKey, value []byte) (bool, error) {
	luaScript := `
	local features = redis.call('lrange', ARGV[1], '0', '-1');
	local existsInTable = function (t, v)
		local i = 1;
		while t[i] ~= nil do
			if t[i] == v then
				return true
			end;
			i = i + 1;
		end;
		return false
	end;
	return existsInTable(features, ARGV[2])`
	b, err := redis.Bool(rdb.pool.Get().Do("EVAL", luaScript, 0, string(setKey), string(value)))
	return b, err
}

func (rdb FlipadelphiaRedisDBV2) getScopeFeatures(conn RedisConnection, scope []byte) (Serializable, error) {
	var keys StringSlice
	keys, err := redis.Strings(conn.Do("HKEYS", string(scope)))
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopeFeatures(scope []byte) (Serializable, error) {
	return rdb.getScopeFeatures(rdb.pool.Get(), scope)
}

func (rdb FlipadelphiaRedisDBV2) getScopeFeaturesFilterByValue(conn RedisConnection, scope, targetValue []byte) (Serializable, error) {
	var features StringSlice
	res, err := redis.StringMap(conn.Do("HGETALL", string(scope)))
	if err != nil {
		return nil, err
	}
	for k, v := range res {
		if v == string(targetValue) {
			features = append(features, k)
		}
	}
	return features, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopeFeaturesFilterByValue(scope, targetValue []byte) (Serializable, error) {
	return rdb.getScopeFeaturesFilterByValue(rdb.pool.Get(), scope, targetValue)
}

func (rdb FlipadelphiaRedisDBV2) getScopes(conn RedisConnection) (Serializable, error) {
	var scopes StringSlice
	scopes, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		return nil, err
	}
	return scopes, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopes() (Serializable, error) {
	return rdb.getScopes(rdb.pool.Get())
}

func (rdb FlipadelphiaRedisDBV2) getScopesWithPrefix(conn RedisConnection, prefix []byte) (Serializable, error) {
	var scopes StringSlice
	match := "MATCH " + string(prefix) + "*"
	scopes, err := redis.Strings(conn.Do("KEYS", match))
	if err != nil {
		return nil, err
	}
	return scopes, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopesWithPrefix(prefix []byte) (Serializable, error) {
	return rdb.getScopesWithPrefix(rdb.pool.Get(), prefix)
}

func (rdb FlipadelphiaRedisDBV2) getScopesWithFeature(conn RedisConnection, key []byte) (Serializable, error) {
	scopes, err := rdb.getScopes(conn)
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes.(StringSlice) {
		if res, err := redis.Bool(conn.Do("HEXISTS", string(key))); err != nil && res == true {
			scopes = append(scopes.(StringSlice), scope)
		}
	}
	return scopes, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopesWithFeature(key []byte) (Serializable, error) {
	return rdb.getScopeFeatures(rdb.pool.Get(), key)
}

func (rdb FlipadelphiaRedisDBV2) GetScopesPaginated(offset, count int) (Serializable, error) {
	var scopes StringSlice
	return scopes, fmt.Errorf("Unimplemented method")
}

func (rdb FlipadelphiaRedisDBV2) GetFeaturesPaginated(offset, count int) (Serializable, error) {
	return nil, fmt.Errorf("Unimplemented method")
}

func (rdb FlipadelphiaRedisDBV2) getFeatures(conn RedisConnection) (Serializable, error) {
	scopes, err := rdb.getScopes(conn)
	fch := make(chan StringSlice)
	ech := make(chan error)
	for _, scope := range scopes.(StringSlice) {
		go func() {
			features, err := rdb.GetScopeFeatures([]byte(scope))
			if err != nil {
				ech <- err
				return
			}
			fch <- features.(StringSlice)
		}()
	}
	var featuresMap = make(map[string]interface{})
	var uniqueFeatures StringSlice
	for F := range fch {
		for _, f := range F {
			if _, ok := featuresMap[f]; !ok {
				featuresMap[f] = nil
				uniqueFeatures = append(uniqueFeatures, f)
			}
		}
	}
	err, ok := <-ech
	if ok {
		return nil, err
	}
	return uniqueFeatures, nil
}

func (rdb FlipadelphiaRedisDBV2) GetFeatures() (Serializable, error) {
	return rdb.getFeatures(rdb.pool.Get())
}

func (rdb FlipadelphiaRedisDBV2) getScopeFeaturesFull(conn RedisConnection, scope []byte) (Serializable, error) {
	var features FlipadelphiaFeatures
	res, err := redis.StringMap(conn.Do("HGETALL", string(scope)))
	if err != nil {
		return nil, err
	}
	for k, v := range res {
		features = append(features, NewFlipadelphiaFeature([]byte(k), []byte(v)))
	}
	return features, nil
}

func (rdb FlipadelphiaRedisDBV2) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
	return rdb.getScopeFeaturesFull(rdb.pool.Get(), scope)
}

func (rdb FlipadelphiaRedisDBV2) Close() error {
	return rdb.pool.Close()
}

func (rdb FlipadelphiaRedisDBV2) CheckScopeExists(scope []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDBV2) CheckFeatureExists(feature []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDBV2) CheckScopeHasFeature(scope, feature []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDBV2) CheckFeatureHasScope(scope, feature []byte) bool {
	return true
}
