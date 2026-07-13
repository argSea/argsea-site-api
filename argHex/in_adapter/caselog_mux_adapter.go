package in_adapter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type caseLogMuxAdapter struct {
	caselog in_port.CaseLogCRUDService
	auth    *WebAuth
}

// NewCaseLogMuxAdapter wires the caselog routes under /1/caselog, mirroring the
// project split: public reads see published logs only, every write is admin.
func NewCaseLogMuxAdapter(caselog in_port.CaseLogCRUDService, auth *WebAuth, router *mux.Router) *caseLogMuxAdapter {
	a := caseLogMuxAdapter{
		caselog: caselog,
		auth:    auth,
	}

	// public reads (Astro build consumes ?published=true)
	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("/{id}", a.Get).Methods("GET")

	// authored writes
	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")
	router.HandleFunc("/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	// lifecycle + history
	router.HandleFunc("/{id}/publish", a.Publish).Methods("POST")
	router.HandleFunc("/{id}/unpublish", a.Unpublish).Methods("POST")
	router.HandleFunc("/{id}/revisions", a.Revisions).Methods("GET")
	router.HandleFunc("/{id}/revisions/{revisionID}/restore", a.Restore).Methods("POST")

	return &a
}

// withBlocks pins the contract that blocks is always an array: a blockless log
// is legal, but its nil slice must serialize as [], not null.
func withBlocks(log domain.CaseLog) domain.CaseLog {
	if nil == log.Blocks {
		log.Blocks = domain.Blocks{}
	}

	return log
}

func withBlocksAll(logs domain.CaseLogs) domain.CaseLogs {
	for i := range logs {
		logs[i] = withBlocks(logs[i])
	}

	return logs
}

func (a caseLogMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	// drafts are for the keeper only: unauthenticated readers always get the
	// published-only view, whatever the query says
	publishedOnly := queryFlag(r, "published") || !a.auth.Authorized(r)

	logs, err := a.caselog.List(publishedOnly, queryLimit(r, 0))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == logs {
		logs = domain.CaseLogs{} // empty list must serialize as [], not null
	}

	w.Header().Add("X-Total-Count", strconv.Itoa(len(logs)))
	writeJSON(w, http.StatusOK, withBlocksAll(logs))
}

func (a caseLogMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	log := a.caselog.Read(mux.Vars(r)["id"])

	// unauthenticated readers only see published documents; a 404 (not 401)
	// avoids confirming that a draft exists
	if domain.StatusPublished != log.Status && !a.auth.Authorized(r) {
		writeError(w, 404, "Not found")
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(log))
}

func (a caseLogMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var log domain.CaseLog

	if err := json.NewDecoder(r.Body).Decode(&log); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.caselog.Create(log)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(saved))
}

func (a caseLogMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var log domain.CaseLog

	if err := json.NewDecoder(r.Body).Decode(&log); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	log.Id = mux.Vars(r)["id"]

	saved, err := a.caselog.Update(log)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(saved))
}

func (a caseLogMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.caselog.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

func (a caseLogMuxAdapter) Publish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.caselog.Publish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(saved))
}

func (a caseLogMuxAdapter) Unpublish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.caselog.Unpublish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(saved))
}

// Revisions lists the last few printings for the rollback UI. Admin-only; the
// history is not public. Defaults to the last 5.
func (a caseLogMuxAdapter) Revisions(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	revisions, err := a.caselog.Revisions(mux.Vars(r)["id"], queryLimit(r, 5))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == revisions {
		revisions = domain.Revisions{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, revisions)
}

func (a caseLogMuxAdapter) Restore(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	vars := mux.Vars(r)
	saved, err := a.caselog.Restore(vars["id"], vars["revisionID"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBlocks(saved))
}
