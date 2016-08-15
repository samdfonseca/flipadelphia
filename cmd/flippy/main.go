package main

import (
	"fmt"
	"os"

	"github.com/samdfonseca/flipadelphia/utils"
	"github.com/urfave/cli"
)

var flippyVersion string

func init() {
	flippyVersion = "dev-build"
	if v := os.Getenv("FLIPPY_BUILD_VERSION"); v != "" {
		flippyVersion = v
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "flippy"
	app.Usage = "A CLI interface to the Flipadelphia server"
	app.Version = flippyVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "url",
			Value:  "localhost:3006",
			Usage:  "The base URL of the flipadelphia server.",
			EnvVar: "FLIPADELPHIA_URL",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "get-scopes",
			Aliases: []string{"gs"},
			Usage:   "Fetch all the existing scopes",
			Action: func(c *cli.Context) error {
				client := NewFlippyClient(c.GlobalString("url"))
				data, _ := client.GetScopes()
				scopes, _ := data.GetStringArray("data")
				utils.Output("get-scopes", "flippy")
				for _, v := range scopes {
					fmt.Println(v)
				}
				return nil
			},
		},
		{
			Name:    "get-features",
			Aliases: []string{"gf"},
			Usage:   "Fetch all the existing features",
			Action: func(c *cli.Context) error {
				client := NewFlippyClient(c.GlobalString("url"))
				data, _ := client.GetFeatures()
				features, _ := data.GetStringArray("data")
				utils.Output("get-features", "flippy")
				for _, v := range features {
					fmt.Println(v)
				}
				return nil
			},
		},
		{
			Name:      "set-feature",
			Aliases:   []string{"sf"},
			Usage:     "Create/update the feature and set its value for the given scope",
			ArgsUsage: "<key> <scope> <value>",
			Action: func(c *cli.Context) error {
				client := NewFlippyClient(c.GlobalString("url"))
				if len(c.Args()) != 3 {
					return fmt.Errorf("Wrong number of args. See usage.")
				}
				data, err := client.SetFeature(c.Args().Get(1), c.Args().Get(0), c.Args().Get(2))
				if err != nil {
					return err
				}
				feature, _ := data.GetObject("data")
				utils.Output("set-feature", "flippy")
				fmt.Printf("%s\n", feature)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
