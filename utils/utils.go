package utils

import (
	"encoding/binary"
	"fmt"
)

func logErr(err error) {
	if err != nil {
		Output(err.Error())
	}
}

func exitErr(err error) {
	panic(err)
}

func Output(s string, appName ...string) {
	if len(appName) == 0 {
		appName = append(appName, "flipadelphia")
	}
	fmt.Printf("%s: %s\n", appName[0], s)
}

func LogOnError(err error, msg string, appendErr ...bool) {
	if err != nil {
		if len(appendErr) == 0 {
			LogOnError(err, msg, true)
		}
		if len(appendErr) >= 1 {
			if appendErr[0] {
				logErr(fmt.Errorf("ERROR - %s: %s", msg, err))
			} else {
				logErr(fmt.Errorf("ERROR - %s", msg))
			}
		}
	}
}

func LogOnSuccess(err error, msg string) {
	if err == nil {
		Output(fmt.Sprintf("SUCCESS - %s", msg))
	}
}

func LogEither(err error, successMsg, errorMsg string, appendErr ...bool) {
	LogOnSuccess(err, successMsg)
	LogOnError(err, errorMsg, appendErr...)
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

func Itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	fmt.Printf("------------- b: %i", b)
	return b
}

func Btos(b []byte) string {
	return fmt.Sprint(b)
}
