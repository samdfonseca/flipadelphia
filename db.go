package main

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
)

type FlipadelphiaDB struct {
	db bolt.DB
}

type FlipadelphiaFeature struct {
	Name  string
	Value string
}

type KeyValuePair [][]byte

var FDB FlipadelphiaDB

func (FDB *FlipadelphiaDB) Set(scope []byte, key []byte, value []byte) error {
	err := FDB.db.Batch(func(tx *bolt.Tx) error {
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

func (FDB *FlipadelphiaDB) Get(scope []byte, key []byte) (feature FlipadelphiaFeature, err error) {
	var value []byte
	err = FDB.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(scope)
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %q", scope)
		}
		var resultBuffer bytes.Buffer
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

func (FDB *FlipadelphiaDB) GetAll(scope []byte) (features []FlipadelphiaFeature, err error) {
	err = FDB.db.View(func(tx *bolt.Tx) error {
		topLevelBucket := tx.Bucket(scope)
		if topLevelBucket == nil {
			return fmt.Errorf("Bucket not found: %q", scope)
		}
		var keyValPairBuffer []KeyValuePair
		getAllKeyValPairsFromBucket(*topLevelBucket, keyValPairBuffer)
		for i := range keyValPairBuffer {
			features = append(features, FlipadelphiaFeature{
				Name: string(keyValPairBuffer[i][0]),
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
