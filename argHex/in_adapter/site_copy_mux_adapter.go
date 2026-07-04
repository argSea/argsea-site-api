package in_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type siteCopyMuxAdapter struct {
	copy in_port.SiteCopyService
	auth *WebAuth
}

func NewSiteCopyMuxAdapter(copy in_port.SiteCopyService, auth *WebAuth, router *mux.Router) *siteCopyMuxAdapter {
	a := siteCopyMuxAdapter{
		copy: copy,
		auth: auth,
	}

	// singleton: GET reads it (public), PUT upserts it (authored)
	router.HandleFunc("", a.Get).Methods("GET")
	router.HandleFunc("/", a.Get).Methods("GET")
	router.HandleFunc("", a.Save).Methods("PUT")
	router.HandleFunc("/", a.Save).Methods("PUT")

	return &a
}

func (a siteCopyMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.copy.Get())
}

func (a siteCopyMuxAdapter) Save(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	var copy domain.SiteCopy

	if err := json.NewDecoder(r.Body).Decode(&copy); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.copy.Save(copy)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}
