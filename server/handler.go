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

func App(db store.PersistenceStore, auth Authenticator) http.Handler {
	router := mux.NewRouter()
	admin_router := router.PathPrefix("/admin").Subrouter()

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/features", checkScopeFeaturesForValueHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]*}", "value", "{value:[0-9A-Za-z_-]*}")
	router.HandleFunc("/features", checkAllScopeFeaturesHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]*}")
	router.HandleFunc("/features/{feature_name}", checkFeatureHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]*}")
	admin_router.HandleFunc("/features/{feature_name}", setFeatureHandler(db, auth)).
		Methods("PUT", "POST")
	n := negroni.Classic()
	n.UseHandler(router)
	return n
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("flipadelphia flips your features"))
}

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

func setFeatureHandler(db store.PersistenceStore, auth Authenticator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthed, err := auth.AuthenticateRequest(r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		} else if !isAuthed {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		vars := mux.Vars(r)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		var setFeatureOptions store.FlipadelphiaSetFeatureOptions
		json.Unmarshal(body, &setFeatureOptions)
		setFeatureOptions.Key = []byte(vars["feature_name"])
		_, err = db.Set(setFeatureOptions.Scope, setFeatureOptions.Key, setFeatureOptions.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			feature, _ := db.Get(setFeatureOptions.Scope, setFeatureOptions.Key)
			w.WriteHeader(http.StatusOK)
			w.Write(feature.Serialize())
		}
	})
}
