package in_adapter

import (
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type activityMuxAdapter struct {
	activity in_port.ActivityService
}

func NewActivityMuxAdapter(activity in_port.ActivityService, router *mux.Router) *activityMuxAdapter {
	a := activityMuxAdapter{
		activity: activity,
	}

	// the ship's log — admin dashboard data, so it is gated behind auth
	router.HandleFunc("", a.Recent).Methods("GET")
	router.HandleFunc("/", a.Recent).Methods("GET")

	return &a
}

// Recent returns the newest log entries first. Defaults to the last 6 (what the
// dashboard shows).
func (a activityMuxAdapter) Recent(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(w, r) {
		return
	}

	entries, err := a.activity.Recent(queryLimit(r, 6))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, entries)
}
