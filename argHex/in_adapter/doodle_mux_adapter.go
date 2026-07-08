package in_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type doodleMuxAdapter struct {
	doodle in_port.DoodleService
	auth   *WebAuth
}

// NewDoodleMuxAdapter wires the doodle routes under the mounted /1/doodle
// prefix. List/Get are public; the site's marginalia renders anonymously;
// the mutations are admin-only like the rest of the content editing.
func NewDoodleMuxAdapter(doodle in_port.DoodleService, auth *WebAuth, router *mux.Router) *doodleMuxAdapter {
	a := doodleMuxAdapter{
		doodle: doodle,
		auth:   auth,
	}

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("/{id}", a.Get).Methods("GET")

	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")
	router.HandleFunc("/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	return &a
}

// withShapes pins the contract that shapes is always an array: a shapeless
// doodle is legal, but its nil slice must serialize as [], not null.
func withDoodleShapes(doodle domain.Doodle) domain.Doodle {
	if nil == doodle.Shapes {
		doodle.Shapes = []domain.Shape{}
	}

	return doodle
}

func withDoodleShapesAll(doodles domain.Doodles) domain.Doodles {
	for i := range doodles {
		doodles[i] = withDoodleShapes(doodles[i])
	}

	return doodles
}

func (a doodleMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	doodles, err := a.doodle.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == doodles {
		doodles = domain.Doodles{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, withDoodleShapesAll(doodles))
}

func (a doodleMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	doodle := a.doodle.Get(mux.Vars(r)["id"])

	if "" == doodle.Id {
		writeError(w, 404, "Not found")
		return
	}

	writeJSON(w, http.StatusOK, withDoodleShapes(doodle))
}

func (a doodleMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var doodle domain.Doodle

	if err := json.NewDecoder(r.Body).Decode(&doodle); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.doodle.Create(doodle)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withDoodleShapes(saved))
}

func (a doodleMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var doodle domain.Doodle

	if err := json.NewDecoder(r.Body).Decode(&doodle); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	doodle.Id = mux.Vars(r)["id"]

	saved, err := a.doodle.Update(doodle)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withDoodleShapes(saved))
}

func (a doodleMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.doodle.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}
