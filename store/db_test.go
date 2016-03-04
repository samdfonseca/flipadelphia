package store

import (
	"github.com/boltdb/bolt"
	"github.com/samdfonseca/flipadelphia/config"
	"testing"
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

func TestSetFeatureSerializes(t *testing.T) {
	InitTestDB()
	TestDB.Set([]byte("scope1"), []byte("feature1"), []byte("on"))
	feature := TestDB.GetAndClose([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"on","data":"true"}`
	if actual := feature.Serialize(); string(actual) != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Feature did not serialize correctly")
	}
}

func TestUnsetFeatureSerializes(t *testing.T) {
	InitTestDB()
	feature := TestDB.GetAndClose([]byte("scope1"), []byte("feature1"))
	target := `{"name":"feature1","value":"","data":"false"}`
	if actual := feature.Serialize(); string(actual) != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Feature did not serialize correctly")
	}
}
