package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"appengine"
	"appengine/datastore"
)

const (
	currentVersion = "0.1"
	entityName     = "Instances"
)

var appIDPattern = regexp.MustCompile(`appid: (.*))`)

func init() {
	http.HandleFunc("/register", Register)
}

type Instance struct {
	Version string `datastore:"version"`
}

type RegisterResponse struct {
	Version string
}

func Register(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Println(r.Method)
	// Only accept post requests.
	if r.Method != "POST" {
		c.Errorf("Invalid method: %s", r.Method)
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}
	// These headers are set by app engine URLFetchService and can not be modified by the application.
	if !strings.Contains(r.UserAgent(), "AppEngine-Google") {
		c.Errorf("Invalid UserAgent: %s", r.UserAgent())
		http.Error(w, "Bad Request.", http.StatusBadRequest)
		return
	}
	m := appIDPattern.FindStringSubmatch(r.UserAgent())
	if m == nil || len(m) != 2 {
		c.Errorf("Invalid UserAgent: %s. RegexpMatch:%s", r.UserAgent(), m)
		http.Error(w, "Bad Request.", http.StatusBadRequest)
		return
	}
	appID := m[1]
	if !strings.Contains(appID, "isumm") && !strings.Contains(appID, "~dev") {
		c.Errorf("Invalid UserAgent: %s", r.UserAgent())
		http.Error(w, "Bad Request.", http.StatusBadRequest)
		return
	}
	// Parsing form and adding instance to datastore.
	version := r.FormValue("version")
	if version == "" {
		c.Errorf("Empty version")
		http.Error(w, "Bad Request.", http.StatusBadRequest)
		return
	}
	k := datastore.NewKey(c, entityName, appID, 0, nil)
	e := &Instance{Version: version}
	if _, err := datastore.Put(c, k, e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Returning current version.
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(&RegisterResponse{Version: currentVersion})
}
