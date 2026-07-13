package in_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

type blockSetMuxAdapter struct {
	blockset in_port.BlockSetService
	auth     *WebAuth
}

// NewBlockSetMuxAdapter wires the blockset routes under /1/blockset. Unlike
// caselogs these have no public read: block sets are an authoring aid, so every
// route is admin-only.
func NewBlockSetMuxAdapter(blockset in_port.BlockSetService, auth *WebAuth, router *mux.Router) *blockSetMuxAdapter {
	a := blockSetMuxAdapter{
		blockset: blockset,
		auth:     auth,
	}

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	return &a
}

// withSetBlocks pins the contract that blocks is always an array: a set's nil
// slice must serialize as [], not null.
func withSetBlocks(set domain.BlockSet) domain.BlockSet {
	if nil == set.Blocks {
		set.Blocks = domain.Blocks{}
	}

	return set
}

func withSetBlocksAll(sets domain.BlockSets) domain.BlockSets {
	for i := range sets {
		sets[i] = withSetBlocks(sets[i])
	}

	return sets
}

func (a blockSetMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	sets, err := a.blockset.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == sets {
		sets = domain.BlockSets{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, withSetBlocksAll(sets))
}

func (a blockSetMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var set domain.BlockSet

	if err := json.NewDecoder(r.Body).Decode(&set); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.blockset.Create(set)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, withSetBlocks(saved))
}

func (a blockSetMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.blockset.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}
