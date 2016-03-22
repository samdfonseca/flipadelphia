package store

import (
	"encoding/json"
	"os/exec"
	"sort"
	"testing"

	"fmt"
	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/config"
)

var (
	TestConfig config.FlipadelphiaConfig
	TestDB     FlipadelphiaDB
)

func init() {
	TestConfig = config.NewFlipadelphiaConfig("config.json", "test")
	_ = exec.Command("touch", TestConfig.DBFile).Run()
	_ = exec.Command("rm", TestConfig.DBFile).Run()
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

func TestSetFeatureSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGet: func(scope, key []byte) (Serializable, error) {
			return FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "on",
				Data:  "true",
			}, nil
		},
	}
	feature, _ := fdb.Get([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"on","data":"true"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestUnsetFeatureSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGet: func(scope, key []byte) (Serializable, error) {
			return FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "",
				Data:  "false",
			}, nil
		},
	}
	feature, _ := fdb.Get([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"","data":"false"}`
	checkResult(string(feature.Serialize()), target, t)
}

func TestGetScopeFeaturesSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopeFeatures: func(scope []byte) (Serializable, error) {
			return FlipadelphiaScopeFeatures{"feature1", "feature2", "feature3"}, nil
		},
	}
	features, _ := fdb.GetScopeFeatures([]byte("scope1"))
	target := `["feature1","feature2","feature3"]`
	checkResult(string(features.Serialize()), target, t)
}

func TestGetEmptyScopeFeaturesSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopeFeatures: func(scope []byte) (Serializable, error) {
			return FlipadelphiaScopeFeatures{}, nil
		},
	}
	features, _ := fdb.GetScopeFeatures([]byte("scope1"))
	target := `[]`
	checkResult(string(features.Serialize()), target, t)
}

func TestGetScopeFeaturesWithCertainValueSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopeFeaturesFilterByValue: func(scope, value []byte) (Serializable, error) {
			return FlipadelphiaScopeFeatures{"feature1", "feature2", "feature4"}, nil
		},
	}
	features, _ := fdb.GetScopeFeaturesFilterByValue([]byte("scope1"), []byte("on"))
	actual := sortFeatures(features)
	target := `["feature1","feature2","feature4"]`
	checkResult(string(actual.Serialize()), target, t)
}

func TestMergeScopeKeyBothValid(t *testing.T) {
	actual, err := MergeScopeKey([]byte("user-1"), []byte("feature1"))
	target := `user-1:feature1`
	if err != nil {
		t.Errorf("Error merging scope and key: %s", err)
	}
	checkResult(string(actual), target, t)
}

func TestMergeScopeKeyInvalidScope(t *testing.T) {
	_, err := MergeScopeKey([]byte("user:1"), []byte("feature1"))
	if err == nil {
		t.Errorf("Invalid scope did not cause error: %s", err)
	}
	target := fmt.Errorf("Invalid scope: Can not contain ':' character")
	checkResult(err.Error(), target.Error(), t)
}

func TestMergeScopeKeyInvalidKey(t *testing.T) {
	_, err := MergeScopeKey([]byte("user-1"), []byte("feature,1"))
	if err == nil {
		t.Errorf("Invalid key did not cause error: %s", err)
	}
	target := fmt.Errorf("Invalid key character '%s': Valid characters are '%s'", ",", validFeatureKeyCharacters)
	t.Logf("%s", target)
	t.Logf("%s", err)
	checkResult(err.Error(), target.Error(), t)
}
