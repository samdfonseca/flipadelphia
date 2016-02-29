package db

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
}

type KeyValuePair [][]byte

var DB FlipadelphiaDB

func NewFlipadelphiaDB(db bolt.DB) FlipadelphiaDB {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("features"))
		return err
	})
	utils.FailOnError(err, "Unable to create features bucket", true)
	return FlipadelphiaDB{db: db}
}

func (fdb *FlipadelphiaDB) Set(scope []byte, key []byte, value []byte) error {
	err := fdb.db.Batch(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(scope)
		if err != nil {
			return err
		}
		err = bucket.Put(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (fdb *FlipadelphiaDB) Get(scope []byte, key []byte) (feature FlipadelphiaFeature, err error) {
	var value []byte
	var resultBuffer bytes.Buffer
	err = fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(scope)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %q", scope)
		}
		resultBuffer.Write(bucket.Get(key))
		value = resultBuffer.Bytes()
		return nil
	})
	feature = FlipadelphiaFeature{
		Name:  string(key),
		Value: string(value),
	}
	return
}

func (fdb *FlipadelphiaDB) GetAll(scope []byte) (features []FlipadelphiaFeature, err error) {
	err = fdb.db.View(func(tx *bolt.Tx) error {
		topLevelBucket := tx.Bucket(scope)
		if topLevelBucket == nil {
			return fmt.Errorf("Bucket not found: %q", scope)
		}
		var keyValPairBuffer []KeyValuePair
		getAllKeyValPairsFromBucket(*topLevelBucket, keyValPairBuffer)
		for i := range keyValPairBuffer {
			features = append(features, FlipadelphiaFeature{
				Name:  string(keyValPairBuffer[i][0]),
				Value: string(keyValPairBuffer[i][1]),
			})
		}
		return nil
	})
	return
}

func getAllKeyValPairsFromBucket(bucket bolt.Bucket, buffer []KeyValuePair) {
	// gets netsted buckets recursively
	cursor := bucket.Cursor()
	for key, val := cursor.First(); key != nil; key, val = cursor.Next() {
		if val != nil {
			buffer = append(buffer, KeyValuePair{key, val})
		} else {
			getAllKeyValPairsFromBucket(*bucket.Bucket(key), buffer)
		}
	}
	return
}

func (feature FlipadelphiaFeature) Serialize() []byte {
	featureMap := map[string]string{
		"name":  feature.Name,
		"value": feature.Value,
		"data":  "true",
	}
	serializedFeature, err := json.Marshal(featureMap)
	if err != nil {
		utils.LogOnError(err, "Unable to serialize feature", true)
		return []byte("")
	}
	return serializedFeature
}
