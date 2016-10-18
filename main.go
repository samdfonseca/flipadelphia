package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/samdfonseca/flipadelphia/config"
	"github.com/samdfonseca/flipadelphia/server"
	"github.com/samdfonseca/flipadelphia/store"
	"github.com/samdfonseca/flipadelphia/utils"
	"github.com/urfave/cli"
)

var flipadelphiaVersion string

func init() {
	flipadelphiaVersion = "dev-build"
	if v := os.Getenv("FLIPADELPHIA_BUILD_VERSION"); v != "" {
		flipadelphiaVersion = v
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "flipadelphia"
	app.Usage = "Start the Flipadelphia server"
	app.Version = flipadelphiaVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "env, e",
			Value:  "development",
			Usage:  "An environment from the config.json file to use",
			EnvVar: "FLIPADELPHIA_ENV",
		},
		cli.StringFlag{
			Name:   "config",
			Usage:  "Path to the config file.",
			Value:  "config.json",
			EnvVar: "FLIPADELPHIA_CONFIG",
		},
	}
	app.Action = func(c *cli.Context) {
		config.Config = config.NewFlipadelphiaConfig(c.String("config"), c.String("env"))
		flipDB := store.NewPersistenceStore(config.Config)
		defer flipDB.Close()
		utils.Output(fmt.Sprintf("Listening on port %s", config.Config.ListenOnPort))
		err := http.ListenAndServe(fmt.Sprintf(":%s", config.Config.ListenOnPort), server.App(flipDB))
		utils.FailOnError(err, "Something went wrong", true)
	}

	app.Run(os.Args)
}
