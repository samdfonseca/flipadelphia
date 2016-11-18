package store

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/config"
	"github.com/samdfonseca/flipadelphia/utils"
)

type KeyValuePair [][]byte

type Serializable interface {
	Serialize() []byte
}

type PersistenceStore interface {
	Get([]byte, []byte) (Serializable, error)
	GetScopeFeatures([]byte) (Serializable, error)
	GetScopeFeaturesFilterByValue([]byte, []byte) (Serializable, error)
	Set([]byte, []byte, []byte) (Serializable, error)
	GetScopes() (Serializable, error)
	GetScopesWithPrefix([]byte) (Serializable, error)
	GetScopesWithFeature([]byte) (Serializable, error)
	GetScopesPaginated(int, int) (Serializable, error)
	GetFeatures() (Serializable, error)
	GetScopeFeaturesFull([]byte) (Serializable, error)
	Close() error
}

// FlipadelphiaFeature holds the name, value and data attributes of a feature.
type FlipadelphiaFeature struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

// FlipadelphiaFeatures is a type alias for []FlipadelphiaFeature
type FlipadelphiaFeatures []FlipadelphiaFeature

// FlipadelphiaSetFeatureOptions is a helper struct to store the values needed to set a feature.
type FlipadelphiaSetFeatureOptions struct {
	Key   string
	Scope string `json:"scope"`
	Value string `json:"value"`
}

// FlipadelphiaScopeFeatures is a type alias for []string.
type FlipadelphiaScopeFeatures []string

// FlipadelphiaScopeList is a type alias for []string.
type FlipadelphiaScopeList []string

type StringSlice []string

var validFeatureKeyCharacters = []byte(`abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-`)

func NewPersistenceStore(c config.FlipadelphiaConfig) PersistenceStore {
	// var ps PersistenceStore
	switch c.PersistenceStoreType {
	case "bolt":
		db, err := bolt.Open(c.DBFile, 0600, nil)
		utils.FailOnError(err, "Unable to open db file", true)
		ps := NewFlipadelphiaBoltDB(db)
		utils.Output(fmt.Sprintf("Using BoltDB persistence store: %s", c.DBFile))
		return ps
	case "redis":
		err := fmt.Errorf("Unable to connect to Redis")
		if c.RedisHost == "" {
			utils.FailOnError(err, "redis_host not set", true)
		}
		ps := NewFlipadelphiaRedisDB(c.RedisHost, c.RedisPassword, c.RedisDB)
		utils.Output(fmt.Sprintf("Using Redis persistence store: %s", c.RedisHost))
		return ps
	case "redisv2":
		err := fmt.Errorf("Unable to connect to Redis")
		if c.RedisHost == "" {
			utils.FailOnError(err, "redis_host not set", true)
		}
		ps := NewFlipadelphiaRedisDBV2(c.RedisHost, c.RedisPassword)
		utils.Output(fmt.Sprintf("Using RedisV2 persistence store: %s", c.RedisHost))
		return ps
	}
	return nil
}

// NewFlipadelphiaFeature returns a new instance of FlipadelphiaFeature.
func NewFlipadelphiaFeature(key []byte, value []byte) FlipadelphiaFeature {
	data := string(value) != ""
	return FlipadelphiaFeature{
		Name:  string(key),
		Value: string(value),
		Data:  fmt.Sprint(data),
	}
}

func (ss StringSlice) Serialize() []byte {
	serializedStringSlice, err := json.Marshal(ss)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize string slice", true)
		return []byte("")
	}
	return serializedStringSlice
}

// Serialize returns the FlipadelphiaFeature as json.
func (feature FlipadelphiaFeature) Serialize() []byte {
	serializedFeature, err := json.Marshal(feature)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize feature", true)
		return []byte("")
	}
	return serializedFeature
}

// Serialize returns the FlipadelphiaScopeFeatures as json.
func (features FlipadelphiaScopeFeatures) Serialize() []byte {
	if features == nil {
		return []byte("[]")
	}
	serializedFeatures, err := json.Marshal(features)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize features", true)
		return []byte("")
	}
	return serializedFeatures
}

// Serialize returns the FlipadelphiaScopeList as json.
func (scopes FlipadelphiaScopeList) Serialize() []byte {
	if scopes == nil {
		return []byte("[]")
	}
	serializedScopes, err := json.Marshal(scopes)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize scopes", true)
		return []byte("")
	}
	return serializedScopes
}

// Serialize returns the []FlipadelphiaFeature as json.
func (ffs FlipadelphiaFeatures) Serialize() []byte {
	if ffs == nil {
		return []byte("[]")
	}
	serializedFeatures, err := json.Marshal(ffs)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize features", true)
		return []byte("")
	}
	return serializedFeatures
}

// func (s []string) Serialize() []byte {
// 	if s == nil {
// 		return []byte("[]")
// 	}
// 	serialized, err := json.Marshal(s)
// 	if err != nil {
// 		utils.LogOnError(err, "Unable to serialize strings", true)
// 		return []byte("")
// 	}
// 	return serialized
// }
