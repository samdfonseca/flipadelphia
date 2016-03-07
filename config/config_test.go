package config

import (
	"fmt"
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
	actual := getFullFilePath(target)
	checkResult(actual, target, t)
}

func TestGetFullFilePathStoredFilePath(t *testing.T) {
	homeDir := os.Getenv("HOME")
	target := fmt.Sprintf("%s/config.json", homeDir)
	actual := getFullFilePath(target)
	checkResult(actual, target, t)
}
