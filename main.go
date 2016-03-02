package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/samdfonseca/flipadelphia/config"
	"github.com/samdfonseca/flipadelphia/server"
	"github.com/samdfonseca/flipadelphia/store"
	"github.com/samdfonseca/flipadelphia/utils"
	"net/http"
	"os"
)

var flipadelphiaVersion = "dev-build"

func main() {
	app := cli.NewApp()
	app.Name = "flipadelphia"
	app.Usage = "flipadelphia flips your features"
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
			Value:  config.GetStoredFilePath("config.json"),
			EnvVar: "FLIPADELPHIA_CONFIG",
		},
	}
	app.Action = func(c *cli.Context) {
		c.Args()
		config.Config = config.NewFlipadelphiaConfig(c)
		db, err := bolt.Open(config.Config.DBFile, 0600, nil)
		utils.FailOnError(err, "Unable to open db file", true)
		defer db.Close()
		flipDB := store.NewFlipadelphiaDB(*db)
		feature1, err := flipDB.Set([]byte("venue-1"), []byte("feature1"), []byte("off"))
		utils.FailOnError(err, "Unable to set feature", true)
		feature1, err = flipDB.Get([]byte("venue-1"), []byte("feature1"))
		utils.FailOnError(err, "Unable to get feature", true)
		utils.Output(string(feature1.Serialize()))
		http.ListenAndServe(fmt.Sprintf(":%s", config.Config.ListenOnPort), server.App(flipDB))
	}

	app.Run(os.Args)
}
