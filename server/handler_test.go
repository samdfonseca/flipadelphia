package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"
	"io/ioutil"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/samdfonseca/flipadelphia/store"
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
		OnCheckFeatureExists: func(feature []byte) bool {
			return true
		},
		OnCheckFeatureHasScope: func(scope, feature []byte) bool {
			return true
		},
		OnCheckScopeExists: func(scope []byte) bool {
			return true
		},
		OnCheckScopeHasFeature: func(scope, feature []byte) bool {
			return true
		},
	}
	server := httptest.NewServer(App(fdb, negroni.New(negroni.NewRecovery())))
	defer server.Close()

	resp, err := http.Get(getCheckFeatureURL(server.URL, "feature1", "user-1"))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"data":{"name":"feature1","value":"on","data":"true"}}`
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
		OnCheckFeatureExists: func(feature []byte) bool {
			return true
		},
		OnCheckFeatureHasScope: func(scope, feature []byte) bool {
			return true
		},
		OnCheckScopeExists: func(scope []byte) bool {
			return true
		},
		OnCheckScopeHasFeature: func(scope, feature []byte) bool {
			return true
		},
	}
	server := httptest.NewServer(App(fdb, negroni.New(negroni.NewRecovery())))
	defer server.Close()

	resp, err := http.Get(getCheckFeatureURL(server.URL, "feature1", "user-1"))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"data":{"name":"feature1","value":"","data":"false"}}`
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
		OnCheckFeatureExists: func(feature []byte) bool {
			return true
		},
		OnCheckFeatureHasScope: func(scope, feature []byte) bool {
			return true
		},
		OnCheckScopeExists: func(scope []byte) bool {
			return true
		},
		OnCheckScopeHasFeature: func(scope, feature []byte) bool {
			return true
		},
	}
	server := httptest.NewServer(App(fdb, negroni.New(negroni.NewRecovery())))
	defer server.Close()

	reqBody := `{"scope":"user-1","value":"on"}`
	resp, err := http.Post(getSetFeatureURL(server.URL, "feature1"), "application/json", strings.NewReader(reqBody))
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
	}

	target := `{"data":{"name":"feature1","value":"on","data":"true"}}`
	checkResult(string(body), target, t)
}

func TestGetScopesPaginatedWithoutOffset_ValidRequest(t *testing.T) {
	testScopes := store.StringSlice{"scope0", "scope1", "scope2", "scope3", "scope4"}
	fdb := store.MockPersistenceStore{
		OnGetScopesPaginated: func(offset, count int) (store.Serializable, error) {
			return testScopes, nil
		},
		OnCheckFeatureExists: func(feature []byte) bool {
			return true
		},
		OnCheckFeatureHasScope: func(scope, feature []byte) bool {
			return true
		},
		OnCheckScopeExists: func(scope []byte) bool {
			return true
		},
		OnCheckScopeHasFeature: func(scope, feature []byte) bool {
			return true
		},
	}
	server := httptest.NewServer(App(fdb, negroni.New(negroni.NewRecovery())))
	defer server.Close()

	resp, err := http.Get(fmt.Sprintf("%s/admin/scopes?count=5", server.URL))
	if err != nil {
		t.Error(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
	}

	checkResult(string(body), fmt.Sprintf(`{"data":%s}`, string(testScopes.Serialize())), t)
}
