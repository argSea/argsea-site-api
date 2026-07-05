package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newProjectRouter wires the project adapter behind real services on in-memory
// fakes, seeded with one draft and one published project, so the read-gating
// rules can be exercised end-to-end.
func newProjectRouter(t *testing.T) (string, string, string, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	projects := service.NewProjectCRUDService(out_adapter.NewProjectFakeOutAdapter(), revisions, activity)

	draft, err := projects.Create(domain.Project{Title: "Draft card"})

	if nil != err {
		t.Fatalf("seed draft failed: %v", err)
	}

	published, err := projects.Create(domain.Project{Title: "Published card"})

	if nil != err {
		t.Fatalf("seed published failed: %v", err)
	}

	if _, err := projects.Publish(published.Id); nil != err {
		t.Fatalf("seed publish failed: %v", err)
	}

	router := mux.NewRouter()
	in_adapter.NewProjectMuxAdapter(projects, webAuth, router.PathPrefix("/1/project").Subrouter())

	token := mintToken(t, authService, time.Now().Add(time.Hour))

	return draft.Id, published.Id, token, router
}

func getProjects(t *testing.T, router *mux.Router, path string, token string) (int, []domain.Project) {
	t.Helper()

	req := httptest.NewRequest("GET", path, nil)

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var projects []domain.Project
	json.Unmarshal(rec.Body.Bytes(), &projects)

	return rec.Code, projects
}

func TestUnauthListExcludesDrafts(t *testing.T) {
	draftID, publishedID, _, router := newProjectRouter(t)

	code, projects := getProjects(t, router, "/1/project", "")

	if http.StatusOK != code {
		t.Fatalf("expected 200, got %d", code)
	}

	for _, p := range projects {
		if p.Id == draftID {
			t.Fatalf("unauthenticated list must not contain drafts")
		}
	}

	if 1 != len(projects) || projects[0].Id != publishedID {
		t.Fatalf("expected exactly the published project, got %d entries", len(projects))
	}
}

func TestAuthedListIncludesDrafts(t *testing.T) {
	draftID, _, token, router := newProjectRouter(t)

	code, projects := getProjects(t, router, "/1/project", token)

	if http.StatusOK != code {
		t.Fatalf("expected 200, got %d", code)
	}

	found := false

	for _, p := range projects {
		if p.Id == draftID {
			found = true
		}
	}

	if !found || 2 != len(projects) {
		t.Fatalf("authenticated list should include the draft (got %d entries, draft found=%v)", len(projects), found)
	}
}

func TestUnauthDraftByIdIs404(t *testing.T) {
	draftID, _, _, router := newProjectRouter(t)

	req := httptest.NewRequest("GET", "/1/project/"+draftID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// 404, not 401 — a draft's existence is not confirmed to the public
	if http.StatusNotFound != rec.Code {
		t.Fatalf("expected 404 for an unauthenticated draft read, got %d", rec.Code)
	}
}

func TestUnauthPublishedByIdIsVisible(t *testing.T) {
	_, publishedID, _, router := newProjectRouter(t)

	req := httptest.NewRequest("GET", "/1/project/"+publishedID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for a published document, got %d", rec.Code)
	}
}

func TestCreateWithInvalidStampIs400(t *testing.T) {
	_, _, token, router := newProjectRouter(t)

	// ink outside the enum — the exact XSS vector the gate exists for
	body := strings.NewReader(`{"title":"Bad stamp","stamp":{"shape":"rect","motif":"sun","ink":"expression(alert(1))"}}`)

	req := httptest.NewRequest("POST", "/1/project", body)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 for an invalid stamp, got %d", rec.Code)
	}

	// the rejection uses the standard errored response envelope
	var envelope map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &envelope)

	if "error" != envelope["status"] {
		t.Fatalf("expected the error envelope, got %s", rec.Body.String())
	}

	// and nothing was written — the authed list still holds only the seeds
	code, projects := getProjects(t, router, "/1/project", token)

	if http.StatusOK != code || 2 != len(projects) {
		t.Fatalf("rejected create must persist nothing, got %d projects", len(projects))
	}
}

func TestReorderWithoutOrderFieldIs400(t *testing.T) {
	draftID, _, token, router := newProjectRouter(t)

	// a reorder must say where the postcard goes — an empty object and broken
	// JSON are both rejected before the service is reached
	for _, body := range []string{`{}`, `{"order":null}`, `not-json`} {
		req := httptest.NewRequest("POST", "/1/project/"+draftID+"/reorder", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if http.StatusBadRequest != rec.Code {
			t.Fatalf("expected 400 for reorder body %q, got %d", body, rec.Code)
		}
	}

	// and the postcard never moved — the draft still holds its seeded position
	req := httptest.NewRequest("GET", "/1/project/"+draftID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var project domain.Project
	json.Unmarshal(rec.Body.Bytes(), &project)

	if 1 != project.Order {
		t.Fatalf("a rejected reorder must not move the postcard, got order %d", project.Order)
	}
}

func TestAuthedDraftByIdIsVisible(t *testing.T) {
	draftID, _, token, router := newProjectRouter(t)

	req := httptest.NewRequest("GET", "/1/project/"+draftID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an authenticated draft read, got %d", rec.Code)
	}
}
