package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/samdfonseca/flipadelphia/store"
)

func App(db store.PersistenceStore) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/features", checkScopeHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]*}")
	router.HandleFunc("/features/{feature_name}", checkFeatureHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]*}")
	router.HandleFunc("/features/{feature_name}", setFeatureHandler(db)).
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

func checkScopeHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := r.FormValue("scope")
		features, err := db.GetScopeFeatures([]byte(scope))
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(features.Serialize())
		}
	})
}

func setFeatureHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		feature_name := []byte(vars["feature_name"])
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		var bodyJson map[string]string
		json.Unmarshal(body, &bodyJson)
		scope := []byte(bodyJson["scope"])
		value := []byte(bodyJson["value"])
		_, err = db.Set(scope, feature_name, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			feature, _ := db.Get(scope, feature_name)
			w.WriteHeader(http.StatusOK)
			w.Write(feature.Serialize())
		}
	})
}
