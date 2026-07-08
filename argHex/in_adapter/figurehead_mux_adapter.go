package in_adapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type figureheadMuxAdapter struct {
	figurehead in_port.FigureheadService
	auth       *WebAuth
}

// NewFigureheadMuxAdapter wires the Figurehead Shop routes. The published read
// is public: the site build consumes it anonymously; everything else is
// admin-only like the rest of the content mutations.
func NewFigureheadMuxAdapter(figurehead in_port.FigureheadService, auth *WebAuth, router *mux.Router) *figureheadMuxAdapter {
	a := figureheadMuxAdapter{
		figurehead: figurehead,
		auth:       auth,
	}

	router.HandleFunc("/published", a.Published).Methods("GET")
	router.HandleFunc("/published/", a.Published).Methods("GET")

	router.HandleFunc("/designs", a.List).Methods("GET")
	router.HandleFunc("/designs/", a.List).Methods("GET")
	router.HandleFunc("/designs", a.Create).Methods("POST")
	router.HandleFunc("/designs/", a.Create).Methods("POST")
	router.HandleFunc("/designs/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/designs/{id}", a.Delete).Methods("DELETE")

	router.HandleFunc("/designs/{id}/publish", a.Publish).Methods("POST")

	return &a
}

// withShapes pins the contract that shapes is always an array: a shapeless
// draft is legal, but its nil slice must serialize as [], not null; the same
// rule the design lists already follow.
func withShapes(design domain.CatDesign) domain.CatDesign {
	if nil == design.Shapes {
		design.Shapes = []domain.Shape{}
	}

	return design
}

func withShapesAll(designs domain.CatDesigns) domain.CatDesigns {
	for i := range designs {
		designs[i] = withShapes(designs[i])
	}

	return designs
}

// Published hands out the design on the bow for each pose; no auth, this is
// what the site builds against.
func (a figureheadMuxAdapter) Published(w http.ResponseWriter, r *http.Request) {
	designs, err := a.figurehead.Published()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == designs {
		designs = domain.CatDesigns{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, withShapesAll(designs))
}

func (a figureheadMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	designs, err := a.figurehead.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == designs {
		designs = domain.CatDesigns{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, withShapesAll(designs))
}

func (a figureheadMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var design domain.CatDesign

	if err := json.NewDecoder(r.Body).Decode(&design); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.figurehead.Create(design)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withShapes(saved))
}

func (a figureheadMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var design domain.CatDesign

	if err := json.NewDecoder(r.Body).Decode(&design); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	design.Id = mux.Vars(r)["id"]

	saved, err := a.figurehead.Update(design)

	if errors.Is(err, in_port.ErrDesignSeeded) {
		writeError(w, 409, err.Error())
		return
	}

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withShapes(saved))
}

// Delete refuses the undeletable with a 409: published designs and the seeded
// v1s are superseded through Publish, never removed.
func (a figureheadMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	err := a.figurehead.Delete(mux.Vars(r)["id"])

	if errors.Is(err, in_port.ErrDesignSeeded) || errors.Is(err, in_port.ErrDesignPublished) {
		writeError(w, 409, err.Error())
		return
	}

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

func (a figureheadMuxAdapter) Publish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.figurehead.Publish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withShapes(saved))
}
