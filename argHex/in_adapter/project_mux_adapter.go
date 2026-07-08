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

type projectMuxAdapter struct {
	project in_port.ProjectCRUDService
	auth    *WebAuth
}

func NewProjectMuxAdapter(project in_port.ProjectCRUDService, auth *WebAuth, router *mux.Router) *projectMuxAdapter {
	a := projectMuxAdapter{
		project: project,
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
	router.HandleFunc("/{id}/reorder", a.Reorder).Methods("POST")
	router.HandleFunc("/{id}/feature", a.Feature).Methods("POST")
	router.HandleFunc("/{id}/unfeature", a.Unfeature).Methods("POST")
	router.HandleFunc("/{id}/revisions", a.Revisions).Methods("GET")
	router.HandleFunc("/{id}/revisions/{revisionID}/restore", a.Restore).Methods("POST")

	return &a
}

func (a projectMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	// drafts are for the keeper only: unauthenticated readers always get the
	// published-only view, whatever the query says
	publishedOnly := queryFlag(r, "published") || !a.auth.Authorized(r)

	projects, err := a.project.List(publishedOnly, queryLimit(r, 0))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == projects {
		projects = domain.Projects{} // empty list must serialize as [], not null
	}

	w.Header().Add("X-Total-Count", strconv.Itoa(len(projects)))
	writeJSON(w, http.StatusOK, projects)
}

func (a projectMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	project := a.project.Read(mux.Vars(r)["id"])

	// unauthenticated readers only see published documents; a 404 (not 401)
	// avoids confirming that a draft exists
	if domain.StatusPublished != project.Status && !a.auth.Authorized(r) {
		writeError(w, 404, "Not found")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (a projectMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var project domain.Project

	if err := json.NewDecoder(r.Body).Decode(&project); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.project.Create(project)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a projectMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var project domain.Project

	if err := json.NewDecoder(r.Body).Decode(&project); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	project.Id = mux.Vars(r)["id"]

	saved, err := a.project.Update(project)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a projectMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.project.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

func (a projectMuxAdapter) Publish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.project.Publish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a projectMuxAdapter) Unpublish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.project.Unpublish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// Reorder moves the postcard to a new rack position; lifecycle-style, so no
// revision snapshot behind it.
func (a projectMuxAdapter) Reorder(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var body struct {
		Order *int `json:"order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	if nil == body.Order {
		writeError(w, 400, "order is required")
		return
	}

	saved, err := a.project.Reorder(mux.Vars(r)["id"], *body.Order)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a projectMuxAdapter) Feature(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.project.Feature(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a projectMuxAdapter) Unfeature(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.project.Unfeature(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// Revisions lists the last few printings for the rollback UI. Admin-only; the
// history is not public. Defaults to the last 5.
func (a projectMuxAdapter) Revisions(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	revisions, err := a.project.Revisions(mux.Vars(r)["id"], queryLimit(r, 5))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == revisions {
		revisions = domain.Revisions{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, revisions)
}

func (a projectMuxAdapter) Restore(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	vars := mux.Vars(r)
	saved, err := a.project.Restore(vars["id"], vars["revisionID"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}
