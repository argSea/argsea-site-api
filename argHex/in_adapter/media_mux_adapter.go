package in_adapter

import (
	"errors"
	"io"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

// mediaUploadMaxBytes caps a multipart upload; photographs, not archives.
const mediaUploadMaxBytes = 32 << 20

type mediaMuxAdapter struct {
	media in_port.MediaService
	auth  *WebAuth
}

// NewMediaMuxAdapter wires the darkroom routes. Every route is authed; the
// site consumes media files straight from disk, never through this API.
func NewMediaMuxAdapter(media in_port.MediaService, auth *WebAuth, router *mux.Router) *mediaMuxAdapter {
	a := mediaMuxAdapter{
		media: media,
		auth:  auth,
	}

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")

	return &a
}

func (a mediaMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	media, err := a.media.ListMedia()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == media {
		media = domain.MediaList{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, media)
}

// Create develops a multipart upload from the "file" field: the part's
// filename and content type travel to the service, which owns the image-only
// gate.
func (a mediaMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, mediaUploadMaxBytes)

	file, header, err := r.FormFile("file")

	if nil != err {
		writeError(w, 400, "multipart upload with a \"file\" field is required")
		return
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)

	if nil != err {
		writeError(w, 400, err.Error())
		return
	}

	saved, err := a.media.CreateMedia(header.Filename, header.Header.Get("Content-Type"), bytes)

	if nil != err {
		writeError(w, mediaErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a mediaMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.media.DeleteMedia(mux.Vars(r)["id"]); nil != err {
		writeError(w, mediaErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

// mediaErrorCode maps a service error onto its status: 400 when the request
// itself was rejected, 500 when the infrastructure (disk, mongo) failed.
func mediaErrorCode(err error) int64 {
	var validation in_port.MediaValidationError

	if errors.As(err, &validation) {
		return 400
	}

	return 500
}
