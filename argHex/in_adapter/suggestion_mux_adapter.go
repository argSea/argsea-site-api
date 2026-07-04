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

type suggestionMuxAdapter struct {
	suggestion in_port.SuggestionService
	auth       *WebAuth
}

func NewSuggestionMuxAdapter(suggestion in_port.SuggestionService, auth *WebAuth, router *mux.Router) *suggestionMuxAdapter {
	a := suggestionMuxAdapter{
		suggestion: suggestion,
		auth:       auth,
	}

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("", a.Add).Methods("POST")
	router.HandleFunc("/", a.Add).Methods("POST")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	return &a
}

func (a suggestionMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	suggestions, err := a.suggestion.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == suggestions {
		suggestions = domain.Suggestions{} // empty list must serialize as [], not null
	}

	w.Header().Add("X-Total-Count", strconv.Itoa(len(suggestions)))
	writeJSON(w, http.StatusOK, suggestions)
}

// Add appends a chip. The body carries just the value: {"value": "kayaking?"}.
func (a suggestionMuxAdapter) Add(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	var body struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.suggestion.Add(body.Value)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a suggestionMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	if err := a.suggestion.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}
