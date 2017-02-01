package commands

import (
	"fmt"
	"regexp"

	"github.com/abiosoft/ishell"
	"github.com/samdfonseca/flipadelphia/cmd/flippy/client"
)


func printFeature(feature map[string]string, c *ishell.Context) {
	data := feature["data"]
	name := feature["name"]
	value := feature["value"]
	c.Println("name: "+name)
	c.Println("data: "+data)
	c.Println("value: "+value)
	c.Println("")
}

func NewGetFeaturesFunc(fc client.FlippyClient) func(*ishell.Context) {
	return func(c *ishell.Context) {
		var featureRegex *regexp.Regexp
		var features []string
		var err error
		if len(c.Args) == 0 {
			c.Args = append(c.Args, `.*`)
		}
		if len(c.Args) >= 2 {
			c.Println(`Too many arguments`)
			c.Println(c.Cmd.HelpText())
			return
		}
		featureRegex, err = regexp.Compile(c.Args[0])
		if err != nil {
			c.Println(err)
			return
		}
		features, err = fc.GetFeatures()
		if err != nil {
			c.Println(err)
			return
		}
		for _, feature := range features {
			if featureRegex.Match([]byte(feature)) {
				c.Println(feature)
			}
		}
		c.Println("")
	}
}

func NewGetScopeFeaturesFunc(fc client.FlippyClient) func(*ishell.Context) {
	getScopeFeaturesList := func(c *ishell.Context) {
		features, err := fc.GetScopeFeatures(c.Args[0])
		if err != nil {
			c.Println(err)
			return
		}
		for _, feature := range features {
			c.Println(feature)
		}
		c.Println("")
	}
	getScopeFeaturesFull := func(c *ishell.Context) {
		features, err := fc.GetScopeFeatures(c.Args[1])
		if err != nil {
			c.Println(err)
			return
		}
		for _, f := range features {
			feature, err := fc.GetScopeFeature(c.Args[1], f)
			if err != nil {
				c.Println(err)
				c.Println("")
				continue
			}
			printFeature(feature, c)
		}
	}
	return func(c *ishell.Context) {
		if len(c.Args) == 0 {
			c.Println(`Missing argument`)
			c.Println(c.Cmd.HelpText())
			return
		}
		if len(c.Args) >= 3 {
			c.Println(`Too many arguments`)
			c.Println(c.Cmd.HelpText())
			return
		}
		if len(c.Args) == 1 {
			getScopeFeaturesList(c)
		}
		if len(c.Args) == 2 {
			if c.Args[0] != "-full" {
				c.Println(fmt.Sprintf(`Unrecognized argument: %s`, c.Args[0]))
				c.Println(c.Cmd.HelpText())
				return
			}
			getScopeFeaturesFull(c)
		}
	}
}

func NewGetScopeFeatureFunc(fc client.FlippyClient) func(*ishell.Context) {
	return func(c *ishell.Context) {
		if len(c.Args) <= 1 {
			c.Println(`Missing argument`)
			c.Println(c.Cmd.HelpText())
			return
		}
		if len(c.Args) >= 3 {
			c.Println(`Too many arguments`)
			c.Println(c.Cmd.HelpText())
			return
		}
		feature, err := fc.GetScopeFeature(c.Args[0], c.Args[1])
		if err != nil {
			c.Println(err)
			return
		}
		printFeature(feature, c)
	}
}

func NewGetScopesFunc(fc client.FlippyClient) func(*ishell.Context) {
	return func(c *ishell.Context) {
		if len(c.Args) == 0 {
			c.Args = append(c.Args, `.*`)
		}
		if len(c.Args) >= 2 {
			c.Println(`Too many arguments`)
			c.Println(c.Cmd.HelpText())
			return
		}
		scopeRegex, err := regexp.Compile(c.Args[0])
		if err != nil {
			c.Println(err)
			return
		}
		scopes, err := fc.GetScopes()
		if err != nil {
			c.Println(err)
			return
		}
		for _, scope := range scopes {
			if scopeRegex.Match([]byte(scope)) {
				c.Println(scope)
			}
		}
		c.Println("")
	}
}

func NewSetScopeFeatureFunc(fc client.FlippyClient) func(*ishell.Context) {
	return func(c *ishell.Context) {
		if len(c.Args) <= 2 {
			c.Println(`Missing scope, feature or value argument`)
			c.Println(c.Cmd.HelpText())
			return
		}
		if len(c.Args) >= 4 {
			c.Println(`To many scope, feature or value arguments`)
			c.Println(c.Cmd.HelpText())
			return
		}
		feature, err := fc.SetFeature(c.Args[0], c.Args[1], c.Args[2])
		if err != nil {
			c.Println(err)
			return
		}
		printFeature(feature, c)
	}
}
