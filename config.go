package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type RuntimeEnvironmentConfig struct {
	EnvironmentName string
	AuthServerURL   string	`json:"auth_server"`
	RedisHost       string	`json:"redis_host"`
	RedisPort       string	`json:"redis_port"`
	RedisProtocol   string	`json:"redis_protocol"`
	DBFile          string	`json:"db_file"`
}

type FlipadelphiaConfig struct {
	RuntimeEnvironment RuntimeEnvironmentConfig
	ListenOnPort       string
}

var FC FlipadelphiaConfig

func getStoredFilePath(fileName string) string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		exitErr(fmt.Errorf("$HOME not set"))
	}
	return fmt.Sprintf("%s/.flipadelphia/%s", homeDir, fileName)
}

func readConfigFile() []byte {
	configData, err := ioutil.ReadFile(getStoredFilePath("config.json"))
	failOnError(err, "Unable to read config file", true)
	return configData
}

func parseConfigFile(rawConfigData []byte) (parsedConfig map[string]RuntimeEnvironmentConfig) {
	//parsedConfig := make(map[string]map[string]string)
	err := json.Unmarshal(rawConfigData, &parsedConfig)
	failOnError(err, "Unable to parse config file", true)
	return parsedConfig
}

func NewFlipadelphiaConfig(runtimeEnvName string, listenOnPort string) FlipadelphiaConfig {
	configData := parseConfigFile(readConfigFile())
	runtimeEnv := configData[runtimeEnvName]
	flipConfig := FlipadelphiaConfig{
		RuntimeEnvironment: runtimeEnv,
		ListenOnPort:       listenOnPort,
	}
	if string(flipConfig.RuntimeEnvironment.DBFile[0]) != "/" {
		flipConfig.RuntimeEnvironment.DBFile = getStoredFilePath(flipConfig.RuntimeEnvironment.DBFile)
	}
	return flipConfig
}

func init() {
	var (
		listenOnPort           string
		runtimeEnvironmentName string
	)
	flag.StringVar(&listenOnPort, "port", "3006", "The port Flipadelphia listens on")
	flag.StringVar(&runtimeEnvironmentName, "env", "development", "An environment from the config.json file to use")
	//flag.StringVar(&logFilePath, "logfile", getStoredFilePath(defaultLogFileName), "Full path to the logfile (doesn't have to exist yet)"
	//flag.BoolVar(&logVerbose, "verbose", false, "Print info level logs to stdout")

	flag.Parse()

	FC = NewFlipadelphiaConfig(runtimeEnvironmentName, listenOnPort)
}
