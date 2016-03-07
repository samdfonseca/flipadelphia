package store

import (
	"encoding/json"
	"sort"
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
	TestDB = NewFlipadelphiaDB(testDB)
}

func sortFeatures(features Serializable) FlipadelphiaScopeFeatures {
	var sortedFeatures []string
	json.Unmarshal(features.Serialize(), &sortedFeatures)
	sort.Strings(sortedFeatures)
	return sortedFeatures
}

func checkResult(actual, target string, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func TestCreateBucket(t *testing.T) {
	testDB, _ := bolt.Open(TestConfig.DBFile, 0600, nil)
	defer testDB.Close()
	// make sure bucket does not already exist
	testDB.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("test"))
		return nil
	})
	err := createBucket(testDB, []byte("test"))
	if err != nil {
		t.Errorf("Error when creating bucket. createBucket returned error != nil")
	}
}

func TestSetFeatureSerializes(t *testing.T) {
	InitTestDB()
	defer TestDB.db.Close()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	feature, _ := TestDB.Get([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"on","data":"true"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestUnsetFeatureSerializes(t *testing.T) {
	InitTestDB()
	defer TestDB.db.Close()
	feature, _ := TestDB.Get([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"","data":"false"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestGetScopeFeaturesSerializes(t *testing.T) {
	InitTestDB()
	defer TestDB.db.Close()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature2"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature3"), []byte("on"))
	features, _ := TestDB.GetScopeFeatures([]byte("scope1"))
	actual := sortFeatures(features)
	target := `["feature1","feature2","feature3"]`
	checkResult(string(actual.Serialize()), target, t)
}

func TestGetEmptyScopeFeaturesSerializes(t *testing.T) {
	InitTestDB()
	defer TestDB.db.Close()
	features, _ := TestDB.GetScopeFeatures([]byte("scope1"))
	target := `[]`
	checkResult(string(features.Serialize()), target, t)
}

func TestGetScopeFeaturesWithCertainValueSerializes(t *testing.T) {
	InitTestDB()
	defer TestDB.db.Close()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature2"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature3"), []byte("off"))
	TestDB.Set([]byte("scope1"), []byte("feature4"), []byte("on"))
	TestDB.Set([]byte("scope1"), []byte("feature5"), []byte("0"))
	TestDB.Set([]byte("scope1"), []byte("feature6"), []byte("ON"))
	TestDB.Set([]byte("scope1"), []byte("feature6"), []byte(""))
	TestDB.Set([]byte("scope2"), []byte("feature6"), []byte("on"))
	features, _ := TestDB.GetScopeFeaturesFilterByValue([]byte("scope1"), []byte("on"))
	actual := sortFeatures(features)
	target := `["feature1","feature2","feature4"]`
	checkResult(string(actual.Serialize()), target, t)
}
