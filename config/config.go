package config

import (
	"encoding/json"
	"fmt"
	"github.com/samdfonseca/flipadelphia/utils"
	"io/ioutil"
	"os"
	"strings"
)

type FlipadelphiaConfig struct {
	EnvironmentName string
	AuthServerURL   string `json:"auth_server"`
	DBFile          string `json:"db_file"`
	ListenOnPort    string `json:"port"`
}

var Config FlipadelphiaConfig

func GetStoredFilePath(fileName string) string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		err := fmt.Errorf("")
		utils.FailOnError(err, "$HOME not set", false)
	}
	return fmt.Sprintf("%s/.flipadelphia/%s", homeDir, fileName)
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

func getRuntimeEnv(configFilePath string, envName string) FlipadelphiaConfig {
	if !strings.HasPrefix(configFilePath, "/") || !strings.HasPrefix(configFilePath, "./") {
		configFilePath = GetStoredFilePath(configFilePath)
	}
	configData := parseConfigFile(readConfigFile(configFilePath))
	runtimeEnv, envExists := configData[envName]
	if !envExists {
		utils.FailOnError(fmt.Errorf(""), fmt.Sprintf("Runtime environment %q not found in %q", envName, configFilePath), false)
	}
	runtimeEnv.EnvironmentName = envName
	if !strings.HasPrefix(runtimeEnv.DBFile, "/") || !strings.HasPrefix(runtimeEnv.DBFile, "./") {
		runtimeEnv.DBFile = GetStoredFilePath(runtimeEnv.DBFile)
	}
	return runtimeEnv
}

func NewFlipadelphiaConfig(configFilePath, envName string) FlipadelphiaConfig {
	return getRuntimeEnv(configFilePath, envName)
}
