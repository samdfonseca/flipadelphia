package config

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/samdfonseca/flipadelphia/utils"
	"io/ioutil"
	"os"
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

func getRuntimeEnv(confileFilePath string, envName string) FlipadelphiaConfig {
	configData := parseConfigFile(readConfigFile(confileFilePath))
	runtimeEnv := configData[envName]
	runtimeEnv.EnvironmentName = envName
	if string(runtimeEnv.DBFile[0]) != "/" {
		runtimeEnv.DBFile = GetStoredFilePath(runtimeEnv.DBFile)
	}
	return runtimeEnv
}

func NewFlipadelphiaConfig(c *cli.Context) FlipadelphiaConfig {
	return getRuntimeEnv(c.String("config"), c.String("env"))
}
