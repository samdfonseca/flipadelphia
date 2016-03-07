package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/samdfonseca/flipadelphia/config"
	"github.com/samdfonseca/flipadelphia/server"
	"github.com/samdfonseca/flipadelphia/store"
	"github.com/samdfonseca/flipadelphia/utils"
)

var flipadelphiaVersion = "dev-build"

func main() {
	app := cli.NewApp()
	app.Name = "flipadelphia"
	app.Usage = "flipadelphia flips your features"
	app.Version = flipadelphiaVersion
	app.Commands = []cli.Command{
		{
			Name:    "sanitycheck",
			Aliases: []string{"c"},
			Usage:   "Run a quick sanity check (DEV PURPOSES ONLY)",
			Action: func(c *cli.Context) {
				config.Config = config.NewFlipadelphiaConfig("config.json", "test")
				exec.Command("rm", "-f", config.Config.DBFile).Run()
				db, err := bolt.Open(config.Config.DBFile, 0600, nil)
				utils.FailOnError(err, "Unable to open db file", true)
				defer db.Close()
				flipDB := store.NewFlipadelphiaDB(*db)
				feature1, err := flipDB.Set([]byte("venue-1"), []byte("feature1"), []byte("off"))
				utils.FailOnError(err, "Unable to set feature", true)
				feature1, err = flipDB.Get([]byte("venue-1"), []byte("feature1"))
				utils.FailOnError(err, "Unable to get feature", true)
				utils.Output(string(feature1.Serialize()))
				exec.Command("rm", "-f", config.Config.DBFile).Run()
			},
		},
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "Start the Flipadelphia server",
			Flags: []cli.Flag{
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
			},
			Action: func(c *cli.Context) {
				config.Config = config.NewFlipadelphiaConfig(c.String("config"), c.String("env"))
				db, err := bolt.Open(config.Config.DBFile, 0600, nil)
				utils.FailOnError(err, "Unable to open db file", true)
				defer db.Close()
				flipDB := store.NewFlipadelphiaDB(*db)
				utils.Output(fmt.Sprintf("Listening on port %s", config.Config.ListenOnPort))
				auth := server.NewAuthSettings(config.Config.AuthRequestURL,
					config.Config.AuthRequestMethod,
					config.Config.AuthRequestHeader,
					config.Config.AuthRequestSuccessStatusCode)
				err = http.ListenAndServe(fmt.Sprintf(":%s", config.Config.ListenOnPort), server.App(flipDB, auth))
				utils.FailOnError(err, "Something went wrong", true)
			},
		},
	}

	app.Run(os.Args)
}
