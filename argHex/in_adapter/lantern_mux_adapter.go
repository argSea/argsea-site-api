package in_adapter

import (
	"errors"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type lanternMuxAdapter struct {
	lantern in_port.LanternService
	auth    *WebAuth
}

// NewLanternMuxAdapter wires the deploy routes. Both are admin-only: a valid
// token without the admin role gets a 403, no token a 401. main.go mounts this
// adapter only when the config has a lantern section.
func NewLanternMuxAdapter(lantern in_port.LanternService, auth *WebAuth, router *mux.Router) *lanternMuxAdapter {
	a := lanternMuxAdapter{
		lantern: lantern,
		auth:    auth,
	}

	router.HandleFunc("", a.Status).Methods("GET")
	router.HandleFunc("/", a.Status).Methods("GET")

	router.HandleFunc("/hoist", a.Hoist).Methods("POST")
	router.HandleFunc("/hoist/", a.Hoist).Methods("POST")

	router.HandleFunc("/rollback", a.Rollback).Methods("POST")
	router.HandleFunc("/rollback/", a.Rollback).Methods("POST")

	return &a
}

// Status is what the admin polls while the boat is out.
func (a lanternMuxAdapter) Status(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	writeJSON(w, http.StatusOK, a.lantern.Status())
}

// Hoist starts a deploy: 202 with the fresh status, or 409 with the current
// one when a hoist is already in flight.
func (a lanternMuxAdapter) Hoist(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	status, err := a.lantern.Hoist()

	if errors.Is(err, in_port.ErrHoistAlreadyRunning) {
		writeJSON(w, http.StatusConflict, status)
		return
	}

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, status)
}

// Rollback re-points the live link at the previous kept build — no rebuild.
// 200 with the status on success; 409 while a hoist is in flight (carrying the
// running status, like Hoist) or when nothing older is kept.
func (a lanternMuxAdapter) Rollback(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	status, err := a.lantern.Rollback()

	if errors.Is(err, in_port.ErrHoistAlreadyRunning) {
		writeJSON(w, http.StatusConflict, status)
		return
	}

	if errors.Is(err, in_port.ErrNoPreviousBuild) {
		writeError(w, 409, err.Error())
		return
	}

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, status)
}
