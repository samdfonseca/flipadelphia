package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/antonholmquist/jason"
)

type MockFlippyHttpClient struct {
	OnGetJson  func(string) (*jason.Object, error)
	OnPostJson func(string, []byte) (*jason.Object, error)
}

func (mClient MockFlippyHttpClient) GetJson(reqUrl string) (*jason.Object, error) {
	return mClient.OnGetJson(reqUrl)
}

func (mClient MockFlippyHttpClient) PostJson(reqUrl string, postBody []byte) (*jason.Object, error) {
	return mClient.OnPostJson(reqUrl, postBody)
}

func assertEqual(actual, target string, t *testing.T) {
	if actual != target {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual value did not match target value")
	}
}

func assertNil(actual interface{}, t *testing.T) {
	if actual != nil {
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual is not nil")
	}
}

func assertErrorEqual(actual, target error, t *testing.T) {
	if fmt.Sprintf("%s", actual) != fmt.Sprintf("%s", target) {
		t.Logf("Target: %s", target)
		t.Logf("Actual: %s", actual)
		t.Errorf("Actual error did not match target error")
	}
}

func stringInSlice(items []string, item string) bool {
	for _, v := range items {
		if strings.Compare(item, v) == 0 {
			return true
		}
	}
	return false
}

func bytesInSlice(items [][]byte, item []byte) bool {
	for _, v := range items {
		if bytes.Compare(item, v) == 0 {
			return true
		}
	}
	return false
}

func intInSlice(items []int, item int) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func floatInSlice(items []float64, item float64) bool {
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func assertContains(items, item interface{}, t *testing.T) {
	var b bool
	s := reflect.ValueOf(items)
	if s.Kind() != reflect.Slice {
		t.Errorf("assertContains() given a non-slice type")
	}
	switch T := item.(type) {
	default:
		t.Errorf("Unexpected type %T\n", T)
	case string:
		stringItems := make([]string, s.Len())
		for i := 0; i < s.Len(); i++ {
			stringItems[i] = s.Index(i).String()
		}
		b = stringInSlice(stringItems, item.(string))
	case []byte:
		bytesItems := make([][]byte, s.Len())
		for i := 0; i < s.Len(); i++ {
			bytesItems[i] = s.Index(i).Bytes()
		}
		b = bytesInSlice(bytesItems, item.([]byte))
	case int:
		intItems := make([]int, s.Len())
		for i := 0; i < s.Len(); i++ {
			intItems[i] = int(s.Index(i).Int())
		}
		b = intInSlice(intItems, item.(int))
	case float64:
		floatItems := make([]float64, s.Len())
		for i := 0; i < s.Len(); i++ {
			floatItems[i] = s.Index(i).Float()
		}
		b = floatInSlice(floatItems, item.(float64))
	}
	if !b {
		if T := reflect.TypeOf(item).Kind(); T != reflect.Int && T != reflect.Float64 && T != reflect.String {
			bytesItems := make([][]byte, s.Len())
			for i := 0; i < s.Len(); i++ {
				bytesItems[i] = s.Index(i).Bytes()
			}
			t.Errorf("Item '%s' not found in '%s'", item, bytesItems)
		}
		t.Errorf("Item '%v' not found in '%v'", item, items)
	}
}

func TestFlippyClient_GetScopes(t *testing.T) {
	client := FlippyClient{
		flipadelphiaUrl: "localhost:3006",
		httpClient: MockFlippyHttpClient{
			OnGetJson: func(reqUrl string) (*jason.Object, error) {
				jsonBody := []byte(`{"data": [
					"scope1",
					"scope2",
					"scope3"
				]}`)
				return jason.NewObjectFromBytes(jsonBody)
			},
		},
	}
	scopes, _ := client.GetScopes()
	actual, _ := scopes.GetStringArray("data")
	target := []string{"scope1", "scope2", "scope3"}
	if len(actual) != len(target) {
		t.Logf("Target: %v\n", target)
		t.Logf("Actual: %v\n", scopes)
		t.Error("Length of returned scopes array not equal to length of target array")
	}
	for _, v := range actual {
		assertContains(target, v, t)
	}
}

func TestFlippyClient_GetFeatures(t *testing.T) {
	client := FlippyClient{
		flipadelphiaUrl: "localhost:3006",
		httpClient: MockFlippyHttpClient{
			OnGetJson: func(reqUrl string) (*jason.Object, error) {
				jsonBody := []byte(`{"data": [
					"feature1",
					"feature2",
					"feature3"
				]}`)
				return jason.NewObjectFromBytes(jsonBody)
			},
		},
	}
	features, _ := client.GetFeatures()
	actual, _ := features.GetStringArray("data")
	target := []string{"feature1", "feature2", "feature3"}
	if len(actual) != len(target) {
		t.Logf("Target: %v\n", target)
		t.Logf("Actual: %v\n", features)
		t.Error("Length of returned features array not equal to length of target array")
	}
	for _, v := range actual {
		assertContains(target, v, t)
	}
}

func TestFlippyClient_SetFeature(t *testing.T) {
	client := FlippyClient{
		flipadelphiaUrl: "localhost:3006",
		httpClient: MockFlippyHttpClient{
			OnPostJson: func(reqUrl string, postBody []byte) (*jason.Object, error) {
				jsonBody := []byte(`{"data": {
					"name": "feature1",
					"value": "1",
					"data": "true"
				}}`)
				return jason.NewObjectFromBytes(jsonBody)
			},
		},
	}
	feature, _ := client.SetFeature("scope1", "feature1", "1")
	data, _ := feature.GetObject("data")
	target := map[string]string{"name": "feature1", "value": "1", "data": "true"}
	if len(data.Map()) != len(target) {
		t.Logf("Target: %v\n", target)
		t.Logf("Actual: %v\n", data.Map())
		t.Error("Returned feature does not have the expected number of keys")
	}
	for k, v := range data.Map() {
		if _, ok := target[k]; !ok {
			t.Logf("Target: %v\n", target)
			t.Logf("Actual: %v\n", data.Map())
			t.Errorf("Returned feature has an unexpected key: %s\n", k)
		}
		vs, _ := v.String()
		assertEqual(target[k], vs, t)
	}
}
