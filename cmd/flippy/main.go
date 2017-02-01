package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/abiosoft/ishell"
	"github.com/samdfonseca/flipadelphia/cmd/flippy/client"
	"github.com/samdfonseca/flipadelphia/cmd/flippy/commands"
)

var (
	flippyVersion string

	flipadelphiaUrl string
)

func init() {
	flippyVersion = "dev-build"
	if v := os.Getenv("FLIPPY_BUILD_VERSION"); v != "" {
		flippyVersion = v
	}

	flag.StringVar(&flipadelphiaUrl, "url", "http://localhost:3006", "Base URL of the flipadelphia server")
}

func main() {
	shell := ishell.New()

	shell.Println(fmt.Sprintf(`Flippy interactive shell. Version %s`, flippyVersion))
	shell.Println(fmt.Sprintf(`Flipadelphia base URL: %s`, flipadelphiaUrl))

	client, err := client.NewFlippyClient(flipadelphiaUrl)
	if err != nil {
		shell.Println(err)
		return
	}

	shell.AddCmd(&ishell.Cmd{
		Name: "features",
		Help: "Get all features matching an optional regex: features [<feature regex>]",
		Func: commands.NewGetFeaturesFunc(client),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "scope-features",
		Help: "Get all features for a scope: scope-features <scope>",
		Func: commands.NewGetScopeFeaturesFunc(client),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "scope-feature",
		Help: "Get a feature set on a scope: scope-feature <scope> <feature>",
		Func: commands.NewGetScopeFeatureFunc(client),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "scopes",
		Help: "Get all scopes matching an optional regex: scopes [<scope regex>]",
		Func: commands.NewGetScopesFunc(client),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set-scope-feature",
		Help: "Sets a feature on a scope: set-scope-feature <scope> <feature> <value>",
		Func: commands.NewSetScopeFeatureFunc(client),
	})

	shell.Start()
}
