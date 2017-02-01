package client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/antonholmquist/jason"
	"io"
)

const (
	GET_SCOPES_PATH         = "/admin/scopes"
	GET_SCOPE_FEATURES_PATH = "/features"
	GET_SCOPE_FEATURE_PATH = "/features/"
	GET_FEATURES_PATH       = "/admin/features"
	SET_FEATURE_PATH        = "/admin/features/"
)

type HttpClient interface {
	Do(*http.Request) (int, string, map[string][]string, io.ReadCloser, error)
}

type IFlippyHttpClient interface {
	// GetJson makes a GET request to the URL and returns the response body
	// as a *jason.Object.
	GetJson(string) (*jason.Object, error)

	// PostJson makes a POST request to the URL with the request body and
	// returns the response body as a *jason.Object.
	PostJson(string, []byte) (*jason.Object, error)
}

type FlippyHttpClient struct{}

type FlippyClient struct {
	FlipadelphiaUrl string
	HttpClient      IFlippyHttpClient
}

func (fhc FlippyHttpClient) GetJson(reqUrl string) (*jason.Object, error) {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	return jason.NewObjectFromReader(resp.Body)
}

func (fhc FlippyHttpClient) PostJson(reqUrl string, postBody []byte) (*jason.Object, error) {
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(postBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	return jason.NewObjectFromReader(resp.Body)
}

func (fc FlippyClient) UrlIsFlipadelphiaServer() bool {
	resp, err := http.Get(fc.FlipadelphiaUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if bytes.Equal(body, []byte("flipadelphia flips your features")) {
		return true
	}
	return false
}

func NewFlippyClient(flipadelphiaUrl string) (FlippyClient, error) {
	if !strings.HasPrefix(flipadelphiaUrl, "http://") && !strings.HasPrefix(flipadelphiaUrl, "https://") {
		flipadelphiaUrl = "http://" + flipadelphiaUrl
	}
	fc := FlippyClient{
		FlipadelphiaUrl: trimTrailingSlash(flipadelphiaUrl),
		HttpClient:      FlippyHttpClient{},
	}
	if !fc.UrlIsFlipadelphiaServer() {
		return fc, errors.New(fmt.Sprintf("URL %s is not a flipadelphia server", flipadelphiaUrl))
	}
	return fc, nil
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

func convertJasonObjectToStringMap(obj *jason.Object) (map[string]string, error) {
	stringMap := make(map[string]string)
	for key, val := range obj.Map() {
		valString, err := val.String()
		if err != nil {
			return nil, err
		}
		stringMap[key] = valString
	}
	return stringMap, nil
}

func (fc FlippyClient) getUrl(urlPath string) string {
	return fmt.Sprintf("%s/%s", fc.FlipadelphiaUrl, trimLeadingSlash(urlPath))
}

func (fc FlippyClient) getJson(urlPath string) (*jason.Object, error) {
	reqUrl := fc.getUrl(urlPath)
	return fc.HttpClient.GetJson(reqUrl)
}

func (fc FlippyClient) postJson(urlPath string, postBody []byte) (*jason.Object, error) {
	reqUrl := fc.getUrl(urlPath)
	return fc.HttpClient.PostJson(reqUrl, postBody)
}

func (fc FlippyClient) GetScopes() ([]string, error) {
	data, err := fc.getJson(GET_SCOPES_PATH)
	if err != nil {
		return nil, err
	}
	return data.GetStringArray("data")
}

func (fc FlippyClient) GetScopeFeatures(scope string) ([]string, error) {
	data, err := fc.getJson(fmt.Sprintf(`%s?scope=%s`, GET_SCOPE_FEATURES_PATH, scope))
	if err != nil {
		return nil, err
	}
	return data.GetStringArray("data")
}

func (fc FlippyClient) GetScopeFeature(scope, feature string) (map[string]string, error) {
	data, err := fc.getJson(fmt.Sprintf(`%s%s?scope=%s`, GET_SCOPE_FEATURE_PATH, feature, scope))
	if err != nil {
		return nil, err
	}
	featureObj, err := data.GetObject("data")
	if err != nil {
		return nil, err
	}
	return convertJasonObjectToStringMap(featureObj)
}

func (fc FlippyClient) GetFeatures() ([]string, error) {
	data, err := fc.getJson(GET_FEATURES_PATH)
	if err != nil {
		return nil, err
	}
	return data.GetStringArray("data")
}

func (fc FlippyClient) SetFeature(scope, key, value string) (map[string]string, error) {
	body := fmt.Sprintf(`{"scope": "%v", "value": "%v"}`, scope, value)
	data, err := fc.postJson(SET_FEATURE_PATH+key, []byte(body))
	if err != nil {
		return nil, err
	}
	feature, err := data.GetObject("data")
	if err != nil {
		return nil, err
	}
	return convertJasonObjectToStringMap(feature)
}
