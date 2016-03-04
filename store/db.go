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

type FlipadelphiaFeatures struct {
	Scope    string `json:"scope"`
	Features []FlipadelphiaFeature
}

func createBucketOrFail(db bolt.DB, bucketName []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	createLog := fmt.Sprintf("CREATE BUCKET - Name: %q", bucketName)
	utils.LogEither(err, fmt.Sprint(createLog), fmt.Sprint(createLog), true)
	utils.FailOnError(err, "EXITING - Unable to create required bucket", false)
}

func NewFlipadelphiaDB(db bolt.DB) FlipadelphiaDB {
	err := db.View(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("features")) != nil {
			return nil
		}
		return fmt.Errorf("Bucket \"features\" already exists")
	})
	if err != nil {
		createBucketOrFail(db, []byte("features"))
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

func (fdb FlipadelphiaDB) getScopeKeys(scope []byte) ([][]byte, error) {
	var keys [][]byte
	err := fdb.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("features")).Cursor()
		for key, _ := cursor.Seek(scope); bytes.HasPrefix(key, scope); key, _ = cursor.Next() {
			splits := bytes.SplitN(key, []byte(":"), 2)
			keys = append(keys, splits[1])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
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
	// setLog := fmt.Sprintf("GET - Feature: %q, Scope: %q, Value: %q", key, scope, value)
	// utils.LogEither(err, fmt.Sprintf("SUCCESS %s", setLog), fmt.Sprintf("FAIL %s", setLog), true)
	return NewFlipadelphiaFeature(key, value), err
}

func (fdb FlipadelphiaDB) GetScopeFeatures(scope []byte) (Serializable, error) {
	var featureList []FlipadelphiaFeature
	features := FlipadelphiaFeatures{
		Scope:    string(scope),
		Features: featureList,
	}
	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopeKeys, err := fdb.getScopeKeys(scope)
		if err != nil {
			return err
		}
		for i := range scopeKeys {
			featureList = append(featureList, FlipadelphiaFeature{Name: string(scopeKeys[i])})
		}
		return nil
	})
	return features, err
}

func (fdb FlipadelphiaDB) addToFeaturesBucket(scope []byte, key []byte, value []byte) (err error) {
	err = fdb.db.Batch(func(tx *bolt.Tx) error {
		featureBucket, err := tx.Bucket([]byte("features")).CreateBucketIfNotExists(key)
		if err != nil {
			return err
		}
		if err = featureBucket.Put(scope, value); err != nil {
			return err
		}
		return nil
	})
	utils.LogEither(err,
		fmt.Sprintf("Added feature %q to features bucket", key),
		fmt.Sprintf("Unable to add feature %q to features bucket", key),
		true)
	return err
}

func getAllKeyValPairsFromBucket(bucket bolt.Bucket) []KeyValuePair {
	var keyValuePairs []KeyValuePair
	bucket.ForEach(func(key, val []byte) error {
		if val != nil {
			keyValuePairs = append(keyValuePairs, KeyValuePair{key, val})
		}
		return nil
	})
	return keyValuePairs
}

func (feature FlipadelphiaFeature) Serialize() []byte {
	serializedFeature, err := json.Marshal(feature)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize feature", true)
		return []byte("")
	}
	return serializedFeature
}

func (features FlipadelphiaFeatures) Serialize() []byte {
	serializedFeatures, err := json.Marshal(features.Features)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize features", true)
		return []byte("")
	}
	return serializedFeatures
}

//func (features [][]byte) Serialize() []byte {
//	serializedFeatures, err := json.Marshal(features)
//	if err != nil {
//		utils.LogOnError(err, "Unable to serialize features", true)
//		return []byte("")
//	}
//	return serializedFeatures
//}
