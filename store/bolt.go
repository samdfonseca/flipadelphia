package store

import (
	"bytes"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
	"github.com/samdfonseca/flipadelphia/utils"
)

// FlipadelphiaBoltDB holds a pointer to the boltdb instance and the name of the main bucket.
type FlipadelphiaBoltDB struct {
	db *bolt.DB
}

type BucketCreator interface {
	CreateBucket([]byte) (*bolt.Bucket, error)
	CreateBucketIfNotExists([]byte) (*bolt.Bucket, error)
}

func createBuckets(bc BucketCreator, bucketNames ...[]byte) error {
	for _, bktname := range bucketNames {
		if _, err := bc.CreateBucketIfNotExists(bktname); err != nil {
			msg := "Failed to create bucket: %q"
			utils.LogOnError(err, fmt.Sprintf(msg, bktname), true)
			return err
		}
		msg := fmt.Sprintf("CREATED BUCKET - Name: %q", bktname)
		utils.LogOnSuccess(nil, msg)
	}
	return nil
}

// NewFlipadelphiaBoltDB creates a new instance of FlipadelphiaBoltDB. The "features" bucket is created
// if it does not yet exist.
func NewFlipadelphiaBoltDB(db *bolt.DB) FlipadelphiaBoltDB {
	requiredBuckets := [][]byte{
		[]byte("features"),
		[]byte("scopes"),
		[]byte("values"),
	}
	db.Update(func(tx *bolt.Tx) error {
		err := createBuckets(tx, requiredBuckets...)
		return err
	})
	return FlipadelphiaBoltDB{db: db}
}

func (fdb FlipadelphiaBoltDB) Close() error {
	return fdb.db.Close()
}

func (fdb FlipadelphiaBoltDB) getScopeFeatureValues(scope []byte) (map[string][]byte, error) {
	var values = make(map[string][]byte)
	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		valuesBkt := tx.Bucket([]byte("values"))
		if valuesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "values"`)
		}
		scopeBkt := scopesBkt.Bucket(scope)
		if scopeBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes/%q"`, scope)
		}
		if err := scopeBkt.ForEach(func(k, v []byte) error {
			value := valuesBkt.Get(v)
			values[utils.Btos(k)] = value
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (fdb FlipadelphiaBoltDB) getScopeKeyValuesWithCertainValue(scope []byte, targetValue []byte) (map[string][]byte, error) {
	keys, err := fdb.getScopeFeatureValues(scope)
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
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		if err := scopesBkt.ForEach(func(scope, valueUUID []byte) error {
			scopes = append(scopes, fmt.Sprintf("%s", scope))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getAllScopesWithPrefix(prefix []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList

	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		cursor := scopesBkt.Cursor()
		for scope, _ := cursor.Seek(prefix); bytes.HasPrefix(scope, prefix); scope, _ = cursor.Next() {
			scopes = append(scopes, fmt.Sprintf("%s", scope))
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getScopesPaginated(offset, count int) (StringSlice, error) {
	var scopes StringSlice

	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		cursor := scopesBkt.Cursor()
		scope, _ := cursor.First()
		// Advance the cursor to the desired offset
		for counter := 0; scope != nil && offset != 0 && counter < offset; scope, _ = cursor.Next() {
			counter++
		}
		// Retrieve the next n scopes, where n=count
		// Checks for key != nil to handle overflow, i.e. a bucket with 10 items, offset=5 and count=10
		for scope != nil && len(scopes) < count {
			scopes = append(scopes, string(scope))
			scope, _ = cursor.Next()
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getFeaturesPaginated(offset, count int) (StringSlice, error) {
	var features StringSlice

	err := fdb.db.View(func(tx *bolt.Tx) error {
		featuresBkt := tx.Bucket([]byte("features"))
		if featuresBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "features"`)
		}
		cursor := featuresBkt.Cursor()
		feature, _ := cursor.First()
		// Advance the cursor to the desired offset
		for counter := 0; feature != nil && offset != 0 && counter < offset; feature, _ = cursor.Next() {
			counter++
		}
		// Retrieve the next n features, where n=count
		// Checks for key != nil to handle overflow, i.e. a bucket with 10 items, offset=5 and count=10
		for feature != nil && len(features) < count {
			features = append(features, string(feature))
			feature, _ = cursor.Next()
		}
		return nil
	})
	return features, err
}

func (fdb FlipadelphiaBoltDB) getAllScopesWithFeature(feature []byte) (FlipadelphiaScopeList, error) {
	var scopes FlipadelphiaScopeList

	err := fdb.db.View(func(tx *bolt.Tx) error {
		featuresBkt := tx.Bucket([]byte("features"))
		if featuresBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "features"`)
		}
		featureBkt := featuresBkt.Bucket(feature)
		if featureBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "features/%q"`, feature)
		}
		if err := featureBkt.ForEach(func(scope, valueUUID []byte) error {
			scopes = append(scopes, fmt.Sprintf("%s", scope))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return scopes, err
}

func (fdb FlipadelphiaBoltDB) getAllFeaturesWithScope(scope []byte) (FlipadelphiaFeatures, error) {
	var features FlipadelphiaFeatures

	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		scopeBkt := scopesBkt.Bucket(scope)
		if scopeBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes/%q"`, scope)
		}
		valuesBkt := tx.Bucket([]byte("values"))
		if valuesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "values"`)
		}
		if err := scopeBkt.ForEach(func(feature, valueUUID []byte) error {
			value := valuesBkt.Get(valueUUID)
			features = append(features, NewFlipadelphiaFeature(feature, value))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return features, err
}

func (fdb FlipadelphiaBoltDB) getAllFeatures() (FlipadelphiaScopeFeatures, error) {
	var features FlipadelphiaScopeFeatures

	err := fdb.db.View(func(tx *bolt.Tx) error {
		featuresBkt := tx.Bucket([]byte("features"))
		if err := featuresBkt.ForEach(func(feature, value []byte) error {
			features = append(features, fmt.Sprintf("%s", feature))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return features, err
}

func (fdb FlipadelphiaBoltDB) setScopeFeature(tx *bolt.Tx, scope, feature, scopeFeatUUID []byte) error {
	scopesBkt := tx.Bucket([]byte("scopes"))
	if scopesBkt == nil {
		if err := createBuckets(tx, []byte("scopes")); err != nil {
			return err
		}
		return fdb.setScopeFeature(tx, scope, feature, scopeFeatUUID)
	}
	scopeBkt := scopesBkt.Bucket(scope)
	if scopeBkt == nil {
		if err := createBuckets(scopesBkt, scope); err != nil {
			return err
		}
		return fdb.setScopeFeature(tx, scope, feature, scopeFeatUUID)
	}
	if err := scopeBkt.Put(feature, scopeFeatUUID); err != nil {
		return err
	}
	return nil
}

func (fdb FlipadelphiaBoltDB) setFeatureScope(tx *bolt.Tx, scope, feature, scopeFeatUUID []byte) error {
	featsBkt := tx.Bucket([]byte("features"))
	if featsBkt == nil {
		if err := createBuckets(tx, []byte("features")); err != nil {
			return err
		}
		return fdb.setFeatureScope(tx, scope, feature, scopeFeatUUID)
	}
	featBkt := featsBkt.Bucket(feature)
	if featBkt == nil {
		if err := createBuckets(featsBkt, feature); err != nil {
			return err
		}
		return fdb.setFeatureScope(tx, scope, feature, scopeFeatUUID)
	}
	if err := featBkt.Put(scope, scopeFeatUUID); err != nil {
		return err
	}
	return nil
}

func (fdb FlipadelphiaBoltDB) setScopeFeatureUUIDValue(tx *bolt.Tx, scopeFeatUUID, value []byte) error {
	valuesBkt := tx.Bucket([]byte("values"))
	if valuesBkt == nil {
		if err := createBuckets(tx, []byte("values")); err != nil {
			return err
		}
		return fdb.setScopeFeatureUUIDValue(tx, scopeFeatUUID, value)
	}
	if err := valuesBkt.Put(scopeFeatUUID, value); err != nil {
		return err
	}
	return nil
}

// Set stores the feature in the database and returns an instance of FlipadelphiaFeature.
func (fdb FlipadelphiaBoltDB) Set(scope []byte, feature []byte, value []byte) (Serializable, error) {
	err := fdb.db.Update(func(tx *bolt.Tx) error {
		scopeFeatUUID := uuid.NewV4().Bytes()
		if err := fdb.setScopeFeature(tx, scope, feature, scopeFeatUUID); err != nil {
			return err
		}

		if err := fdb.setFeatureScope(tx, scope, feature, scopeFeatUUID); err != nil {
			return err
		}

		if err := fdb.setScopeFeatureUUIDValue(tx, scopeFeatUUID, value); err != nil {
			return err
		}

		return nil
	})
	return NewFlipadelphiaFeature(feature, value), err
}

// Get retrieves the feature from the database and returns an instance of FlipadelphiaFeature.
func (fdb FlipadelphiaBoltDB) Get(scope []byte, feature []byte) (Serializable, error) {
	var value []byte

	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		scopeBkt := scopesBkt.Bucket(scope)
		scopeFeatUUID := scopeBkt.Get(feature)
		if scopeFeatUUID == nil {
			return fmt.Errorf("Feature %q not set for scope %q", feature, scope)
		}

		valuesBkt := tx.Bucket([]byte("values"))
		value = valuesBkt.Get(scopeFeatUUID)
		if value == nil {
			return fmt.Errorf("Feature %q not set for scope %q", feature, scope)
		}
		return nil
	})
	return NewFlipadelphiaFeature(feature, value), err
}

// GetScopeFeatures returns all features set on the given scope.
func (fdb FlipadelphiaBoltDB) GetScopeFeatures(scope []byte) (Serializable, error) {
	var featureList FlipadelphiaScopeFeatures

	scopeKeys, err := fdb.getScopeFeatureValues(scope)
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

func (fdb FlipadelphiaBoltDB) GetFeaturesPaginated(offset, count int) (Serializable, error) {
	features, err := fdb.getFeaturesPaginated(offset, count)
	return features, err
}

// GetFeatures returns a list of all features set on all scopes.
func (fdb FlipadelphiaBoltDB) GetFeatures() (Serializable, error) {
	features, err := fdb.getAllFeatures()
	return features, err
}

// GetScopeFeaturesFull returns a list of FlipadelphiaFeature objects for the given scope.
func (fdb FlipadelphiaBoltDB) GetScopeFeaturesFull(scope []byte) (Serializable, error) {
	var features FlipadelphiaFeatures

	keyVals, err := fdb.getScopeFeatureValues(scope)
	if err != nil {
		return FlipadelphiaFeatures{}, err
	}
	for key, val := range keyVals {
		features = append(features, NewFlipadelphiaFeature([]byte(key), []byte(val)))
	}
	return features, nil
}

func (fdb FlipadelphiaBoltDB) CheckScopeExists(scope []byte) bool {
	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopesBkt := tx.Bucket([]byte("scopes"))
		if scopesBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes"`)
		}
		scopeBkt := scopesBkt.Bucket(scope)
		if scopeBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "scopes/%q"`, scope)
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

func (fdb FlipadelphiaBoltDB) CheckFeatureExists(feature []byte) bool {
	err := fdb.db.View(func(tx *bolt.Tx) error {
		featuresBkt := tx.Bucket([]byte("features"))
		if featuresBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "features"`)
		}
		featureBkt := featuresBkt.Bucket(feature)
		if featureBkt == nil {
			return fmt.Errorf(`Bucket does not exist: "features/%q"`, feature)
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

func (fdb FlipadelphiaBoltDB) CheckScopeHasFeature(scope, feature []byte) bool {
	if scopeExists := fdb.CheckScopeExists(scope); !scopeExists {
		return scopeExists
	}
	if featureExists := fdb.CheckFeatureExists(feature); !featureExists {
		return featureExists
	}
	err := fdb.db.View(func(tx *bolt.Tx) error {
		scopeBkt := tx.Bucket([]byte("scopes")).Bucket(scope)
		if val := scopeBkt.Get(feature); val == nil {
			return fmt.Errorf(`Bucket key does not exist: "scopes/%q/%q"`, scope, feature)
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

func (fdb FlipadelphiaBoltDB) CheckFeatureHasScope(scope, feature []byte) bool {
	if scopeExists := fdb.CheckScopeExists(scope); !scopeExists {
		return scopeExists
	}
	if featureExists := fdb.CheckFeatureExists(feature); !featureExists {
		return featureExists
	}
	err := fdb.db.View(func(tx *bolt.Tx) error {
		featureBkt := tx.Bucket([]byte("features")).Bucket(feature)
		if val := featureBkt.Get(scope); val == nil {
			return fmt.Errorf(`Bucket key does not exist: "features/%q/%q"`, feature, scope)
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}
