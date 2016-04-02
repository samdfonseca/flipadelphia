package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
		Methods("PUT", "POST")
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

	n := negroni.Classic()
	n.UseHandler(router)
	return n
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("flipadelphia flips your features"))
}

// Handler for GET to "/features/{feature_name}?scope=..."
func checkFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		scope := r.FormValue("scope")
		utils.Output(fmt.Sprintf("Scope: %q", scope))
		feature_name := vars["feature_name"]
		feature, err := db.Get([]byte(scope), []byte(feature_name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(feature.Serialize())
		}
	})
}

// Handler for GET to "/features?scope=..."
func checkAllScopeFeaturesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := r.FormValue("scope")
		defer r.Body.Close()
		features, err := db.GetScopeFeatures([]byte(scope))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(features.Serialize())
		}
	})
}

// Handler for GET to "/features?scope=...&value=..."
func checkScopeFeaturesForValueHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := r.FormValue("scope")
		value := r.FormValue("value")
		defer r.Body.Close()
		features, err := db.GetScopeFeaturesFilterByValue([]byte(scope), []byte(value))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(features.Serialize())
		}
	})
}

// Handler for "/admin/features/{feature_name}"
func setFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		utils.Output(string(body))
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
		} else {
			feature, _ := db.Get([]byte(setFeatureOptions.Scope), []byte(setFeatureOptions.Key))
			w.WriteHeader(http.StatusOK)
			w.Write(feature.Serialize())
		}
	})
}

// Handler for GET to "/admin/scopes"
func getScopesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes, err := db.GetScopes()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(scopes.Serialize())
		}
	})
}

// Handler for GET to "/admin/scopes?prefix=..."
func getScopesWithPrefixHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		prefix := vars["prefix"]
		scopes, err := db.GetScopesWithPrefix([]byte(prefix))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(scopes.Serialize())
		}
	})
}

// Handler for GET to "/admin/scopes?feature=..."
func getScopesWithFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		feature := vars["feature"]
		scopes, err := db.GetScopesWithFeature([]byte(feature))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(scopes.Serialize())
		}
	})
}

// Handler for GET to "/admin/features"
func getAllFeaturesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		features, err := db.GetFeatures()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(features.Serialize())
		}
	})
}
