package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func checkResult(actual, target interface{}, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func TestGetFullFilePathAbsolutePath(t *testing.T) {
	target := `/tmp/config.json`
	actual := getFullFilePath(target)
	checkResult(actual, target, t)
}

func TestGetFullFilePathRelativePath(t *testing.T) {
	cwd, _ := os.Getwd()
	target := fmt.Sprintf("%s/config.json", cwd)
	actual := getFullFilePath("./config.json")
	checkResult(actual, target, t)
}

func TestGetFullFilePathStoredFilePath(t *testing.T) {
	homeDir, ok := os.LookupEnv("HOME")
	if !ok {
		err := os.Setenv("HOME", "/tmp")
		if err != nil {
			t.Fatal("Unable to set HOME env variable")
		}
		defer os.Setenv("HOME", "")
	}
	target := fmt.Sprintf("%s/.flipadelphia/config.json", homeDir)
	actual := getFullFilePath("config.json")
	checkResult(actual, target, t)
}

func TestReadFileContent(t *testing.T) {
	content := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	tmpfile, err := ioutil.TempFile("", "flipadelphia")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	actual := readConfigFile(tmpfile.Name())
	checkResult(string(actual), string(content), t)
}

func TestParseConfigFile(t *testing.T) {
	targetPersistenceStoreType := "bolt"
	targetDBFile := getFullFilePath("test.db")
	targetLogFile := getFullFilePath("test.log")
	targetRedisHost := "localhost:6379"
	targetRedisPassword := "password"
	targetRedisDB := 0
	targetPort := 3006
	content := []byte(fmt.Sprintf(`{"test": {
	"persistence_store_type": %q,
    "db_file": %q,
	"log_file": %q,
	"redis_host": %q,
	"redis_password": %q,
	"redis_db": %d,
    "port": %d}}`,
		targetPersistenceStoreType,
		targetDBFile,
		targetLogFile,
		targetRedisHost,
		targetRedisPassword,
		targetRedisDB,
		targetPort))
	parsedContent := parseConfigFile(content)
	configData := parsedContent["test"]
	checkResult(configData.PersistenceStoreType, targetPersistenceStoreType, t)
	checkResult(configData.DBFile, targetDBFile, t)
	checkResult(configData.LogFile, targetLogFile, t)
	checkResult(configData.RedisHost, targetRedisHost, t)
	checkResult(configData.RedisPassword, targetRedisPassword, t)
	checkResult(configData.RedisDB, targetRedisDB, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}

func TestParseConfigFileWithoutRedis(t *testing.T) {
	targetPersistenceStoreType := "bolt"
	targetDBFile := getFullFilePath("test.db")
	targetLogFile := getFullFilePath("test.log")
	targetPort := 3006
	content := []byte(fmt.Sprintf(`{"test": {
	"persistence_store_type": %q,
    "db_file": %q,
	"log_file": %q,
    "port": %d}}`,
		targetPersistenceStoreType,
		targetDBFile,
		targetLogFile,
		targetPort))
	parsedContent := parseConfigFile(content)
	configData := parsedContent["test"]
	checkResult(configData.PersistenceStoreType, targetPersistenceStoreType, t)
	checkResult(configData.DBFile, targetDBFile, t)
	checkResult(configData.LogFile, targetLogFile, t)
	checkResult(configData.RedisHost, "", t)
	checkResult(configData.RedisPassword, "", t)
	checkResult(configData.RedisDB, 0, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}

func TestParseConfigFileWithoutDBFile(t *testing.T) {
	targetPersistenceStoreType := "bolt"
	targetLogFile := getFullFilePath("test.log")
	targetRedisHost := "localhost:6379"
	targetRedisPassword := "password"
	targetRedisDB := 0
	targetPort := 3006
	content := []byte(fmt.Sprintf(`{"test": {
	"persistence_store_type": %q,
	"log_file": %q,
    "redis_host": %q,
    "redis_password": %q,
    "redis_db": %d,
    "port": %d}}`,
		targetPersistenceStoreType,
		targetLogFile,
		targetRedisHost,
		targetRedisPassword,
		targetRedisDB,
		targetPort))
	parsedContent := parseConfigFile(content)
	configData := parsedContent["test"]
	checkResult(configData.PersistenceStoreType, targetPersistenceStoreType, t)
	checkResult(configData.DBFile, "", t)
	checkResult(configData.LogFile, targetLogFile, t)
	checkResult(configData.RedisHost, targetRedisHost, t)
	checkResult(configData.RedisPassword, targetRedisPassword, t)
	checkResult(configData.RedisDB, targetRedisDB, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}

func TestGetRuntimeEnv(t *testing.T) {
	targetPersistenceStoreType := "bolt"
	targetDBFile := getFullFilePath("test.db")
	targetLogFile := getFullFilePath("test.log")
	targetRedisHost := "localhost:6379"
	targetRedisPassword := "password"
	targetRedisDB := 0
	targetPort := 3006
	content := []byte(fmt.Sprintf(`{"test": {
	"persistence_store_type": %q,
    "db_file": %q,
	"log_file": %q,
	"redis_host": %q,
	"redis_password": %q,
	"redis_db": %d,
    "port": %d}}`,
		targetPersistenceStoreType,
		targetDBFile,
		targetLogFile,
		targetRedisHost,
		targetRedisPassword,
		targetRedisDB,
		targetPort))
	tmpfile, err := ioutil.TempFile("", "flipadelphia")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	configData := getRuntimeEnv(tmpfile.Name(), "test")
	checkResult(configData.PersistenceStoreType, targetPersistenceStoreType, t)
	checkResult(configData.DBFile, targetDBFile, t)
	checkResult(configData.LogFile, targetLogFile, t)
	checkResult(configData.RedisHost, targetRedisHost, t)
	checkResult(configData.RedisPassword, targetRedisPassword, t)
	checkResult(configData.RedisDB, targetRedisDB, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}
