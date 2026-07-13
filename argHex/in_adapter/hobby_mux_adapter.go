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

type hobbyMuxAdapter struct {
	hobby in_port.HobbyCRUDService
	auth  *WebAuth
}

func NewHobbyMuxAdapter(hobby in_port.HobbyCRUDService, auth *WebAuth, router *mux.Router) *hobbyMuxAdapter {
	a := hobbyMuxAdapter{
		hobby: hobby,
		auth:  auth,
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

// List returns the ship's log. ?active=true narrows to the moored ships, the
// ones at their berth; without it, every ship in the log comes back.
func (a hobbyMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	hobbies, err := a.hobby.List(queryFlag(r, "active"))

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == hobbies {
		hobbies = domain.Hobbies{} // empty list must serialize as [], not null
	}

	w.Header().Add("X-Total-Count", strconv.Itoa(len(hobbies)))
	writeJSON(w, http.StatusOK, hobbies)
}

func (a hobbyMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.hobby.Read(mux.Vars(r)["id"]))
}

func (a hobbyMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var hobby domain.Hobby

	if err := json.NewDecoder(r.Body).Decode(&hobby); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.hobby.Create(hobby)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a hobbyMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var hobby domain.Hobby

	if err := json.NewDecoder(r.Body).Decode(&hobby); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	hobby.Id = mux.Vars(r)["id"]

	saved, err := a.hobby.Update(hobby)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a hobbyMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.hobby.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}
