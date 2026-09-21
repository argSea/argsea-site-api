package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// publishedShelf wires a resume shelf holding one published cut, which is what
// the hoist guard asks for before it starts anything. Nothing here uploads, so
// the file half is never reached.
func publishedShelf() in_port.ResumeService {
	repo := out_adapter.NewResumeFakeOutAdapter()
	repo.Add(domain.Resume{Title: "senior software engineer", Published: true, Filename: "kept.pdf", URL: "/media/images/kept.pdf"})

	return service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter("", "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)
}

// newPublishedResumeRouter mounts the resume adapter over a shelf holding one
// published cut, the only state the public read has anything to hand out. It
// builds its own shelf rather than borrowing publishedShelf, whose webstore has
// no save path: the delete route reaches the file half, and an empty save path
// resolves the stored name against the working directory. The pdf is real so a
// delete has something to clear. The auth service and the shelf come back so a
// test can mint an admin token and find the live cut's id.
func newPublishedResumeRouter(t *testing.T) (in_port.AuthService, in_port.ResumeService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "kept.pdf"), []byte("%PDF-fake"), 0600); nil != err {
		t.Fatalf("could not lay down the stored pdf: %v", err)
	}

	repo := out_adapter.NewResumeFakeOutAdapter()
	repo.Add(domain.Resume{Title: "senior software engineer", Published: true, Filename: "kept.pdf", URL: "/media/images/kept.pdf"})

	resumes := service.NewResumeService(
		repo,
		out_adapter.NewMediaWebstoreAdapter(dir+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewResumeMuxAdapter(resumes, webAuth, router.PathPrefix("/1/resume").Subrouter())

	return authService, resumes, router
}

// newResumeRouter mounts the resume adapter over a temp-dir webstore and a fake
// record store, behind the real JWT gate.
func newResumeRouter(t *testing.T) *mux.Router {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	resumeService := service.NewResumeService(
		out_adapter.NewResumeFakeOutAdapter(),
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewResumeMuxAdapter(resumeService, webAuth, router.PathPrefix("/1/resume").Subrouter())

	return router
}

// newResumeRouterWithAuth is newResumeRouter plus the auth service, so a test
// can mint a token carrying a role.
func newResumeRouterWithAuth(t *testing.T) (in_port.AuthService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	resumeService := service.NewResumeService(
		out_adapter.NewResumeFakeOutAdapter(),
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewResumeMuxAdapter(resumeService, webAuth, router.PathPrefix("/1/resume").Subrouter())

	return authService, router
}

// TestResumeWritesAreAdminOnly pins the role, not just the presence of a token.
// A write downgraded to any valid token is a change no anonymous-only test can
// see.
func TestResumeWritesAreAdminOnly(t *testing.T) {
	authService, router := newResumeRouterWithAuth(t)
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	writes := []struct {
		method string
		path   string
	}{
		{"POST", "/1/resume/"},
		{"PUT", "/1/resume/some-id"},
		{"POST", "/1/resume/some-id/publish"},
		{"POST", "/1/resume/some-id/unpublish"},
		{"DELETE", "/1/resume/some-id"},
	}

	for _, c := range writes {
		req := httptest.NewRequest(c.method, c.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if http.StatusForbidden != rec.Code {
			t.Fatalf("expected 403 for a plain-user %s %s, got %d", c.method, c.path, rec.Code)
		}
	}
}

func TestResumeRoutesAreAuthGated(t *testing.T) {
	router := newResumeRouter(t)

	// every route but the published read is keeper-only: the notes on the shelf
	// are his working copy
	cases := []struct {
		method string
		path   string
	}{
		{"GET", "/1/resume"},
		{"GET", "/1/resume/"},
		{"GET", "/1/resume/some-id"},
		{"POST", "/1/resume/"},
		{"PUT", "/1/resume/some-id"},
		{"POST", "/1/resume/some-id/publish"},
		{"POST", "/1/resume/some-id/unpublish"},
		{"DELETE", "/1/resume/some-id"},
	}

	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if http.StatusUnauthorized != rec.Code {
			t.Fatalf("expected 401 for anonymous %s %s, got %d", c.method, c.path, rec.Code)
		}
	}
}

// TestPublishedResumeReadIsPublic pins the one anonymous route on the shelf,
// both spellings. The site build sends no auth header, so a gate creeping onto
// this route breaks the build rather than a page.
func TestPublishedResumeReadIsPublic(t *testing.T) {
	_, _, router := newPublishedResumeRouter(t)

	for _, path := range []string{"/1/resume/published", "/1/resume/published/"} {
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if http.StatusOK != rec.Code {
			t.Fatalf("%s: expected 200 for the anonymous published read, got %d: %s", path, rec.Code, rec.Body.String())
		}

		var published struct {
			URL string `json:"url"`
		}

		json.Unmarshal(rec.Body.Bytes(), &published)

		if "/media/images/kept.pdf" != published.URL {
			t.Fatalf("%s: expected the live cut's url, got %q", path, published.URL)
		}
	}
}

// TestPublishedResumeReadHandsOutNothingButTheUrl derives the cut rather than
// listing what to block: the body must carry exactly one key. A list of json
// names holds only against the fields whoever wrote it thought of, and the
// shelf's own filename rides out past it under any name not on the list.
func TestPublishedResumeReadHandsOutNothingButTheUrl(t *testing.T) {
	_, _, router := newPublishedResumeRouter(t)

	req := httptest.NewRequest("GET", "/1/resume/published", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body map[string]interface{}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); nil != err {
		t.Fatalf("the public read must answer with an object, got %s: %v", rec.Body.String(), err)
	}

	if 1 != len(body) {
		t.Fatalf("expected the url and nothing else, got %v", body)
	}

	if _, carried := body["url"]; !carried {
		t.Fatalf("expected the one key to be url, got %v", body)
	}
}

// TestPublishedResumeReadIs404WithNothingOnTheShelf also pins the route order:
// /published registered after /{id} would be swallowed by it, and the authed
// Get handler would answer this anonymous call with a 401 instead.
func TestPublishedResumeReadIs404WithNothingOnTheShelf(t *testing.T) {
	router := newResumeRouter(t)

	req := httptest.NewRequest("GET", "/1/resume/published", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusNotFound != rec.Code {
		t.Fatalf("expected 404 with nothing published, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestDeletingThePublishedResumeIs409 pins the refusal reaching the admin as a
// conflict rather than as the 500 everything unrecognised maps to.
func TestDeletingThePublishedResumeIs409(t *testing.T) {
	authService, resumes, router := newPublishedResumeRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	listed, _ := resumes.List()

	if 1 != len(listed) {
		t.Fatalf("expected the one published cut on the shelf, got %+v", listed)
	}

	req := httptest.NewRequest("DELETE", "/1/resume/"+listed[0].Id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 deleting the published cut, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUnpublishThenDeleteTheOnlyCut walks the whole route the delete refusal
// points the keeper down, over a shelf holding exactly one cut: the case that
// had no way out before unpublish existed. Both spellings of the path are
// exercised, the first for real and the second against the cut already down.
func TestUnpublishThenDeleteTheOnlyCut(t *testing.T) {
	authService, resumes, router := newPublishedResumeRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	listed, _ := resumes.List()
	id := listed[0].Id

	if rec := resumeRequest(t, router, "DELETE", "/1/resume/"+id, token); http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 while the cut is up, got %d", rec.Code)
	}

	for _, path := range []string{"/1/resume/" + id + "/unpublish", "/1/resume/" + id + "/unpublish/"} {
		rec := resumeRequest(t, router, "POST", path, token)

		if http.StatusOK != rec.Code {
			t.Fatalf("%s: expected 200 for the admin unpublish, got %d: %s", path, rec.Code, rec.Body.String())
		}

		var saved domain.Resume
		json.Unmarshal(rec.Body.Bytes(), &saved)

		// the id, not just the flag: a handler reaching for the wrong cut hands
		// back something equally down and nothing else tells them apart
		if saved.Published || id != saved.Id {
			t.Fatalf("%s: expected the named cut handed back down, got %+v", path, saved)
		}
	}

	// the public read has nothing to hand out once the hoist is empty
	if rec := resumeRequest(t, router, "GET", "/1/resume/published", ""); http.StatusNotFound != rec.Code {
		t.Fatalf("expected 404 from the public read with the hoist empty, got %d", rec.Code)
	}

	if rec := resumeRequest(t, router, "DELETE", "/1/resume/"+id, token); http.StatusOK != rec.Code {
		t.Fatalf("expected 200 deleting the cut once it is down, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUnpublishAnUnknownIdIs400 pins the unknown id landing on the shelf's
// validation branch rather than the 500 everything unrecognised maps to.
func TestUnpublishAnUnknownIdIs400(t *testing.T) {
	authService, _, router := newPublishedResumeRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	if rec := resumeRequest(t, router, "POST", "/1/resume/nope/unpublish", token); http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 unpublishing an unknown id, got %d: %s", rec.Code, rec.Body.String())
	}
}

// resumeRequest fires one call at the shelf, with a bearer token when given.
func resumeRequest(t *testing.T, router *mux.Router, method string, path string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}
