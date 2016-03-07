package store

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/config"
)

var (
	TestConfig config.FlipadelphiaConfig
	TestDB     FlipadelphiaDB
)

func init() {
	TestConfig = config.NewFlipadelphiaConfig("config.json", "test")
}

func InitTestDB() {
	testDB, _ := bolt.Open(TestConfig.DBFile, 0600, nil)
	testDB.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("features"))
		return nil
	})
	TestDB = NewFlipadelphiaDB(*testDB)
}

func (db FlipadelphiaDB) GetAndClose(scope, key []byte) Serializable {
	feature, _ := db.Get(scope, key)
	db.db.Close()
	return feature
}

func checkResult(actual, target string, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func TestSetFeatureSerializes(t *testing.T) {
	InitTestDB()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	feature := TestDB.GetAndClose([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"on","data":"true"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestUnsetFeatureSerializes(t *testing.T) {
	InitTestDB()
	feature := TestDB.GetAndClose([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"","data":"false"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestGetScopeFeatures(t *testing.T) {
	InitTestDB()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature2"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature3"), []byte("on"))
	features, _ := TestDB.GetScopeFeatures([]byte("scope1"))
	TestDB.db.Close()
	target := `["feature1","feature2","feature3"]`
	checkResult(string(features.Serialize()), target, t)
}

func TestGetEmptyScopeFeatures(t *testing.T) {
	InitTestDB()
	features, _ := TestDB.GetScopeFeatures([]byte("scope1"))
	TestDB.db.Close()
	target := `[]`
	checkResult(string(features.Serialize()), target, t)
}
