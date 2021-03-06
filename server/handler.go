package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/samdfonseca/flipadelphia/store"
	"github.com/samdfonseca/flipadelphia/utils"
)

func ClassicNegroniStack() *negroni.Negroni {
	return negroni.Classic()
}

func App(db store.PersistenceStore, n *negroni.Negroni) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)

	// c := cors.New(cors.Options{
	// 	AllowedOrigins: "*",

	// GET /features?scope=...&value=...
	router.HandleFunc("/features", checkScopeFeaturesForValueHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}", "value", "{value:[0-9A-Za-z_-]+}")
	// GET /features?scope=...
	router.HandleFunc("/features", checkAllScopeFeaturesHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}")
	// GET /scopes/{scope_name}
	router.HandleFunc("/scopes/{scope_name}", checkAllScopeFeaturesHandler(db)).
		Methods("GET").
		Queries("scope", "{scope_name:[0-9A-Za-z_-]+}")
	// GET /features/{feature_name}?scope=...
	router.HandleFunc("/features/{feature_name}", checkFeatureHandler(db)).
		Methods("GET").
		Queries("scope", "{scope:[0-9A-Za-z_-]+}")

	// POST /admin/features/{feature_name}
	router.HandleFunc("/admin/features/{feature_name}", setFeatureHandler(db)).
		Methods("POST")
	// GET /admin/scopes?prefix=...
	router.HandleFunc("/admin/scopes", getScopesWithPrefixHandler(db)).
		Methods("GET").
		Queries("prefix", "{prefix:[0-9A-Za-z_-]+}")
	// GET /admin/scopes?feature=...
	router.HandleFunc("/admin/scopes", getScopesWithFeatureHandler(db)).
		Methods("GET").
		Queries("feature", "{feature:.+}")
	// GET /admin/scopes?count=...
	router.HandleFunc("/admin/scopes", getScopesPaginatedHandler(db)).
		Methods("GET").
		Queries("count", "{count:[0-9]+}")
	// GET /admin/scopes?count=...&offset=...
	router.HandleFunc("/admin/scopes", getScopesPaginatedHandler(db)).
		Methods("GET").
		Queries("count", "{count:[0-9]+}", "offset", "{offset:[0-9]+}")
	// GET /admin/features?count=...&offset=...
	router.HandleFunc("/admin/features", getFeaturesPaginatedHandler(db)).
		Methods("GET").
		Queries("count", "{count:[0-9]+}", "offset", "{offset:[0-9]+}")
	// GET /admin/scopes
	router.HandleFunc("/admin/scopes", getScopesHandler(db)).
		Methods("GET")
	// GET /admin/features
	router.HandleFunc("/admin/features", getAllFeaturesHandler(db)).
		Methods("GET")
	// GET /admin/scopes/{scope}/features
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

	n.UseFunc(allowCORSOnRequestOrigin)
	n.UseFunc(responseContentTypeJson)
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

func responseContentTypeJson(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	next(w, r)
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
		defer r.Body.Close()
		if len(r.Form) != 1 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		scope := r.FormValue("scope")
		feature_name := vars["feature_name"]
		if featureHasScope := db.CheckFeatureHasScope([]byte(scope), []byte(feature_name)); !featureHasScope {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		feature, err := db.Get([]byte(scope), []byte(feature_name))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteResponseBody(feature, w)
	})
}

// Handler for GET to "/features?scope=..." and "/scopes/{scopes_name}"
func checkAllScopeFeaturesHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		defer r.Body.Close()
		if len(r.Form) != 1 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		scope := r.FormValue("scope")
		if scopeExists := db.CheckScopeExists([]byte(scope)); !scopeExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
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
		defer r.Body.Close()
		if len(r.Form) != 2 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		scope := r.FormValue("scope")
		value := r.FormValue("value")
		if scopeExists := db.CheckScopeExists([]byte(scope)); !scopeExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		features, err := db.GetScopeFeaturesFilterByValue([]byte(scope), []byte(value))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteResponseBody(features, w)
	})
}

// Handler for POST to "/admin/features/{feature_name}"
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

func getScopesPaginatedHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 1 && len(r.Form) != 2 {
			utils.Output(fmt.Sprintf("len(r.Form) = %s", len(r.Form)))
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}

		count, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Unable to parse 'count' param in query: %q", r.Form.Encode())))
			return
		}

		offset, err := strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			offset = 0
		}
		scopes, err := db.GetScopesPaginated(offset, count)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(scopes, w)
	})
}

func getFeaturesPaginatedHandler(db store.PersistenceStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if len(r.Form) != 1 && len(r.Form) != 2 {
			utils.Output(fmt.Sprintf("len(r.Form) = %s", len(r.Form)))
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}

		count, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Unable to parse 'count' param in query: %q", r.Form.Encode())))
			return
		}

		offset, err := strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			offset = 0
		}
		features, err := db.GetFeaturesPaginated(offset, count)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s", err)))
			return
		}
		WriteResponseBody(features, w)
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
		defer r.Body.Close()
		if len(r.Form) != 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			w.Write([]byte(fmt.Sprintf("Unrecognized query: %q", r.Form.Encode())))
			return
		}
		vars := mux.Vars(r)
		if scopeExists := db.CheckScopeExists([]byte(vars["scope"])); !scopeExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
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
