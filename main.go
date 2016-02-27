package main

import (

	"github.com/boltdb/bolt"
)

func main() {
	db, err := bolt.Open(FC.RuntimeEnvironment.DBFile, 0600, nil)
	failOnError(err, "Unable to open db file", true)
	defer db.Close()
	FDB.db = *db
	err = FDB.Set([]byte("venue-1"), []byte("feature1"), []byte("1"))
	failOnError(err, "Unable to set feature", true)
	feature1, err := FDB.Get([]byte("venue-1"), []byte("feature1"))
	failOnError(err, "Unable to get feature", true)
	output(feature1.Name + ": " + feature1.Value)
}