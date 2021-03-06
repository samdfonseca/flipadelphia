package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/boltdb/bolt"
)

func RunTestWithTempDB(t *testing.T, test func(db FlipadelphiaBoltDB, t *testing.T)) {
	dir, err := ioutil.TempDir("", "flipadelphia_test")
	defer os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
	tmpPath := path.Join(dir, "test.db")
	tmpBolt, _ := bolt.Open(tmpPath, 0600, nil)
	defer tmpBolt.Close()
	tmpBolt.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("features"))
		return nil
	})
	testDB := NewFlipadelphiaBoltDB(tmpBolt)
	test(testDB, t)
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
		t.Logf("Target: %s", target)
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

func TestGetScopesPaginatedWithOffset(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testScopes := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testFeatures := []string{"feature1", "feature2", "feature3"}
		for _, scope := range testScopes {
			for _, feature := range testFeatures {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		paginatedScopes, _ := db.getScopesPaginated(5, 10)
		if len(paginatedScopes) != 10 {
			t.Logf("Target Length: 10")
			t.Logf("Actual Length: %s", len(paginatedScopes))
			t.Errorf("Length of paginatedScopes did not equal 10")
		}
		for i := range paginatedScopes {
			assertEqual(paginatedScopes[i], testScopes[i+5], t)
		}
	})
}

func TestGetScopesPaginatedWithoutOffset(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testScopes := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testFeatures := []string{"feature1", "feature2", "feature3"}
		for _, scope := range testScopes {
			for _, feature := range testFeatures {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		paginatedScopes, _ := db.getScopesPaginated(0, 10)
		if len(paginatedScopes) != 10 {
			t.Logf("Target Length: 10")
			t.Logf("Actual Length: %s", len(paginatedScopes))
			t.Errorf("Length of paginatedScopes did not equal 10")
		}
		for i := range paginatedScopes {
			assertEqual(paginatedScopes[i], testScopes[i], t)
		}
	})
}

func TestGetScopesPaginatedWithoutOffsetCountGreaterThanAvailable(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testScopes := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testFeatures := []string{"feature1", "feature2", "feature3"}
		for _, scope := range testScopes {
			for _, feature := range testFeatures {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		paginatedScopes, _ := db.getScopesPaginated(0, 100)
		if len(paginatedScopes) != 20 {
			t.Logf("Target Length: 20")
			t.Logf("Actual Length: %s", len(paginatedScopes))
			t.Errorf("Length of paginatedScopes did not equal 20")
		}
		for i := range paginatedScopes {
			assertEqual(paginatedScopes[i], testScopes[i], t)
		}
	})
}

func TestGetScopesPaginatedWithOffsetCountGreaterThanAvailable(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testScopes := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testFeatures := []string{"feature1", "feature2", "feature3"}
		for _, scope := range testScopes {
			for _, feature := range testFeatures {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		paginatedScopes, _ := db.getScopesPaginated(10, 100)
		if len(paginatedScopes) != 10 {
			t.Logf("Target Length: 10")
			t.Logf("Actual Length: %s", len(paginatedScopes))
			t.Errorf("Length of paginatedScopes did not equal 10")
		}
		for i := range paginatedScopes {
			assertEqual(paginatedScopes[i], testScopes[i+10], t)
		}
	})
}

func TestGetFeaturesPaginatedWithOffset(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testFeatures := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testScopes := []string{"scope1", "scope2", "scope3"}
		for _, feature := range testFeatures {
			for _, scope := range testScopes {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		paginatedFeatures, _ := db.getFeaturesPaginated(5, 10)
		if len(paginatedFeatures) != 10 {
			t.Logf("Target Length: 10")
			t.Logf("Actual Length: %s", len(paginatedFeatures))
			t.Errorf("Length of paginatedFeatures did not equal 10")
		}
		for i := range paginatedFeatures {
			assertEqual(paginatedFeatures[i], testFeatures[i+5], t)
		}
	})
}

func TestGetAllFeatures(t *testing.T) {
	RunTestWithTempDB(t, func(db FlipadelphiaBoltDB, t *testing.T) {
		testFeatures := []string{
			"a",
			"ab",
			"amet",
			"at",
			"cupiditate",
			"ea",
			"eum",
			"fugiat",
			"magnam",
			"maxime",
			"mollitia",
			"nihil",
			"quaerat",
			"quas",
			"quidem",
			"reiciendis",
			"repudiandae",
			"velit",
			"veritatis",
			"voluptas",
		}
		testScopes := []string{"scope1", "scope2", "scope3"}
		for _, feature := range testFeatures {
			for _, scope := range testScopes {
				db.Set([]byte(scope), []byte(feature), []byte("on"))
			}
		}
		allFeatures, _ := db.getAllFeatures()
		if len(allFeatures) != len(testFeatures) {
			t.Logf("Target Length: %s", len(testFeatures))
			t.Logf("Actual Length: %s", len(allFeatures))
			t.Errorf("Length of allFeatures did not equal %s", len(testFeatures))
		}
		for i := range allFeatures {
			assertEqual(allFeatures[i], testFeatures[i], t)
		}
	})
}
