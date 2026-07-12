package in_adapter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

// sightingBodyCap bounds an ingest body. A real beacon is a few dozen bytes;
// anything past this is junk or an attempt to bloat the ledger.
const sightingBodyCap = 2 << 10

// sightingDefaultDays is the aggregate window when the caller names none. The
// service clamps whatever it is handed to the allowed band.
const sightingDefaultDays = 7

type sightingMuxAdapter struct {
	sighting in_port.SightingService
	auth     *WebAuth
}

// NewSightingMuxAdapter wires the harbor's tally. Ingest is public: the shore
// pings it with navigator.sendBeacon, which sends a bare text/plain body, so it
// must take any content type and never require a token. Traffic is the watch
// room's aggregate read, gated like the keeper's log.
func NewSightingMuxAdapter(sighting in_port.SightingService, auth *WebAuth, router *mux.Router) *sightingMuxAdapter {
	a := sightingMuxAdapter{
		sighting: sighting,
		auth:     auth,
	}

	router.HandleFunc("", a.Ingest).Methods("POST")
	router.HandleFunc("/", a.Ingest).Methods("POST")

	router.HandleFunc("/traffic", a.Traffic).Methods("GET")
	router.HandleFunc("/traffic/", a.Traffic).Methods("GET")

	return &a
}

// Ingest records one anonymous ping. The body is JSON but parsed regardless of
// Content-Type, since sendBeacon sends text/plain to stay a simple request. A
// stored ping and a dropped bot both answer 204 with no body: fast, and the
// endpoint never echoes the ledger back to the shore.
func (a sightingMuxAdapter) Ingest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, sightingBodyCap+1))

	if nil != err {
		writeError(w, 400, "unreadable body")
		return
	}

	if len(body) > sightingBodyCap {
		writeError(w, 400, "sighting too large")
		return
	}

	var beacon domain.SightingBeacon

	if err := json.Unmarshal(body, &beacon); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	err = a.sighting.Record(beacon, clientIP(r), r.UserAgent())

	if errors.Is(err, in_port.ErrSightingRejected) {
		writeError(w, 400, err.Error())
		return
	}

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Traffic hands the watch room the aggregate over the requested window.
func (a sightingMuxAdapter) Traffic(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	report, err := a.sighting.Traffic(trafficDays(r))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// clientIP is the shore's address as nginx forwards it: the first hop of
// X-Forwarded-For, then X-Real-IP, falling back to the raw connection. Only the
// visitor hash ever sees it, and never leaves the API.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); "" != forwarded {
		if comma := strings.Index(forwarded, ","); comma >= 0 {
			return strings.TrimSpace(forwarded[:comma])
		}

		return strings.TrimSpace(forwarded)
	}

	if real := r.Header.Get("X-Real-IP"); "" != real {
		return real
	}

	return r.RemoteAddr
}

// trafficDays reads an optional ?days= integer, defaulting when it is absent or
// unparseable. The service clamps the value to the allowed band.
func trafficDays(r *http.Request) int {
	raw := r.URL.Query().Get("days")

	if "" == raw {
		return sightingDefaultDays
	}

	days, err := strconv.Atoi(raw)

	if nil != err {
		return sightingDefaultDays
	}

	return days
}
