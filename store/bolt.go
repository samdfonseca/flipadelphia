package store

import (
	"bytes"
	"fmt"

	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/utils"
)

// FlipadelphiaDB holds a pointer to the boltdb instance and the name of the main bucket.
type FlipadelphiaDB struct {
	db         *bolt.DB
	bucketName string
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

// NewFlipadelphiaDB creates a new instance of FlipadelphiaDB. The "features" bucket is created
// if it does not yet exist.
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
	return FlipadelphiaDB{db: db, bucketName: "features"}
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

// MergeScopeKey joins two []byte around the ":" character.
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

// SplitScopeKey splits a []byte on the first ":" character.
func SplitScopeKey(scopeKey []byte) ([]byte, []byte, error) {
	if !bytes.Contains(scopeKey, []byte(":")) {
		err := fmt.Errorf(`ScopeKey missing ":" character`)
		return []byte{}, []byte{}, err
	}
	splits := bytes.SplitN(scopeKey, []byte(":"), 2)
	return splits[0], splits[1], nil
}

func (fdb FlipadelphiaDB) getScopeKeyValues(scope []byte) (map[string][]byte, error) {
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

func (fdb FlipadelphiaDB) getAllScopes() (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList
	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		var previousScope []byte
		bucket.ForEach(func(key, val []byte) error {
			scope, _, err := SplitScopeKey(key)
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

func (fdb FlipadelphiaDB) getAllScopesWithPrefix(prefix []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList
	err := fdb.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("features")).Cursor()
		var previousScope []byte
		for key, _ := cursor.Seek(prefix); bytes.HasPrefix(key, prefix); key, _ = cursor.Next() {
			scope, _, err := SplitScopeKey(key)
			if err == nil && !bytes.Equal(scope, previousScope) {
				scopes = append(scopes, fmt.Sprintf("%s", scope))
				previousScope = scope
			}
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaDB) getAllScopesWithFeature(feature []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList
	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		bucket.ForEach(func(scopeKey, val []byte) error {
			scope, key, err := SplitScopeKey(scopeKey)
			if err == nil && bytes.Equal(feature, key) {
				scopes = append(scopes, fmt.Sprintf("%s", scope))
			}
			return nil
		})
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaDB) getAllFeatures() (FlipadelphiaScopeFeatures, error) {
	var features FlipadelphiaScopeFeatures
	err := fdb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("features"))
		var previousFeature []byte
		bucket.ForEach(func(key, val []byte) error {
			_, feature, err := SplitScopeKey(key)
			if err == nil && !bytes.Equal(feature, previousFeature) {
				features = append(features, fmt.Sprintf("%s", feature))
				previousFeature = feature
			}
			return nil
		})
		return nil
	})
	return features, err
}

// Set stores the feature in the database and returns an instance of FlipadelphiaFeature.
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

// Get retrieves the feature from the database and returns an instance of FlipadelphiaFeature.
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

// GetScopeFeatures returns all features set on the given scope.
func (fdb FlipadelphiaDB) GetScopeFeatures(scope []byte) (Serializable, error) {
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
func (fdb FlipadelphiaDB) GetScopeFeaturesFilterByValue(scope []byte, value []byte) (Serializable, error) {
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
func (fdb FlipadelphiaDB) GetScopes() (Serializable, error) {
	scopes, err := fdb.getAllScopes()
	return scopes, err
}

// GetScopesWithPrefix returns all scopes with a certain prefix.
func (fdb FlipadelphiaDB) GetScopesWithPrefix(prefix []byte) (Serializable, error) {
	scopes, err := fdb.getAllScopesWithPrefix(prefix)
	return scopes, err
}

// GetScopesWithFeature returns all scopes that have a certain feature set.
func (fdb FlipadelphiaDB) GetScopesWithFeature(feature []byte) (Serializable, error) {
	scopes, err := fdb.getAllScopesWithFeature(feature)
	return scopes, err
}

// GetFeatures returns a list of all features set on all scopes.
func (fdb FlipadelphiaDB) GetFeatures() (Serializable, error) {
	features, err := fdb.getAllFeatures()
	return features, err
}

// GetScopeFeaturesFull returns a list of FlipadelphiaFeature objects for the given scope.
func (fdb FlipadelphiaDB) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
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
