package main
import (
	"os"
	"fmt"
	"github.com/fatih/color"
	//"github.com/google/logger"
)

//var defaultLogFileName = fmt.Sprintf("%s.log", FC.RuntimeEnvironment.EnvironmentName)
//var logPath = flag.String("logfile", getStoredFilePath(defaultLogFileName), "Full path to the logfile (doesn't have to exist yet)"
//var verbose = flag.Bool("verbose", false, "Print info level logs to stdout")

func exitErr(err error) {
	output(color.RedString(err.Error()))
	os.Exit(1)
}

func output(s string) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("flipadelphia"), s)
}

//func logit(s string) {
//
//}

func failOnError(err error, msg string, appendErr bool) {
	if err != nil {
		if appendErr {
			exitErr(fmt.Errorf("%s: %s", msg, err))
		} else {
			exitErr(fmt.Errorf("%s", msg))
		}
	}
}
