package in_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type watchMuxAdapter struct {
	watch in_port.WatchService
	auth  *WebAuth
}

func NewWatchMuxAdapter(watch in_port.WatchService, auth *WebAuth, router *mux.Router) *watchMuxAdapter {
	a := watchMuxAdapter{
		watch: watch,
		auth:  auth,
	}

	// singleton: GET reads it (public), PUT upserts it (authed)
	router.HandleFunc("", a.Get).Methods("GET")
	router.HandleFunc("/", a.Get).Methods("GET")
	router.HandleFunc("", a.Save).Methods("PUT")
	router.HandleFunc("/", a.Save).Methods("PUT")

	return &a
}

func (a watchMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.watch.Get())
}

func (a watchMuxAdapter) Save(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var watch domain.Watch

	if err := json.NewDecoder(r.Body).Decode(&watch); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.watch.Save(watch)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}
