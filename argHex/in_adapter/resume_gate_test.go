package in_adapter_test

import (
	"net/http"
	"net/http/httptest"
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

	// the whole shelf is keeper-only: the notes on it are his working copy
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
