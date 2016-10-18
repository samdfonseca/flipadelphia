package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func checkResult(actual, target string, t *testing.T) {
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
	homeDir := os.Getenv("HOME")
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
	targetAuthUrl := "http://localhost:3005/session"
	targetAuthHeader := "X-SESSION-TOKEN"
	targetAuthMethod := "GET"
	targetAuthSuccessStatusCode := "200"
	targetPersistenceStoreType := "bolt"
	targetDBFile := getFullFilePath("test.db")
	targetRedisHost := "localhost"
	targetRedisPassword := ""
	targetPort := "3006"
	content := []byte(fmt.Sprintf(`{"test": {
    "auth_url": %q,
    "auth_header": %q,
    "auth_method": %q,
    "auth_success_status_code": %q,
	"persistence_store_type": %q,
    "db_file": %q,
	"redis_host": %q,
	"redis_password": %q,
    "port": %q}}`,
		targetAuthUrl,
		targetAuthHeader,
		targetAuthMethod,
		targetAuthSuccessStatusCode,
		targetPersistenceStoreType,
		targetDBFile,
		targetPort))
	parsedContent := parseConfigFile(content)
	configData := parsedContent["test"]
	checkResult(configData.AuthRequestURL, targetAuthUrl, t)
	checkResult(configData.AuthRequestHeader, targetAuthHeader, t)
	checkResult(configData.AuthRequestMethod, targetAuthMethod, t)
	checkResult(configData.AuthRequestSuccessStatusCode, targetAuthSuccessStatusCode, t)
	checkResult(configData.PersistenceStoreType, targetPersistenceStoreType, t)
	checkResult(configData.DBFile, targetDBFile, t)
	checkResult(configData.RedisHost, targetRedisHost, t)
	checkResult(configData.RedisPassword, targetRedisPassword, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}

func TestGetRuntimeEnv(t *testing.T) {
	targetAuthUrl := "http://localhost:3005/session"
	targetAuthHeader := "X-SESSION-TOKEN"
	targetAuthMethod := "GET"
	targetAuthSuccessStatusCode := "200"
	targetPersistenceStoreType := "bolt"
	targetDBFile := getFullFilePath("test.db")
	targetPort := "3006"
	content := []byte(fmt.Sprintf(`{"test": {
    "auth_url": %q,
    "auth_header": %q,
    "auth_method": %q,
    "auth_success_status_code": %q,
	"persistence_store_type": %q,
    "db_file": %q,
    "port": %q}}`,
		targetAuthUrl,
		targetAuthHeader,
		targetAuthMethod,
		targetAuthSuccessStatusCode,
		targetPersistenceStoreType,
		targetDBFile,
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
	checkResult(configData.AuthRequestURL, targetAuthUrl, t)
	checkResult(configData.AuthRequestHeader, targetAuthHeader, t)
	checkResult(configData.AuthRequestMethod, targetAuthMethod, t)
	checkResult(configData.AuthRequestSuccessStatusCode, targetAuthSuccessStatusCode, t)
	checkResult(configData.DBFile, targetDBFile, t)
	checkResult(configData.ListenOnPort, targetPort, t)
}
