package store

import (
	"fmt"

	"gopkg.in/redis.v5"
)

type FlipadelphiaRedisDB struct {
	client *redis.Client
}

func NewFlipadelphiaRedisDB(host, password string, db int) FlipadelphiaRedisDB {
	return FlipadelphiaRedisDB{
		client: redis.NewClient(&redis.Options{
			Addr:     host,
			Password: password,
			DB:       db,
		}),
	}
}

func (rdb FlipadelphiaRedisDB) Get(scope, key []byte) (Serializable, error) {
	value, err := rdb.client.HGet(string(scope), string(key)).Bytes()
	if err != nil {
		return nil, err
	}
	return NewFlipadelphiaFeature(key, value), nil
}

func (rdb FlipadelphiaRedisDB) Set(scope, key, value []byte) (Serializable, error) {
	err := rdb.client.HSet(string(scope), string(key), string(value)).Err()
	return NewFlipadelphiaFeature(key, value), err
}

func (rdb FlipadelphiaRedisDB) GetScopeFeatures(scope []byte) (Serializable, error) {
	var keys FlipadelphiaScopeFeatures
	keys, err := rdb.client.HKeys(string(scope)).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (rdb FlipadelphiaRedisDB) GetScopeFeaturesFilterByValue(scope []byte, targetValue []byte) (Serializable, error) {
	var features FlipadelphiaScopeFeatures
	res, err := rdb.client.HGetAll(string(scope)).Result()
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

func (rdb FlipadelphiaRedisDB) GetScopes() (Serializable, error) {
	var scopes FlipadelphiaScopeList
	var cursor uint64
	for {
		keys, cursor, err := rdb.client.Scan(cursor, "", 10).Result()
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, keys...)
		if cursor == 0 {
			break
		}
	}
	return scopes, nil
}

func (rdb FlipadelphiaRedisDB) GetScopesWithPrefix(prefix []byte) (Serializable, error) {
	var scopes FlipadelphiaScopeList
	var cursor uint64
	match := string(prefix) + "*"
	for {
		keys, cursor, err := rdb.client.Scan(cursor, match, 10).Result()
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, keys...)
		if cursor == 0 {
			break
		}
	}
	return scopes, nil
}

func (rdb FlipadelphiaRedisDB) GetScopesWithFeature(key []byte) (Serializable, error) {
	var scopesWithFeature FlipadelphiaScopeList
	scopes, err := rdb.GetScopes()
	if err != nil {
		return nil, err
	}
	for _, scope := range scopes.(FlipadelphiaScopeList) {
		if res, err := rdb.client.HExists(scope, string(key)).Result(); err != nil && res == true {
			scopesWithFeature = append(scopesWithFeature, scope)
		}
	}
	return scopesWithFeature, nil
}

func (rdb FlipadelphiaRedisDB) GetScopesPaginated(offset, count int) (Serializable, error) {
	var scopes StringSlice
	return scopes, fmt.Errorf("Unimplemented method")
}

func (rdb FlipadelphiaRedisDB) GetFeaturesPaginated(offset, count int) (Serializable, error) {
	return nil, fmt.Errorf("Unimplemented method")
}

func (rdb FlipadelphiaRedisDB) GetFeatures() (Serializable, error) {
	scopes, err := rdb.GetScopes()
	fch := make(chan FlipadelphiaScopeFeatures)
	ech := make(chan error)
	for _, scope := range scopes.(FlipadelphiaScopeList) {
		go func() {
			features, err := rdb.GetScopeFeatures([]byte(scope))
			if err != nil {
				ech <- err
				return
			}
			fch <- features.(FlipadelphiaScopeFeatures)
		}()
	}
	var featuresMap = make(map[string]interface{})
	var uniqueFeatures FlipadelphiaScopeFeatures
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

func (rdb FlipadelphiaRedisDB) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
	var features FlipadelphiaFeatures
	res, err := rdb.client.HGetAll(string(scope)).Result()
	if err != nil {
		return nil, err
	}
	for k, v := range res {
		features = append(features, NewFlipadelphiaFeature([]byte(k), []byte(v)))
	}
	return features, nil
}

func (rdb FlipadelphiaRedisDB) Close() error {
	return rdb.client.Close()
}

func (rdb FlipadelphiaRedisDB) CheckScopeExists(scope []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDB) CheckFeatureExists(feature []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDB) CheckScopeHasFeature(scope, feature []byte) bool {
	return true
}

func (rdb FlipadelphiaRedisDB) CheckFeatureHasScope(scope, feature []byte) bool {
	return true
}
