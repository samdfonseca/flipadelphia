package utils

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func logErr(err error) {
	if err != nil {
		Output(color.RedString(err.Error()))
	}
}

func exitErr(err error) {
	logErr(err)
	os.Exit(1)
}

func Output(s string, appName ...string) {
	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	if len(appName) == 0 {
		appName = append(appName, "flipadelphia")
	}
	fmt.Printf("%s: %s\n", green(bold(appName[0])), s)
}

func LogOnError(err error, msg string, appendErr ...bool) {
	if err != nil {
		if len(appendErr) > 0 {
			if appendErr[0] {
				logErr(fmt.Errorf("%s: %s", msg, err))
			} else {
				logErr(fmt.Errorf("%s", msg))
			}
		}
	}
}

func LogOnSuccess(err error, msg string) {
	if err == nil {
		Output(msg)
	}
}

func LogEither(err error, successMsg, errorMsg string, appendErr ...bool) {
	LogOnSuccess(err, fmt.Sprintf("SUCCESS %s", successMsg))
	LogOnError(err, fmt.Sprintf("FAIL %s", errorMsg), appendErr...)
}

func FailOnError(err error, msg string, appendErr bool) {
	if err != nil {
		if appendErr {
			exitErr(fmt.Errorf("%s: %s", msg, err))
		} else {
			exitErr(fmt.Errorf("%s", msg))
		}
	}
}
