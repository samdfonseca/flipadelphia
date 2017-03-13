package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/samdfonseca/flipadelphia/utils"
)

type FlipadelphiaConfig struct {
	EnvironmentName      string
	PersistenceStoreType string `json:"persistence_store_type"`
	DBFile               string `json:"db_file"`
	RedisHost            string `json:"redis_host"`
	RedisPassword        string `json:"redis_password"`
	RedisDB              int    `json:"redis_db"`
	LogFile              string `json:"log_file"`
	ListenOnPort         int    `json:"port"`
}

var Config FlipadelphiaConfig

func getStoredFilePath(fileName string) string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		err := fmt.Errorf("")
		utils.FailOnError(err, "$HOME not set", false)
	}
	return fmt.Sprintf("%s/.flipadelphia/%s", homeDir, fileName)
}

func getFullFilePath(filePath string) string {
	if filePath == "" {
		return ""
	}
	if path.IsAbs(filePath) {
		return filePath
	}
	if strings.HasPrefix(filePath, "./") {
		cwd, _ := os.Getwd()
		fullFilePath := fmt.Sprintf("%s/%s", path.Clean(cwd), path.Clean(filePath))
		return fullFilePath
	}
	return getStoredFilePath(filePath)
}

func readConfigFile(configFilePath string) []byte {
	configData, err := ioutil.ReadFile(configFilePath)
	utils.FailOnError(err, "Unable to read config file", true)
	return configData
}

func parseConfigFile(rawConfigData []byte) (parsedConfig map[string]FlipadelphiaConfig) {
	err := json.Unmarshal(rawConfigData, &parsedConfig)
	utils.FailOnError(err, "Unable to parse config file", true)
	return parsedConfig
}

// Gets the config for a named env from a flipadelphia config file.
// Config file path can be relative or absolute. If the config file path
// is absolute, getRuntimeEnv looks for a file at the exact path. If the
// config file path starts with "./", getRuntimeEnv looks for a file
// relative to $PWD. Otherwise, getRuntimeEnv looks for a file relative
// to $HOME/.flipadelphia.
//
// Ex:
//   /etc/flipadelphia/config.json -> /etc/flipadelphia/config.json
//   ./flipadelphia_config.json -> $PWD/flipadelphia_config.json
//   config.json -> $HOME/.flipadelphia/config.json
func getRuntimeEnv(configFilePath string, envName string) FlipadelphiaConfig {
	fullConfigFilePath := getFullFilePath(configFilePath)
	configData := parseConfigFile(readConfigFile(fullConfigFilePath))
	runtimeEnv, envExists := configData[envName]
	if !envExists {
		utils.FailOnError(fmt.Errorf(""),
			fmt.Sprintf("Runtime environment %q not found in %q", envName, configFilePath), false)
	}
	runtimeEnv.EnvironmentName = envName
	runtimeEnv.DBFile = getFullFilePath(runtimeEnv.DBFile)
	runtimeEnv.LogFile = getFullFilePath(runtimeEnv.LogFile)
	return runtimeEnv
}

func NewFlipadelphiaConfig(configFilePath, envName string) FlipadelphiaConfig {
	return getRuntimeEnv(configFilePath, envName)
}
