package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/samdfonseca/flipadelphia/store"
	"github.com/samdfonseca/flipadelphia/utils"
)

func App(db store.PersistenceStore) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)

	router.HandleFunc("/features", checkScopeFeaturesForValueHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}", "value", "{value:[0-9A-Za-z_-]+}")
	router.HandleFunc("/features", checkAllScopeFeaturesHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}")
	router.HandleFunc("/features/{feature_name}", checkFeatureHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}")

	router.HandleFunc("/admin/features/{feature_name}", setFeatureHandler(db)).
		Methods("POST")
	router.HandleFunc("/admin/scopes", getScopesWithPrefixHandler(db)).
		Methods("GET").
		Queries("prefix", "{prefix:[0-9A-Za-z_-]+}")
	router.HandleFunc("/admin/scopes", getScopesWithFeatureHandler(db)).
		Methods("GET").
		Queries("feature", "{feature:.+}")
	router.HandleFunc("/admin/scopes", getScopesHandler(db)).
		Methods("GET")
	router.HandleFunc("/admin/features", getAllFeaturesHandler(db)).
		Methods("GET")
	router.HandleFunc("/admin/scopes/{scope:[0-9A-Za-z_-]+}/features", getScopeFeaturesFullHandler(db)).
		Methods("GET")

	router.HandleFunc("/features", allowCORSHandler("GET", "OPTIONS")).
		Methods("OPTIONS")
	router.HandleFunc("/features/{feature_name}", allowCORSHandler("GET", "OPTIONS")).
		Methods("OPTIONS")
	router.HandleFunc("/admin/features/{feature_name}", allowCORSHandler("POST", "OPTIONS")).
		Methods("OPTIONS")
	router.HandleFunc("/admin/scopes", allowCORSHandler("GET", "OPTIONS")).
		Methods("OPTIONS")
	router.HandleFunc("/admin/features", allowCORSHandler("GET", "OPTIONS")).
		Methods("OPTIONS")

	n := negroni.Classic()
	n.UseFunc(allowCORSOnRequestOrigin)
	n.UseHandler(router)
	return n
}

func WriteResponseBody(s store.Serializable, w http.ResponseWriter) {
	body, err := json.Marshal(map[string]store.Serializable{"data": s})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("flipadelphia flips your features"))
}

func allowCORSOnRequestOrigin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	next(w, r)
}

// Handler for GET to "/features/{feature_name}?scope=..."
func checkFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 1 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		scope := r.FormValue("scope")
		utils.Output(fmt.Sprintf("Scope: %q", scope))
		feature_name := vars["feature_name"]
		feature, err := db.Get([]byte(scope), []byte(feature_name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteResponseBody(feature, w)
	})
}

// Handler for GET to "/features?scope=..."
func checkAllScopeFeaturesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 1 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		scope := r.FormValue("scope")
		defer r.Body.Close()
		features, err := db.GetScopeFeatures([]byte(scope))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteResponseBody(features, w)
	})
}

// Handler for GET to "/features?scope=...&value=..."
func checkScopeFeaturesForValueHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 2 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		scope := r.FormValue("scope")
		value := r.FormValue("value")
		defer r.Body.Close()
		features, err := db.GetScopeFeaturesFilterByValue([]byte(scope), []byte(value))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteResponseBody(features, w)
	})
}

// Handler for "/admin/features/{feature_name}"
func setFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error reading request body"))
			return
		}
		var setFeatureOptions store.FlipadelphiaSetFeatureOptions
		err = json.Unmarshal(body, &setFeatureOptions)
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			errMsg := fmt.Sprintf("Unprocessable entity: %s", err.Error())
			w.Write([]byte(errMsg))
			return
		}
		setFeatureOptions.Key = vars["feature_name"]
		_, err = db.Set([]byte(setFeatureOptions.Scope), []byte(setFeatureOptions.Key), []byte(setFeatureOptions.Value))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		feature, _ := db.Get([]byte(setFeatureOptions.Scope), []byte(setFeatureOptions.Key))
		WriteResponseBody(feature, w)
	})
}

// Handler for GET to "/admin/scopes"
func getScopesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		scopes, err := db.GetScopes()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(scopes, w)
	})
}

// Handler for GET to "/admin/scopes?prefix=..."
func getScopesWithPrefixHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Form) != 1 {
			r.Form.Del("prefix")
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		prefix := vars["prefix"]
		scopes, err := db.GetScopesWithPrefix([]byte(prefix))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(scopes, w)
	})
}

// Handler for GET to "/admin/scopes?feature=..."
func getScopesWithFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 1 {
			r.Form.Del("feature")
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		feature := vars["feature"]
		scopes, err := db.GetScopesWithFeature([]byte(feature))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(scopes, w)
	})
}

// Handler for GET to "/admin/features"
func getAllFeaturesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		features, err := db.GetFeatures()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(features, w)
	})
}

// Handler for GET to "/admin/scopes/{scope}/features"
func getScopeFeaturesFullHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		features, err := db.GetScopeFeaturesFull([]byte(vars["scope"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(features, w)
	})
}

// Handler for OPTIONS on all endpoints
func allowCORSHandler(allowMethods ...string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(allowMethods) == 0 {
			allowMethods = append(allowMethods, "GET", "OPTIONS", "POST")
		}
		methods := strings.Join(allowMethods, ", ")
		w.Header().Set("Access-Control-Allow-Methods", methods)
		return
	})
}
