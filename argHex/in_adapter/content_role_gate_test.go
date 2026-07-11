package in_adapter_test

import (
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

// newContentGateRouter mounts the content adapters whose mutations now demand
// the admin role, behind the real JWT gate, and hands back the auth service so
// a test can mint both an admin and a plain-user token. It seeds one published
// project so the public read path has something to return.
func newContentGateRouter(t *testing.T) (in_port.AuthService, string, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	noteRepo := out_adapter.NewNoteFakeOutAdapter()
	projects := service.NewProjectCRUDService(out_adapter.NewProjectFakeOutAdapter(), noteRepo, revisions, activity)
	notes := service.NewNoteCRUDService(noteRepo, revisions, activity)
	media := service.NewMediaService(
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images"),
		out_adapter.NewMediaMetaFakeOutAdapter(),
		activity,
	)
	users := service.NewUserCRUDService(out_adapter.NewUserFakeOutAdapter())

	published, err := projects.Create(domain.Project{Title: "Published card"})

	if nil != err {
		t.Fatalf("seed published failed: %v", err)
	}

	if _, err := projects.Publish(published.Id); nil != err {
		t.Fatalf("seed publish failed: %v", err)
	}

	router := mux.NewRouter()
	in_adapter.NewProjectMuxAdapter(projects, webAuth, router.PathPrefix("/1/project").Subrouter())
	in_adapter.NewNoteMuxAdapter(notes, webAuth, router.PathPrefix("/1/note").Subrouter())
	in_adapter.NewMediaMuxAdapter(media, webAuth, router.PathPrefix("/1/media").Subrouter())
	in_adapter.NewUserMuxAdapter(users, media, webAuth, router.PathPrefix("/1/user").Subrouter())

	return authService, published.Id, router
}

func gateRequest(t *testing.T, router *mux.Router, method string, path string, body string, token string) int {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec.Code
}

// contentMutations is a representative write on each content adapter whose gate
// moved from any-valid-token to admin-only. The gate runs before any body
// parsing, so a bare path is enough to prove the bounce.
var contentMutations = []struct {
	name   string
	method string
	path   string
	body   string
}{
	{"project create", "POST", "/1/project", `{"title":"x"}`},
	{"project delete", "DELETE", "/1/project/whatever", ""},
	{"project publish", "POST", "/1/project/whatever/publish", ""},
	{"note create", "POST", "/1/note", `{"title":"x"}`},
	{"note delete", "DELETE", "/1/note/whatever", ""},
	{"media delete", "DELETE", "/1/media/whatever", ""},
	{"user create", "POST", "/1/user", `{"userName":"x"}`},
}

func TestContentMutationsRejectPlainUser(t *testing.T) {
	authService, _, router := newContentGateRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	for _, m := range contentMutations {
		if code := gateRequest(t, router, m.method, m.path, m.body, token); http.StatusForbidden != code {
			t.Fatalf("%s: expected 403 for a plain-user token, got %d", m.name, code)
		}
	}
}

func TestContentMutationsRejectAnonymous(t *testing.T) {
	_, _, router := newContentGateRouter(t)

	for _, m := range contentMutations {
		if code := gateRequest(t, router, m.method, m.path, m.body, ""); http.StatusUnauthorized != code {
			t.Fatalf("%s: expected 401 for an anonymous call, got %d", m.name, code)
		}
	}
}

func TestContentMutationAllowsAdmin(t *testing.T) {
	authService, _, router := newContentGateRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	// an admin still writes: the create lands 200 exactly as before the gate moved
	if code := gateRequest(t, router, "POST", "/1/project", `{"title":"Admin card"}`, token); http.StatusOK != code {
		t.Fatalf("expected an admin project create to succeed with 200, got %d", code)
	}

	if code := gateRequest(t, router, "POST", "/1/note", `{"title":"Admin note"}`, token); http.StatusOK != code {
		t.Fatalf("expected an admin note create to succeed with 200, got %d", code)
	}
}

func TestPublicReadStaysOpenAfterAdminGate(t *testing.T) {
	_, publishedID, router := newContentGateRouter(t)

	// admin-gating the writes must not close the public door: the published
	// reads are still anonymous
	if code := gateRequest(t, router, "GET", "/1/project/"+publishedID, "", ""); http.StatusOK != code {
		t.Fatalf("expected 200 for an anonymous published read, got %d", code)
	}

	if code := gateRequest(t, router, "GET", "/1/project", "", ""); http.StatusOK != code {
		t.Fatalf("expected 200 for the anonymous published list, got %d", code)
	}
}
