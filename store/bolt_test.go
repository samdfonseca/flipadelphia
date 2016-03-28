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

func assertEqual(actual, target string, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func assertNil(actual interface{}, t *testing.T) {
	if actual != nil {
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual is not nil")
	}
}

func assertErrorEqual(actual, target error, t *testing.T) {
	if fmt.Sprintf("%s", actual) != fmt.Sprintf("%s", target) {
		t.Logf("Target: %s", nil)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual error did not match target error")
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
	assertEqual(string(feature.Serialize()), target, t)
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
	assertEqual(string(feature.Serialize()), target, t)
}

func TestGetScopeFeaturesSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopeFeatures: func(scope []byte) (Serializable, error) {
			return FlipadelphiaScopeFeatures{"feature1", "feature2", "feature3"}, nil
		},
	}
	features, _ := fdb.GetScopeFeatures([]byte("scope1"))
	target := `["feature1","feature2","feature3"]`
	assertEqual(string(features.Serialize()), target, t)
}

func TestGetEmptyScopeFeaturesSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopeFeatures: func(scope []byte) (Serializable, error) {
			return FlipadelphiaScopeFeatures{}, nil
		},
	}
	features, _ := fdb.GetScopeFeatures([]byte("scope1"))
	target := `[]`
	assertEqual(string(features.Serialize()), target, t)
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
	assertEqual(string(actual.Serialize()), target, t)
}

func TestMergeScopeKeyBothValid(t *testing.T) {
	actual, err := MergeScopeKey([]byte("user-1"), []byte("feature1"))
	target := `user-1:feature1`
	if err != nil {
		t.Errorf("Error merging scope and key: %s", err)
	}
	assertEqual(string(actual), target, t)
}

func TestMergeScopeKeyInvalidScope(t *testing.T) {
	_, err := MergeScopeKey([]byte("user:1"), []byte("feature1"))
	if err == nil {
		t.Errorf("Invalid scope did not cause error: %s", err)
	}
	target := fmt.Errorf("Invalid scope: Can not contain ':' character")
	assertEqual(err.Error(), target.Error(), t)
}

func TestMergeScopeKeyInvalidKey(t *testing.T) {
	_, err := MergeScopeKey([]byte("user-1"), []byte("feature,1"))
	if err == nil {
		t.Errorf("Invalid key did not cause error: %s", err)
	}
	target := fmt.Errorf("Invalid key character '%s': Valid characters are '%s'", ",", validFeatureKeyCharacters)
	assertEqual(err.Error(), target.Error(), t)
}

func TestSplitScopeKeyValidScopeKey(t *testing.T) {
	actualScope, actualKey, err := SplitScopeKey([]byte("user-1:feature1"))
	assertNil(err, t)
	assertEqual(string(actualScope), "user-1", t)
	assertEqual(string(actualKey), "feature1", t)
}

func TestSplitScopeKeyInvalidMissingColon(t *testing.T) {
	_, _, err := SplitScopeKey([]byte("user-1 feature1"))
	target := fmt.Errorf(`ScopeKey missing ":" character`)
	assertErrorEqual(err, target, t)
}

func TestGetScopesSerializes(t *testing.T) {
	fdb := MockPersistenceStore{
		OnGetScopes: func() (Serializable, error) {
			return FlipadelphiaScopeList{"user-1", "user-2"}, nil
		},
	}
	actual, _ := fdb.GetScopes()
	target := `["user-1","user-2"]`
	assertEqual(string(actual.Serialize()), target, t)
}
