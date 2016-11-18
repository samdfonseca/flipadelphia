package store

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/utils"
)

// FlipadelphiaBoltDB holds a pointer to the boltdb instance and the name of the main bucket.
type FlipadelphiaBoltDB struct {
	db         *bolt.DB
	bucketName string
}

func createBucket(db *bolt.DB, bucketName []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	createLog := fmt.Sprintf("CREATE BUCKET - Name: %q", bucketName)
	utils.LogOnError(err, fmt.Sprint(createLog), true)
	return err
}

// NewFlipadelphiaBoltDB creates a new instance of FlipadelphiaBoltDB. The "features" bucket is created
// if it does not yet exist.
func NewFlipadelphiaBoltDB(db *bolt.DB) FlipadelphiaBoltDB {
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
	return FlipadelphiaBoltDB{db: db, bucketName: "features"}
}

// mergeScopeKey joins two []byte around the ":" character.
func mergeScopeKey(scope, key []byte) ([]byte, error) {
	if bytes.Contains(scope, []byte(":")) {
		//noinspection GoPlaceholderCount
		return []byte{}, fmt.Errorf("Invalid scope: Can not contain ':' character")
	}
	for _, b := range key {
		if !bytes.Contains(validFeatureKeyCharacters, []byte{b}) {
			return []byte{}, fmt.Errorf("Invalid key character '%s': Valid characters are '%s'", string(b), validFeatureKeyCharacters)
		}
	}
	return bytes.Join([][]byte{scope, key}, []byte(":")), nil
}

// splitScopeKey splits a []byte on the first ":" character.
func splitScopeKey(scopeKey []byte) ([]byte, []byte, error) {
	if !bytes.Contains(scopeKey, []byte(":")) {
		//noinspection GoPlaceholderCount
		err := fmt.Errorf(`ScopeKey missing ":" character`)
		return []byte{}, []byte{}, err
	}
	splits := bytes.SplitN(scopeKey, []byte(":"), 2)
	return splits[0], splits[1], nil
}

func mustGetScopeFromScopeKey(scopeKey []byte) []byte {
	scope, _, _ := splitScopeKey(scopeKey)
	return scope
}

func mustGetKeyFromScopeKey(scopeKey []byte) []byte {
	_, key, _ := splitScopeKey(scopeKey)
	return key
}

func (fdb FlipadelphiaBoltDB) Close() error {
	return fdb.db.Close()
}

func (fdb FlipadelphiaBoltDB) getScopeKeyValues(scope []byte) (map[string][]byte, error) {
	keys := make(map[string][]byte)
	err := fdb.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("features")).Cursor()
		for key, val := cursor.Seek(scope); bytes.HasPrefix(append(key, ':'), scope) && key != nil; key, val = cursor.Next() {
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

func (fdb FlipadelphiaBoltDB) getScopeKeyValuesWithCertainValue(scope []byte, targetValue []byte) (map[string][]byte, error) {
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

func (fdb FlipadelphiaBoltDB) getAllScopes() (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList

	err := fdb.db.View(func(tx *bolt.Tx) error {
		var previousScope []byte

		bucket := tx.Bucket([]byte("features"))
		bucket.ForEach(func(key, val []byte) error {
			scope, _, err := splitScopeKey(key)
			if err == nil && !bytes.Equal(scope, previousScope) {
				scopes = append(scopes, fmt.Sprintf("%s", scope))
				previousScope = scope
			}
			return nil
		})
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getAllScopesWithPrefix(prefix []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList

	err := fdb.db.View(func(tx *bolt.Tx) error {
		var previousScope []byte

		cursor := tx.Bucket([]byte("features")).Cursor()
		for key, _ := cursor.Seek(prefix); bytes.HasPrefix(key, prefix); key, _ = cursor.Next() {
			scope, _, err := splitScopeKey(key)
			if err == nil && !bytes.Equal(scope, previousScope) {
				scopes = append(scopes, fmt.Sprintf("%s", scope))
				previousScope = scope
			}
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getScopesPaginated(offset, count int) (StringSlice, error) {
	var scopes StringSlice

	err := fdb.db.View(func(tx *bolt.Tx) error {
		var previousScope []byte

		cursor := tx.Bucket([]byte("features")).Cursor()
		key, _ := cursor.First()
		if key != nil {
			previousScope = mustGetScopeFromScopeKey(key)
		}
		for counter := 0; key != nil && offset != 0 && counter < offset; key, _ = cursor.Next() {
			scope, _, _ := splitScopeKey(key)
			for bytes.Equal(previousScope, scope) {
				key, _ = cursor.Next()
				scope = mustGetScopeFromScopeKey(key)
			}
			previousScope = scope
			counter++
		}
		for key != nil && len(scopes) < count {
			scope, _, _ := splitScopeKey(key)
			if len(scopes) == 0 || !bytes.Equal(scope, []byte(scopes[len(scopes)-1])) {
				scopes = append(scopes, string(scope))
			}
			key, _ = cursor.Next()
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getAllScopesWithFeature(feature []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList

	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		bucket.ForEach(func(scopeKey, val []byte) error {
			scope, key, err := splitScopeKey(scopeKey)
			if err == nil && bytes.Equal(feature, key) {
				scopes = append(scopes, fmt.Sprintf("%s", scope))
			}
			return nil
		})
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getAllFeatures() (FlipadelphiaScopeFeatures, error) {
	var features FlipadelphiaScopeFeatures

	err := fdb.db.View(func(tx *bolt.Tx) error {
		var previousFeature []byte

		bucket := tx.Bucket([]byte("features"))
		bucket.ForEach(func(key, val []byte) error {
			_, feature, err := splitScopeKey(key)
			if err == nil && !bytes.Equal(feature, previousFeature) {
				features = append(features, fmt.Sprintf("%s", feature))
				previousFeature = feature
			}
			return nil
		})
		return nil
	})
	sort.Strings(features)
	var uniqueFeatures FlipadelphiaScopeFeatures
	for i := range features {
		if i < len(features)-1 {
			if features[i] != features[i+1] {
				uniqueFeatures = append(uniqueFeatures, features[i])
			}
		} else {
			uniqueFeatures = append(uniqueFeatures, features[i])
		}
	}
	return uniqueFeatures, err
}

// Set stores the feature in the database and returns an instance of FlipadelphiaFeature.
func (fdb FlipadelphiaBoltDB) Set(scope []byte, key []byte, value []byte) (Serializable, error) {
	err := fdb.db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		scopeKey, err := mergeScopeKey(scope, key)
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

// Get retrieves the feature from the database and returns an instance of FlipadelphiaFeature.
func (fdb FlipadelphiaBoltDB) Get(scope []byte, key []byte) (Serializable, error) {
	var value []byte
	var resultBuffer bytes.Buffer

	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		mergedScopeKey, err := mergeScopeKey(scope, key)
		if err != nil {
			return err
		}
		resultBuffer.Write(bucket.Get(mergedScopeKey))
		value = resultBuffer.Bytes()
		return nil
	})
	return NewFlipadelphiaFeature(key, value), err
}

// GetScopeFeatures returns all features set on the given scope.
func (fdb FlipadelphiaBoltDB) GetScopeFeatures(scope []byte) (Serializable, error) {
	var featureList FlipadelphiaScopeFeatures

	scopeKeys, err := fdb.getScopeKeyValues(scope)
	if err != nil {
		return featureList, err
	}
	for key := range scopeKeys {
		featureList = append(featureList, key)
	}
	return featureList, err
}

// GetScopeFeaturesFilterByValue returns all features on the given scope with a certain value.
func (fdb FlipadelphiaBoltDB) GetScopeFeaturesFilterByValue(scope []byte, value []byte) (Serializable, error) {
	var featureList FlipadelphiaScopeFeatures

	scopeKeys, err := fdb.getScopeKeyValuesWithCertainValue(scope, value)
	if err != nil {
		return featureList, err
	}
	for key := range scopeKeys {
		featureList = append(featureList, key)
	}
	return featureList, err
}

// GetScopes returns all scopes.
func (fdb FlipadelphiaBoltDB) GetScopes() (Serializable, error) {
	scopes, err := fdb.getAllScopes()
	return scopes, err
}

// GetScopesWithPrefix returns all scopes with a certain prefix.
func (fdb FlipadelphiaBoltDB) GetScopesWithPrefix(prefix []byte) (Serializable, error) {
	scopes, err := fdb.getAllScopesWithPrefix(prefix)
	return scopes, err
}

// GetScopesWithFeature returns all scopes that have a certain feature set.
func (fdb FlipadelphiaBoltDB) GetScopesWithFeature(feature []byte) (Serializable, error) {
	scopes, err := fdb.getAllScopesWithFeature(feature)
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) GetScopesPaginated(offset, count int) (Serializable, error) {
	scopes, err := fdb.getScopesPaginated(offset, count)
	return scopes, err
}

// GetFeatures returns a list of all features set on all scopes.
func (fdb FlipadelphiaBoltDB) GetFeatures() (Serializable, error) {
	features, err := fdb.getAllFeatures()
	return features, err
}

// GetScopeFeaturesFull returns a list of FlipadelphiaFeature objects for the given scope.
func (fdb FlipadelphiaBoltDB) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
	var features FlipadelphiaFeatures

	keyVals, err := fdb.getScopeKeyValues(scope)
	if err != nil {
		return FlipadelphiaFeatures{}, err
	}
	for key, val := range keyVals {
		features = append(features, NewFlipadelphiaFeature([]byte(key), []byte(val)))
	}
	return features, nil
}
