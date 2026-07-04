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

type noteMuxAdapter struct {
	note in_port.NoteCRUDService
	auth *WebAuth
}

func NewNoteMuxAdapter(note in_port.NoteCRUDService, auth *WebAuth, router *mux.Router) *noteMuxAdapter {
	a := noteMuxAdapter{
		note: note,
		auth: auth,
	}

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("/{id}", a.Get).Methods("GET")

	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")
	router.HandleFunc("/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	router.HandleFunc("/{id}/publish", a.Publish).Methods("POST")
	router.HandleFunc("/{id}/unpublish", a.Unpublish).Methods("POST")
	router.HandleFunc("/{id}/revisions", a.Revisions).Methods("GET")
	router.HandleFunc("/{id}/revisions/{revisionID}/restore", a.Restore).Methods("POST")

	return &a
}

func (a noteMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	notes, err := a.note.List(queryFlag(r, "published"), queryLimit(r, 0))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Add("X-Total-Count", strconv.Itoa(len(notes)))
	writeJSON(w, http.StatusOK, notes)
}

func (a noteMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.note.Read(mux.Vars(r)["id"]))
}

func (a noteMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	var note domain.Note

	if err := json.NewDecoder(r.Body).Decode(&note); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.note.Create(note)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a noteMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	var note domain.Note

	if err := json.NewDecoder(r.Body).Decode(&note); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	note.Id = mux.Vars(r)["id"]

	saved, err := a.note.Update(note)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a noteMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	if err := a.note.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

func (a noteMuxAdapter) Publish(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	saved, err := a.note.Publish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a noteMuxAdapter) Unpublish(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	saved, err := a.note.Unpublish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a noteMuxAdapter) Revisions(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	revisions, err := a.note.Revisions(mux.Vars(r)["id"], queryLimit(r, 5))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, revisions)
}

func (a noteMuxAdapter) Restore(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	vars := mux.Vars(r)
	saved, err := a.note.Restore(vars["id"], vars["revisionID"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}
