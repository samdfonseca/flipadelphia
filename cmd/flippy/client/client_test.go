package client

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/antonholmquist/jason"
	. "github.com/smartystreets/goconvey/convey"
)

type MockFlippyHttpClient struct {
	OnGetJson  func(string) (*jason.Object, error)
	OnPostJson func(string, []byte) (*jason.Object, error)
}

type HandlerTransport struct {
	h http.Handler
}

type pipeResponseWriter struct {
	r     *io.PipeReader
	w     *io.PipeWriter
	resp  *http.Response
	ready chan<- struct{}
}

func (ht HandlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r, w := io.Pipe()
	resp := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       r,
		Request:    req,
	}
	ready := make(chan struct{})
	prw := &pipeResponseWriter{r, w, resp, ready}
	go func() {
		defer w.Close()
		ht.h.ServeHTTP(prw, req)
	}()
	<-ready
	return resp, nil
}

func (w *pipeResponseWriter) Header() http.Header {
	return w.resp.Header
}

func (w *pipeResponseWriter) Write(p []byte) (int, error) {
	if w.ready != nil {
		w.WriteHeader(http.StatusOK)
	}
	return w.w.Write(p)
}

func (w *pipeResponseWriter) WriteHeader(status int) {
	if w.ready == nil {
		// already called
		return
	}
	w.resp.StatusCode = status
	w.resp.Status = fmt.Sprintf("%d %s", status, http.StatusText(status))
	close(w.ready)
	w.ready = nil
}

func (mClient MockFlippyHttpClient) GetJson(reqUrl string) (*jason.Object, error) {
	return mClient.OnGetJson(reqUrl)
}

func (mClient MockFlippyHttpClient) PostJson(reqUrl string, postBody []byte) (*jason.Object, error) {
	return mClient.OnPostJson(reqUrl, postBody)
}

func TestFlippyClient_NewFlippyClient(t *testing.T) {
	Convey(`Given a flipadelphia base URL of "http://localhost:3006"`, t, func() {
		baseUrl := "http://localhost:3006"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "http://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "http://localhost:3006")
			})
		})
	})
	Convey(`Given a flipadelphia base URL of "https://localhost:3006"`, t, func() {
		baseUrl := "https://localhost:3006"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "https://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "https://localhost:3006")
			})
		})
	})
	Convey(`Given a flipadelphia base URL of "http://localhost:3006/"`, t, func() {
		baseUrl := "http://localhost:3006/"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "http://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "http://localhost:3006")
			})
		})
	})
	Convey(`Given a flipadelphia base URL of "https://localhost:3006/"`, t, func() {
		baseUrl := "https://localhost:3006/"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "https://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "https://localhost:3006")
			})
		})
	})
	Convey(`Given a flipadelphia base URL of "localhost:3006"`, t, func() {
		baseUrl := "localhost:3006"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "http://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "http://localhost:3006")
			})
		})
	})
	Convey(`Given a flipadelphia base URL of "localhost:3006/"`, t, func() {
		baseUrl := "localhost:3006/"
		Convey(`When creating a new FlippyClient instance`, func() {
			client, _ := NewFlippyClient(baseUrl)
			Convey(`The GetFlipadelphiaUrl method should return "http://localhost:3006"`, func() {
				So(client.FlipadelphiaUrl, ShouldEqual, "http://localhost:3006")
			})
		})
	})
}

func TestFlippyClient_GetJson(t *testing.T) {
	http.DefaultClient.Transport = HandlerTransport{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": ["scope1", "scope2", "scope3"]}`))
		}),
	}
	Convey(`Given a new flippy client instance`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient:      FlippyHttpClient{},
		}
		Convey(`When the client instance's GetJson method is called`, func() {
			data, err := client.HttpClient.GetJson(client.FlipadelphiaUrl)
			Convey(`The error value should be nil`, func() {
				So(err, ShouldBeNil)
			})
			Convey(`The returned *jason.Object has a "data" key`, func() {
				So(data.Map(), ShouldContainKey, "data")
			})
			scopes, _ := data.GetStringArray("data")
			Convey(`The "data" object should have a length of 3`, func() {
				So(scopes, ShouldHaveLength, 3)
			})
			Convey(`The "data" object should be a list containing "scope1", "scope2", and "scope3"`, func() {
				So(scopes, ShouldContain, "scope1")
				So(scopes, ShouldContain, "scope2")
				So(scopes, ShouldContain, "scope3")
			})
		})
	})
}

func TestFlippyClient_PostJson(t *testing.T) {
	http.DefaultClient.Transport = HandlerTransport{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": {
				"data": "true",
				"value": "1",
				"name": "feature123"
			}}`))
		}),
	}
	Convey(`Given a new flippy client instance`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient:      FlippyHttpClient{},
		}
		Convey(`When the client instance's PostJson method is called`, func() {
			data, err := client.HttpClient.PostJson(client.FlipadelphiaUrl, []byte(""))
			Convey(`The error value should be nil`, func() {
				So(err, ShouldBeNil)
			})
			Convey(`The returned *jason.Object has a "data" key`, func() {
				So(data.Map(), ShouldContainKey, "data")
			})
			feature, _ := data.GetObject("data")
			Convey(`The "data" object should have 3 keys`, func() {
				So(feature.Map(), ShouldHaveLength, 3)
			})
			Convey(`The "data" object should have the keys "data", "value", and "name"`, func() {
				So(feature.Map(), ShouldContainKey, "data")
				So(feature.Map(), ShouldContainKey, "value")
				So(feature.Map(), ShouldContainKey, "name")
			})
			Convey(`The "data" key should have the value "true"`, func() {
				featureData, _ := feature.GetString("data")
				So(featureData, ShouldEqual, "true")
			})
			Convey(`The "value" key should have the value "1"`, func() {
				featureValue, _ := feature.GetString("value")
				So(featureValue, ShouldEqual, "1")
			})
			Convey(`The "name" key should have the value "feature123"`, func() {
				featureName, _ := feature.GetString("name")
				So(featureName, ShouldEqual, "feature123")
			})
		})
	})
}

func TestFlippyClient_GetScopes(t *testing.T) {
	Convey(`Given a flipadelphia server with scopes "scope1", "scope2", and "scope3"`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient: MockFlippyHttpClient{
				OnGetJson: func(reqUrl string) (*jason.Object, error) {
					return jason.NewObjectFromBytes([]byte(`{"data": [
						"scope1",
						"scope2",
						"scope3"
					]}`))
				},
			},
		}
		Convey(`When the client fetches the scopes`, func() {
			actual, _ := client.GetScopes()
			Convey(`The number of returned scopes should be 3`, func() {
				So(actual, ShouldHaveLength, 3)
			})
			Convey(`The returned scopes should be "scope1", "scope2", and "scope3"`, func() {
				So(actual, ShouldContain, "scope1")
				So(actual, ShouldContain, "scope2")
				So(actual, ShouldContain, "scope3")
			})
		})
	})
}

func TestFlippyClient_GetScopeFeatures(t *testing.T) {
	Convey(`Given a flipadelphia server with features "feature1", "feature2", and "feature3" set on scope "scope1"`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient: MockFlippyHttpClient{
				OnGetJson: func(reqUrl string) (*jason.Object, error) {
					return jason.NewObjectFromBytes([]byte(`{"data": [
						"feature1",
						"feature2",
						"feature3"
					]}`))
				},
			},
		}
		Convey(`When the client fetches the scope's features`, func() {
			actual, _ := client.GetScopeFeatures("scope1")
			Convey(`The number of returned features should be 3`, func() {
				So(actual, ShouldHaveLength, 3)
			})
			Convey(`The returned features should be "feature1", "feature2", and "feature3"`, func() {
				So(actual, ShouldContain, "feature1")
				So(actual, ShouldContain, "feature2")
				So(actual, ShouldContain, "feature3")
			})
		})
	})
}

func TestFlippyClient_GetFeatures(t *testing.T) {
	Convey(`Given a flipadelphia server with features "feature1", "feature2", and "feature3"`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient: MockFlippyHttpClient{
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
		Convey(`When the client fetches the features`, func() {
			actual, _ := client.GetFeatures()
			Convey(`The number of returned features should be 3`, func() {
				So(actual, ShouldHaveLength, 3)
			})
			Convey(`The returned features should be "feature1", "feature2", and "feature3"`, func() {
				So(actual, ShouldContain, "feature1")
				So(actual, ShouldContain, "feature2")
				So(actual, ShouldContain, "feature3")
			})
		})
	})
}

func TestFlippyClient_SetFeature(t *testing.T) {
	Convey(`Given a flipadelphia server with feature "feature1" not set on scope "scope1"`, t, func() {
		client := FlippyClient{
			FlipadelphiaUrl: "localhost:3006",
			HttpClient: MockFlippyHttpClient{
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
		Convey(`When the client sets the feature "feature1" on scope "scope1" to value "1"`, func() {
			actual, err := client.SetFeature("scope1", "feature1", "1")
			Convey(`The error value should be nil`, func() {
				So(err, ShouldBeNil)
			})
			Convey(`The response data should be a JSON object with 3 keys`, func() {
				So(actual, ShouldHaveLength, 3)
			})
			Convey(`The response data should have the keys "name", "value", and "data"`, func() {
				So(actual, ShouldContainKey, "name")
				So(actual, ShouldContainKey, "value")
				So(actual, ShouldContainKey, "data")
			})
			Convey(`The "name" key should have value "feature1"`, func() {
				So(actual["name"], ShouldEqual, "feature1")
			})
			Convey(`The "value" key should have value "1"`, func() {
				So(actual["value"], ShouldEqual, "1")
			})
			Convey(`The "data" key should have value "true"`, func() {
				So(actual["data"], ShouldEqual, "true")
			})
		})
	})
}
