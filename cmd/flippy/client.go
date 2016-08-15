package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/gorilla/http"
)

const (
	GET_SCOPES_PATH   = "/admin/scopes"
	GET_FEATURES_PATH = "/admin/features"
	SET_FEATURE_PATH  = "/admin/features/"
)

type IFlippyHttpClient interface {
	GetJson(string) (*jason.Object, error)
	PostJson(string, []byte) (*jason.Object, error)
}

type FlippyHttpClient struct{}

type FlippyClient struct {
	flipadelphiaUrl string
	httpClient      IFlippyHttpClient
}

func (fhc FlippyHttpClient) GetJson(reqUrl string) (*jason.Object, error) {
	headers := map[string][]string{
		"Content-Type": []string{"application/json"},
	}
	rs, _, rc, err := http.DefaultClient.Get(reqUrl, headers)
	if err != nil {
		return nil, err
	}
	if rs.Code >= 400 {
		return nil, fmt.Errorf("%s", rs.String())
	}
	defer rc.Close()
	return jason.NewObjectFromReader(rc)
}

func (fhc FlippyHttpClient) PostJson(reqUrl string, postBody []byte) (*jason.Object, error) {
	headers := map[string][]string{
		"Content-Type": []string{"application/json"},
	}
	postBodyReader := bytes.NewBuffer(postBody)
	rs, _, rc, err := http.DefaultClient.Post(reqUrl, headers, postBodyReader)
	if err != nil {
		return nil, err
	}
	if rs.Code >= 400 {
		return nil, fmt.Errorf("%s", rs.String())
	}
	defer rc.Close()
	return jason.NewObjectFromReader(rc)
}

func NewFlippyClient(flipadelphiaUrl string) FlippyClient {
	if !strings.HasPrefix(flipadelphiaUrl, "http://") && !strings.HasPrefix(flipadelphiaUrl, "https://") {
		flipadelphiaUrl = "http://" + flipadelphiaUrl
	}
	return FlippyClient{
		flipadelphiaUrl: trimTrailingSlash(flipadelphiaUrl),
		httpClient:      FlippyHttpClient{},
	}
}

func trimLeadingSlash(s string) string {
	if strings.HasPrefix(s, "/") {
		return trimTrailingSlash(strings.TrimPrefix(s, "/"))
	}
	return s
}

func trimTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return trimTrailingSlash(strings.TrimSuffix(s, "/"))
	}
	return s
}

func (fc FlippyClient) getUrl(urlPath string) string {
	return fmt.Sprintf("%s/%s", fc.flipadelphiaUrl, trimLeadingSlash(urlPath))
}

func (fc FlippyClient) getJson(urlPath string) (*jason.Object, error) {
	reqUrl := fc.getUrl(urlPath)
	return fc.httpClient.GetJson(reqUrl)
}

func (fc FlippyClient) postJson(urlPath string, postBody []byte) (*jason.Object, error) {
	reqUrl := fc.getUrl(urlPath)
	return fc.httpClient.PostJson(reqUrl, postBody)
}

func (fc FlippyClient) GetScopes() (*jason.Object, error) {
	return fc.getJson(GET_SCOPES_PATH)
}

func (fc FlippyClient) GetFeatures() (*jason.Object, error) {
	return fc.getJson(GET_FEATURES_PATH)
}

func (fc FlippyClient) SetFeature(scope, key, value string) (*jason.Object, error) {
	body := fmt.Sprintf(`{"scope": "%v", "value": "%v"}`, scope, value)
	return fc.postJson(SET_FEATURE_PATH+key, []byte(body))
}
