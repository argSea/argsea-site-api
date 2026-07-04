package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	webAuth := in_adapter.NewWebAuth(authService, testSecret)

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
