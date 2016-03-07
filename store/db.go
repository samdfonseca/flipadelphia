package store

import (
	"bytes"
	"fmt"

	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/utils"
)

type FlipadelphiaDB struct {
	db bolt.DB
}

type FlipadelphiaFeature struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

type FlipadelphiaScopeFeatures []string

func createBucket(db bolt.DB, bucketName []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	createLog := fmt.Sprintf("CREATE BUCKET - Name: %q", bucketName)
	utils.LogEither(err, fmt.Sprint(createLog), fmt.Sprint(createLog), true)
	return err
}

func NewFlipadelphiaDB(db bolt.DB) FlipadelphiaDB {
	err := db.View(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("features")) != nil {
			return nil
		}
		return fmt.Errorf("Bucket \"features\" already exists")
	})
	if err != nil {
		if err := createBucket(db, []byte("features")); err != nil {
			utils.FailOnError(err, "EXITING - Unable to create required bucket", false)
		}
	}
	return FlipadelphiaDB{db: db}
}

func NewFlipadelphiaFeature(key []byte, value []byte) FlipadelphiaFeature {
	data := string(value) != ""
	return FlipadelphiaFeature{
		Name:  string(key),
		Value: string(value),
		Data:  fmt.Sprint(data),
	}
}

func (fdb FlipadelphiaDB) getScopeKeyValues(scope []byte) (map[string][]byte, error) {
	keys := make(map[string][]byte)
	err := fdb.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("features")).Cursor()
		for key, val := cursor.Seek(scope); bytes.HasPrefix(key, scope); key, val = cursor.Next() {
			splits := bytes.SplitN(key, []byte(":"), 2)
			keys[string(splits[1])] = val
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (fdb FlipadelphiaDB) getScopeKeyValuesWithCertainValue(scope []byte, targetValue []byte) (map[string][]byte, error) {
	keys, err := fdb.getScopeKeyValues(scope)
	if err != nil {
		return keys, err
	}
	for key, val := range keys {
		if !bytes.Equal(targetValue, val) {
			delete(keys, key)
		}
	}
	return keys, err
}

func (fdb FlipadelphiaDB) Set(scope []byte, key []byte, value []byte) (Serializable, error) {
	err := fdb.db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		scopeKey := bytes.Join([][]byte{scope, key}, []byte(":"))
		err := bucket.Put(scopeKey, value)
		if err != nil {
			return err
		}
		return nil
	})
	return NewFlipadelphiaFeature(key, value), err
}

func (fdb FlipadelphiaDB) Get(scope []byte, key []byte) (Serializable, error) {
	var value []byte
	var resultBuffer bytes.Buffer
	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		mergedScopeKey := bytes.Join([][]byte{scope, key}, []byte(":"))
		resultBuffer.Write(bucket.Get(mergedScopeKey))
		value = resultBuffer.Bytes()
		return nil
	})
	return NewFlipadelphiaFeature(key, value), err
}

func (fdb FlipadelphiaDB) GetScopeFeatures(scope []byte) (Serializable, error) {
	var featureList FlipadelphiaScopeFeatures
	scopeKeys, err := fdb.getScopeKeyValues(scope)
	if err != nil {
		return featureList, err
	}
	for key, _ := range scopeKeys {
		featureList = append(featureList, key)
	}
	return featureList, err
}

func (fdb FlipadelphiaDB) GetScopeFeaturesFilterByValue(scope []byte, value []byte) (Serializable, error) {
	var featureList FlipadelphiaScopeFeatures
	scopeKeys, err := fdb.getScopeKeyValuesWithCertainValue(scope, value)
	if err != nil {
		return featureList, err
	}
	for key, _ := range scopeKeys {
		featureList = append(featureList, key)
	}
	return featureList, err
}

func (feature FlipadelphiaFeature) Serialize() []byte {
	serializedFeature, err := json.Marshal(feature)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize feature", true)
		return []byte("")
	}
	return serializedFeature
}

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
