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

type carvingMuxAdapter struct {
	carving in_port.CarvingService
	auth    *WebAuth
}

// NewCarvingMuxAdapter wires the carving shop routes. List is public: the
// site build reads the catalog anonymously; every mutation is admin-only
// like the rest of the content editing.
func NewCarvingMuxAdapter(carving in_port.CarvingService, auth *WebAuth, router *mux.Router) *carvingMuxAdapter {
	a := carvingMuxAdapter{
		carving: carving,
		auth:    auth,
	}

	router.HandleFunc("/carvings", a.List).Methods("GET")
	router.HandleFunc("/carvings/", a.List).Methods("GET")
	router.HandleFunc("/carvings", a.Create).Methods("POST")
	router.HandleFunc("/carvings/", a.Create).Methods("POST")
	router.HandleFunc("/carvings/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/carvings/{id}", a.Delete).Methods("DELETE")

	router.HandleFunc("/carvings/{id}/bolt", a.Bolt).Methods("POST")

	return &a
}

// withBoltedTo pins the contract that boltedTo is always an array: a fresh
// carving is legal with nothing bolted, but its nil slice must serialize as
// [], not null, the same rule the figurehead wardrobe's shapes already
// follow.
func withBoltedTo(carving domain.Carving) domain.Carving {
	if nil == carving.BoltedTo {
		carving.BoltedTo = []string{}
	}

	return carving
}

func withBoltedToAll(carvings domain.Carvings) domain.Carvings {
	for i := range carvings {
		carvings[i] = withBoltedTo(carvings[i])
	}

	return carvings
}

// List hands out the whole catalog; no auth, this is what the site builds
// against.
func (a carvingMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	carvings, err := a.carving.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == carvings {
		carvings = domain.Carvings{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, withBoltedToAll(carvings))
}

func (a carvingMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var carving domain.Carving

	if err := json.NewDecoder(r.Body).Decode(&carving); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.carving.Create(carving)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBoltedTo(saved))
}

func (a carvingMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var carving domain.Carving

	if err := json.NewDecoder(r.Body).Decode(&carving); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	carving.Id = mux.Vars(r)["id"]

	saved, err := a.carving.Update(carving)

	if errors.Is(err, in_port.ErrCarvingBuiltin) {
		writeError(w, 409, err.Error())
		return
	}

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBoltedTo(saved))
}

// Delete refuses a builtin outright with a 409: the seven v1 carvings are
// permanent so every spot always has a v1 to bolt back to.
func (a carvingMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	err := a.carving.Delete(mux.Vars(r)["id"])

	if errors.Is(err, in_port.ErrCarvingBuiltin) {
		writeError(w, 409, err.Error())
		return
	}

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

// Bolt moves a spot onto this carving, stripping it from whoever held it
// before; the body carries the spot the same way a project reorder carries
// its new position.
func (a carvingMuxAdapter) Bolt(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var body struct {
		Spot string `json:"spot"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.carving.Bolt(mux.Vars(r)["id"], body.Spot)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withBoltedTo(saved))
}
