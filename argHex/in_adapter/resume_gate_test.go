package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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
// published cut, the only state the public read has anything to hand out. The
// auth service and the shelf come back so a test can mint an admin token and
// find the live cut's id.
func newPublishedResumeRouter(t *testing.T) (in_port.AuthService, in_port.ResumeService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")
	resumes := publishedShelf()

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

// TestPublishedResumeReadHandsOutNothingButTheUrl pins the cut field by field
// against the raw body rather than a decoded struct, which would silently drop
// whatever it does not name. The keeper's working copy stays off a route
// anybody can call.
func TestPublishedResumeReadHandsOutNothingButTheUrl(t *testing.T) {
	_, _, router := newPublishedResumeRouter(t)

	req := httptest.NewRequest("GET", "/1/resume/published", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	body := rec.Body.String()

	for _, field := range []string{"id", "title", "notes", "filename", "published", "createdAt", "updatedAt"} {
		if strings.Contains(body, `"`+field+`"`) {
			t.Fatalf("the public read handed out %q: %s", field, body)
		}
	}

	if strings.Contains(body, "senior software engineer") {
		t.Fatalf("the public read handed out the keeper's title: %s", body)
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
