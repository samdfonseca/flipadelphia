package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samdfonseca/flipadelphia/utils"
)

type AuthSettings struct {
	Url               string
	Method            string
	Header            string
	SuccessStatusCode string
}

type Authenticator interface {
	AuthenticateRequest(*http.Request) (bool, error)
}

func NewAuthSettings(url, method, header, successCode string) AuthSettings {
	var Method string
	for _, m := range []string{"GET", "HEAD", "POST", "PUT"} {
		if strings.ToUpper(method) == m {
			Method = method
		}
	}
	if strings.EqualFold(Method, "") {
		utils.FailOnError(fmt.Errorf(""), "Invalid request method %q", false)
	}
	return AuthSettings{Url: url, Method: Method, Header: header, SuccessStatusCode: successCode}
}

func (auth AuthSettings) AuthenticateRequest(r *http.Request) (bool, error) {
	client := http.Client{}
	req, err := http.NewRequest(auth.Method, auth.Url, nil)
	if err != nil {
		utils.LogOnError(err, "FAILED AUTH", true)
		return false, err
	}
	header := r.Header.Get(auth.Header)
	if header == "" {
		err = fmt.Errorf("Missing %q header", auth.Header)
		utils.LogOnError(err, "FAILED AUTH", true)
		return false, err
	}
	req.Header.Set(auth.Header, header)
	resp, err := client.Do(req)
	if err != nil {
		utils.LogOnError(err, "FAILED AUTH", true)
		return false, err
	}
	isAuthorized := string(resp.StatusCode) == auth.SuccessStatusCode
	return isAuthorized, nil
}
