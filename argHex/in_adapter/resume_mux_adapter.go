package in_adapter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

// resumeUploadMaxBytes caps a multipart upload; a resume is a few pages of PDF,
// not an archive.
const resumeUploadMaxBytes = 32 << 20

type resumeMuxAdapter struct {
	resume in_port.ResumeService
	auth   *WebAuth
}

// NewResumeMuxAdapter wires the resume shelf's routes. The published read is
// public: the site build consumes it anonymously, the same way it consumes the
// figurehead shop's. Every other route is authed, because the titles and notes
// on the shelf are the keeper's own working copy, and the PDFs themselves are
// served static off disk, never through this API.
func NewResumeMuxAdapter(resume in_port.ResumeService, auth *WebAuth, router *mux.Router) *resumeMuxAdapter {
	a := resumeMuxAdapter{
		resume: resume,
		auth:   auth,
	}

	// ahead of /{id}: mux takes the first route that matches, so a literal
	// registered after the id pattern would never be reached
	router.HandleFunc("/published", a.Published).Methods("GET")
	router.HandleFunc("/published/", a.Published).Methods("GET")

	router.HandleFunc("", a.List).Methods("GET")
	router.HandleFunc("/", a.List).Methods("GET")
	router.HandleFunc("", a.Create).Methods("POST")
	router.HandleFunc("/", a.Create).Methods("POST")

	router.HandleFunc("/{id}", a.Get).Methods("GET")
	router.HandleFunc("/{id}", a.Update).Methods("PUT")
	router.HandleFunc("/{id}", a.Delete).Methods("DELETE")
	router.HandleFunc("/{id}/publish", a.Publish).Methods("POST")
	router.HandleFunc("/{id}/publish/", a.Publish).Methods("POST")
	router.HandleFunc("/{id}/unpublish", a.Unpublish).Methods("POST")
	router.HandleFunc("/{id}/unpublish/", a.Unpublish).Methods("POST")

	return &a
}

// publicResume is the published read's wire shape: the url of the live PDF and
// nothing else. Everything else a Resume carries is either the keeper's own
// working copy (title, notes) or shelf bookkeeping (id, filename, stamps), and
// published itself is a constant on a route that serves nothing else.
type publicResume struct {
	URL string `json:"url"`
}

// Published hands out the one live cut; no auth, this is what the site builds
// its resume link against. Nothing on the shelf is a 404 rather than an empty
// object, so a consumer cannot mistake "none published" for a blank url.
func (a resumeMuxAdapter) Published(w http.ResponseWriter, r *http.Request) {
	resume, err := a.resume.Published()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if "" == resume.Id {
		writeError(w, 404, "no resume is published")
		return
	}

	writeJSON(w, http.StatusOK, publicResume{URL: resume.URL})
}

func (a resumeMuxAdapter) List(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	resumes, err := a.resume.List()

	if nil != err {
		writeError(w, 500, err.Error())
		return
	}

	if nil == resumes {
		resumes = domain.Resumes{} // empty list must serialize as [], not null
	}

	writeJSON(w, http.StatusOK, resumes)
}

// Get hands one stored cut back so the keeper can open it for comparison.
func (a resumeMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(a.auth, w, r) {
		return
	}

	writeJSON(w, http.StatusOK, a.resume.Read(mux.Vars(r)["id"]))
}

// Create stores a multipart upload from the "file" field, with the title and
// notes riding alongside it as form values. The part's content type travels to
// the service, which owns the pdf-only gate.
func (a resumeMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, resumeUploadMaxBytes)

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

	resume := domain.Resume{
		Title: r.FormValue("title"),
		Notes: r.FormValue("notes"),
	}

	saved, err := a.resume.Create(resume, header.Header.Get("Content-Type"), bytes)

	if nil != err {
		writeError(w, resumeErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a resumeMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	var resume domain.Resume

	if err := json.NewDecoder(r.Body).Decode(&resume); nil != err {
		writeError(w, 400, err.Error())
		return
	}

	resume.Id = mux.Vars(r)["id"]

	saved, err := a.resume.Update(resume)

	if nil != err {
		writeError(w, resumeErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// Publish makes one cut the live one; whichever held it is cleared by the same
// call, so the shelf never shows two.
func (a resumeMuxAdapter) Publish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.resume.Publish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, resumeErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// Unpublish takes the live cut down and leaves the hoist empty, which is what a
// keeper who means to delete it does first.
func (a resumeMuxAdapter) Unpublish(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	saved, err := a.resume.Unpublish(mux.Vars(r)["id"])

	if nil != err {
		writeError(w, resumeErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

func (a resumeMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(a.auth, w, r) {
		return
	}

	if err := a.resume.Delete(mux.Vars(r)["id"]); nil != err {
		writeError(w, resumeErrorCode(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data_objects.ItemLessResponseObject{Status: "ok", Code: 200})
}

// resumeErrorCode maps a service error onto its status: 409 when the shelf
// refuses a transition it will not make, 400 when the request itself was
// rejected, 500 when the infrastructure (disk, mongo) failed. Either way the
// message is the error's own, so a permission failure on the media directory
// reaches the admin as one.
func resumeErrorCode(err error) int64 {
	if errors.Is(err, in_port.ErrResumePublished) {
		return 409
	}

	var validation in_port.ResumeValidationError

	if errors.As(err, &validation) {
		return 400
	}

	return 500
}
