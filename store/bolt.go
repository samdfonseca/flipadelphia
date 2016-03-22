package store

import (
	"bytes"
	"fmt"

	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/utils"
)

type FlipadelphiaDB struct {
	db         *bolt.DB
	bucketName string "features"
}

type FlipadelphiaFeature struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Data  string `json:"data"`
}

type FlipadelphiaSetFeatureOptions struct {
	Key   []byte
	Scope []byte `json:"scope"`
	Value []byte `json:"value"`
}

type pageIndexValue [][]byte

type PagerIndex struct {
	db *bolt.DB
}

type FlipadelphiaScopeFeatures []string

var validFeatureKeyCharacters = []byte(`abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-`)

func createBucket(db *bolt.DB, bucketName []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	createLog := fmt.Sprintf("CREATE BUCKET - Name: %q", bucketName)
	utils.LogEither(err, fmt.Sprint(createLog), fmt.Sprint(createLog), true)
	return err
}

func NewFlipadelphiaDB(db *bolt.DB) FlipadelphiaDB {
	requiredBuckets := [][]byte{
		[]byte("features"),
	}
	for _, bucket := range requiredBuckets {
		err := db.View(func(tx *bolt.Tx) error {
			if tx.Bucket(bucket) != nil {
				return nil
			}
			return fmt.Errorf(`Bucket "%s" already exists`, bucket)
		})
		if err != nil {
			if err := createBucket(db, bucket); err != nil {
				utils.FailOnError(err, fmt.Sprintf("EXITING - Unable to create required bucket '%s'", bucket), false)
			}
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

func MergeScopeKey(scope, key []byte) ([]byte, error) {
	if bytes.Contains(scope, []byte(":")) {
		return []byte{}, fmt.Errorf("Invalid scope: Can not contain ':' character")
	}
	for _, b := range key {
		if !bytes.Contains(validFeatureKeyCharacters, []byte{b}) {
			return []byte{}, fmt.Errorf("Invalid key character '%s': Valid characters are '%s'", string(b), validFeatureKeyCharacters)
		}
	}
	return bytes.Join([][]byte{scope, key}, []byte(":")), nil
}

func (fdb FlipadelphiaDB) getScopeKeyValues(scope []byte) (map[string][]byte, error) {
	keys := make(map[string][]byte)
	err := fdb.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("features")).Cursor()
		for key, val := cursor.Seek(scope); bytes.HasPrefix(key, scope); key, val = cursor.Next() {
			splits := bytes.SplitN(key, []byte(":"), 2)
			if bytes.Equal(scope, splits[0]) {
				keys[string(splits[1])] = val
			}
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
		scopeKey, err := MergeScopeKey(scope, key)
		if err != nil {
			return err
		}
		err = bucket.Put(scopeKey, value)
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
