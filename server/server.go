package server

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/samdfonseca/flipadelphia/db"
	"github.com/samdfonseca/flipadelphia/utils"
)

func App(fdb db.FlipadelphiaDB) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/features/{feature_name}", checkFeatureHandler(fdb)).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(router)
	return n
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("flipadelphia flips your features"))
}

func checkFeatureHandler(fdb db.FlipadelphiaDB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		scope := r.FormValue("scope")
		feature_name := vars["feature_name"]
		utils.Output("Scope: " + scope + ", Feature Name: " + feature_name)
		feature, err := fdb.Get([]byte(scope), []byte(feature_name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(feature.Serialize())
		}
	})
}
