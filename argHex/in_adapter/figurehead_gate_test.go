package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newFigureheadRouter wires the figurehead adapter behind real services on
// in-memory fakes, seeded with the two v1 cats, so the public/admin split can
// be exercised end-to-end.
func newFigureheadRouter(t *testing.T) (in_port.AuthService, in_port.FigureheadService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	figureheads := service.NewFigureheadService(out_adapter.NewCatDesignFakeOutAdapter(), activity)

	if err := figureheads.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	router := mux.NewRouter()
	in_adapter.NewFigureheadMuxAdapter(figureheads, webAuth, router.PathPrefix("/1/figurehead").Subrouter())

	return authService, figureheads, router
}

func figureheadRequest(t *testing.T, router *mux.Router, method string, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestPublishedReadIsPublic(t *testing.T) {
	_, _, router := newFigureheadRouter(t)

	rec := figureheadRequest(t, router, "GET", "/1/figurehead/published", "", "")

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for the anonymous published read, got %d", rec.Code)
	}

	var designs []domain.CatDesign
	json.Unmarshal(rec.Body.Bytes(), &designs)

	if 2 != len(designs) {
		t.Fatalf("expected both seeded poses, got %d designs", len(designs))
	}

	poses := map[string]bool{}

	for _, design := range designs {
		poses[design.Pose] = true

		if !design.Published || 0 == len(design.Shapes) {
			t.Fatalf("the published read handed out a bare design: %+v", design)
		}
	}

	if !poses[domain.PosePerched] || !poses[domain.PoseLying] {
		t.Fatalf("expected one design per pose, got %v", poses)
	}
}

// figureheadMutations is a representative call on every admin-gated figurehead
// route. The gate runs before any body parsing, so a bare path is enough to
// prove the bounce.
var figureheadMutations = []struct {
	name   string
	method string
	path   string
	body   string
}{
	{"designs list", "GET", "/1/figurehead/designs", ""},
	{"design create", "POST", "/1/figurehead/designs", `{"pose":"lying","label":"x"}`},
	{"design update", "PUT", "/1/figurehead/designs/whatever", `{"label":"x"}`},
	{"design delete", "DELETE", "/1/figurehead/designs/whatever", ""},
	{"design publish", "POST", "/1/figurehead/designs/whatever/publish", ""},
}

func TestFigureheadRoutesRejectAnonymous(t *testing.T) {
	_, _, router := newFigureheadRouter(t)

	for _, m := range figureheadMutations {
		if rec := figureheadRequest(t, router, m.method, m.path, m.body, ""); http.StatusUnauthorized != rec.Code {
			t.Fatalf("%s: expected 401 for an anonymous call, got %d", m.name, rec.Code)
		}
	}
}

func TestFigureheadRoutesRejectPlainUser(t *testing.T) {
	authService, _, router := newFigureheadRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	for _, m := range figureheadMutations {
		if rec := figureheadRequest(t, router, m.method, m.path, m.body, token); http.StatusForbidden != rec.Code {
			t.Fatalf("%s: expected 403 for a plain-user token, got %d", m.name, rec.Code)
		}
	}
}

func TestFigureheadAdminFlowCreatePublishDelete(t *testing.T) {
	authService, figureheads, router := newFigureheadRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := figureheadRequest(t, router, "POST", "/1/figurehead/designs", `{"pose":"lying","label":"rain hat","viewBox":"0 0 100 48","shapes":[{"id":"tail","type":"path","d":"M0 0","fill":"#232a4d"}]}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin create, got %d: %s", rec.Code, rec.Body.String())
	}

	var created domain.CatDesign
	json.Unmarshal(rec.Body.Bytes(), &created)

	if "" == created.Id || created.Published {
		t.Fatalf("expected a stored draft, got %+v", created)
	}

	if rec := figureheadRequest(t, router, "POST", "/1/figurehead/designs/"+created.Id+"/publish", "", token); http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin publish, got %d", rec.Code)
	}

	// the anonymous published read now hands out the new lying cat
	rec = figureheadRequest(t, router, "GET", "/1/figurehead/published", "", "")

	var published []domain.CatDesign
	json.Unmarshal(rec.Body.Bytes(), &published)

	for _, design := range published {
		if domain.PoseLying == design.Pose && design.Id != created.Id {
			t.Fatalf("the published read still hands out the old lying cat: %+v", design)
		}
	}

	// deleting it while published is refused; superseding frees it
	if rec := figureheadRequest(t, router, "DELETE", "/1/figurehead/designs/"+created.Id, "", token); http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 deleting a published design, got %d", rec.Code)
	}

	current := map[string]string{}

	for _, design := range published {
		current[design.Pose] = design.Id
	}

	seeds, _ := figureheads.List()

	for _, seed := range seeds {
		if seed.Seed && domain.PoseLying == seed.Pose {
			if rec := figureheadRequest(t, router, "POST", "/1/figurehead/designs/"+seed.Id+"/publish", "", token); http.StatusOK != rec.Code {
				t.Fatalf("expected 200 re-publishing the v1 seed, got %d", rec.Code)
			}
		}
	}

	if rec := figureheadRequest(t, router, "DELETE", "/1/figurehead/designs/"+created.Id, "", token); http.StatusOK != rec.Code {
		t.Fatalf("expected 200 deleting the superseded design, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFigureheadSeedDeleteAndUpdateAre409(t *testing.T) {
	authService, figureheads, router := newFigureheadRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	designs, _ := figureheads.List()

	for _, seed := range designs {
		if rec := figureheadRequest(t, router, "DELETE", "/1/figurehead/designs/"+seed.Id, "", token); http.StatusConflict != rec.Code {
			t.Fatalf("expected 409 deleting the %s seed, got %d", seed.Pose, rec.Code)
		}

		if rec := figureheadRequest(t, router, "PUT", "/1/figurehead/designs/"+seed.Id, `{"label":"defaced"}`, token); http.StatusConflict != rec.Code {
			t.Fatalf("expected 409 editing the %s seed, got %d", seed.Pose, rec.Code)
		}
	}
}
