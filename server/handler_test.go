package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"
	"github.com/samdfonseca/flipadelphia/store"
	"io/ioutil"
	"strings"
)

func checkResult(actual, target string, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func getCheckFeatureURL(server, feature, scope string) string {
	return fmt.Sprintf("%s/features/%s?scope=%s", server, feature, scope)
}

func getSetFeatureURL(server, feature string) string {
	return fmt.Sprintf("%s/admin/features/%s", server, feature)
}

func getCheckAllScopeFeaturesURL(server, scope string) string {
	return fmt.Sprintf("%s/features?scope=%s", server, scope)
}

func getCheckScopeFeaturesForValueURL(server, scope, value string) string {
	return fmt.Sprintf("%s/features?scope=%s&value=%s", server, scope, value)
}

func TestCheckFeatureHandler_ValidRequest_PresetFeature(t *testing.T) {
	fdb := store.MockPersistenceStore{
		OnGet: func(scope, key []byte) (store.Serializable, error) {
			return store.FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "on",
				Data:  "true",
			}, nil
		},
	}
	server := httptest.NewServer(App(fdb, AuthSettings{}))
	defer server.Close()

	resp, err := http.Get(getCheckFeatureURL(server.URL, "feature1", "user-1"))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"name":"feature1","value":"on","data":"true"}`
	checkResult(string(body), target, t)
}

func TestCheckFeatureHandler_ValidRequest_UnsetFeature(t *testing.T) {
	fdb := store.MockPersistenceStore{
		OnGet: func(scope, key []byte) (store.Serializable, error) {
			return store.FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "",
				Data:  "false",
			}, nil
		},
	}
	server := httptest.NewServer(App(fdb, AuthSettings{}))
	defer server.Close()

	resp, err := http.Get(getCheckFeatureURL(server.URL, "feature1", "user-1"))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"name":"feature1","value":"","data":"false"}`
	checkResult(string(body), target, t)
}

func TestSetFeatureHandler_ValidRequest(t *testing.T) {
	fdb := store.MockPersistenceStore{
		OnGet: func(scope, key []byte) (store.Serializable, error) {
			return store.FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "on",
				Data:  "true",
			}, nil
		},
		OnSet: func(scope, key, value []byte) (store.Serializable, error) {
			return store.FlipadelphiaFeature{
				Name:  fmt.Sprintf("%s", key),
				Value: "on",
				Data:  "true",
			}, nil
		},
	}
	auth := MockAuth{
		OnAuthenticateRequest: func(r *http.Request) (bool, error) {
			return true, nil
		},
	}
	server := httptest.NewServer(App(fdb, auth))
	defer server.Close()

	reqBody := `{"scope":"user-1","value":"on"}`
	resp, err := http.Post(getSetFeatureURL(server.URL, "feature1"), "application/json", strings.NewReader(reqBody))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"name":"feature1","value":"on","data":"true"}`
	checkResult(string(body), target, t)
}
