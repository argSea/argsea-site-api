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

// newCarvingRouter wires the carving adapter behind real services on
// in-memory fakes, seeded with the shipped builtin carvings, so the
// public/admin split can be exercised end-to-end.
func newCarvingRouter(t *testing.T) (in_port.AuthService, in_port.CarvingService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	carvings := service.NewCarvingService(out_adapter.NewCarvingFakeOutAdapter(), activity)

	if err := carvings.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	router := mux.NewRouter()
	in_adapter.NewCarvingMuxAdapter(carvings, webAuth, router.PathPrefix("/1/carving").Subrouter())

	return authService, carvings, router
}

func carvingRequest(t *testing.T, router *mux.Router, method string, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestCarvingListReadIsPublic(t *testing.T) {
	_, _, router := newCarvingRouter(t)

	rec := carvingRequest(t, router, "GET", "/1/carving/carvings", "", "")

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for the anonymous list read, got %d", rec.Code)
	}

	var carvings []domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &carvings)

	if 25 != len(carvings) {
		t.Fatalf("expected all twenty-five seeded carvings, got %d", len(carvings))
	}

	for _, carving := range carvings {
		if nil == carving.BoltedTo {
			t.Fatalf("the list read handed out a null boltedTo: %+v", carving)
		}
	}
}

// carvingMutations is a representative call on every admin-gated carving
// route. The gate runs before any body parsing, so a bare path is enough to
// prove the bounce.
var carvingMutations = []struct {
	name   string
	method string
	path   string
	body   string
}{
	{"carving create", "POST", "/1/carving/carvings", `{"name":"x","svg":"<svg></svg>"}`},
	{"carving update", "PUT", "/1/carving/carvings/whatever", `{"name":"x"}`},
	{"carving delete", "DELETE", "/1/carving/carvings/whatever", ""},
	{"carving bolt", "POST", "/1/carving/carvings/whatever/bolt", `{"spot":"paw"}`},
}

func TestCarvingRoutesRejectAnonymous(t *testing.T) {
	_, _, router := newCarvingRouter(t)

	for _, m := range carvingMutations {
		if rec := carvingRequest(t, router, m.method, m.path, m.body, ""); http.StatusUnauthorized != rec.Code {
			t.Fatalf("%s: expected 401 for an anonymous call, got %d", m.name, rec.Code)
		}
	}
}

func TestCarvingRoutesRejectPlainUser(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	for _, m := range carvingMutations {
		if rec := carvingRequest(t, router, m.method, m.path, m.body, token); http.StatusForbidden != rec.Code {
			t.Fatalf("%s: expected 403 for a plain-user token, got %d", m.name, rec.Code)
		}
	}
}

func TestCarvingAdminFlowCreateUpdateDelete(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", `{"name":"rain hat","svg":"<svg>rain</svg>"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin create, got %d: %s", rec.Code, rec.Body.String())
	}

	var created domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &created)

	if "" == created.Id || created.Builtin {
		t.Fatalf("expected a stored, non-builtin carving, got %+v", created)
	}

	rec = carvingRequest(t, router, "PUT", "/1/carving/carvings/"+created.Id, `{"name":"rain hat mk2","svg":"<svg>rain</svg>"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin update, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = carvingRequest(t, router, "DELETE", "/1/carving/carvings/"+created.Id, "", token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin delete, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCarvingBoltSwapsExclusivelyOverHTTP(t *testing.T) {
	authService, carvings, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", `{"name":"new boat","svg":"<svg>new</svg>"}`, token)

	var created domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &created)

	rec = carvingRequest(t, router, "POST", "/1/carving/carvings/"+created.Id+"/bolt", `{"spot":"boat"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin bolt, got %d: %s", rec.Code, rec.Body.String())
	}

	all, _ := carvings.List()

	holders := 0
	for _, carving := range all {
		for _, spot := range carving.BoltedTo {
			if "boat" == spot {
				holders++

				if carving.Id != created.Id {
					t.Fatalf("the boat spot is still held by the old carving %+v", carving)
				}
			}
		}
	}

	if 1 != holders {
		t.Fatalf("expected exactly one holder of the boat spot, got %d", holders)
	}
}

func TestCarvingBoltRejectsUnknownSpotOverHTTP(t *testing.T) {
	authService, carvings, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	all, _ := carvings.List()
	seed := all[0]

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings/"+seed.Id+"/bolt", `{"spot":"crows-nest"}`, token)

	if http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 for an unknown spot id, got %d", rec.Code)
	}
}

func TestCarvingBoltRejectsEmptySvgOverHTTP(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", `{"name":"blank"}`, token)

	var blank domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &blank)

	rec = carvingRequest(t, router, "POST", "/1/carving/carvings/"+blank.Id+"/bolt", `{"spot":"paw"}`, token)

	if http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 bolting a carving with no svg, got %d", rec.Code)
	}
}

func TestCarvingCreateRejectsOversizedSvgOverHTTP(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	big := `{"name":"too big","svg":"<svg>` + strings.Repeat("a", 100*1024-10) + `</svg>"}` // one byte over 100KB

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", big, token)

	if http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 for an oversized svg, got %d", rec.Code)
	}
}

func TestCarvingBuiltinDeleteAndNameSvgUpdateAre409(t *testing.T) {
	authService, carvings, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	all, _ := carvings.List()

	for _, seed := range all {
		if rec := carvingRequest(t, router, "DELETE", "/1/carving/carvings/"+seed.Id, "", token); http.StatusConflict != rec.Code {
			t.Fatalf("expected 409 deleting the %s seed, got %d", seed.Name, rec.Code)
		}

		body, _ := json.Marshal(domain.Carving{Name: "defaced", Svg: seed.Svg})
		if rec := carvingRequest(t, router, "PUT", "/1/carving/carvings/"+seed.Id, string(body), token); http.StatusConflict != rec.Code {
			t.Fatalf("expected 409 renaming the %s seed, got %d", seed.Name, rec.Code)
		}
	}
}

func TestCarvingUpdateRejectsBlankSvgOnABoltedCarvingOverHTTP(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", `{"name":"new boat","svg":"<svg>new</svg>"}`, token)

	var created domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &created)

	rec = carvingRequest(t, router, "POST", "/1/carving/carvings/"+created.Id+"/bolt", `{"spot":"boat"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin bolt, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = carvingRequest(t, router, "PUT", "/1/carving/carvings/"+created.Id, `{"name":"new boat","svg":""}`, token)

	if http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 blanking a bolted carving's svg, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCarvingDeleteRejectsBoltedCarvingOverHTTP(t *testing.T) {
	authService, _, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := carvingRequest(t, router, "POST", "/1/carving/carvings", `{"name":"new boat","svg":"<svg>new</svg>"}`, token)

	var created domain.Carving
	json.Unmarshal(rec.Body.Bytes(), &created)

	rec = carvingRequest(t, router, "POST", "/1/carving/carvings/"+created.Id+"/bolt", `{"spot":"boat"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an admin bolt, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = carvingRequest(t, router, "DELETE", "/1/carving/carvings/"+created.Id, "", token)

	if http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 deleting a bolted carving, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCarvingBuiltinBoltAcceptsOverHTTP(t *testing.T) {
	authService, carvings, router := newCarvingRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	all, _ := carvings.List()
	seed := all[0]
	spot := seed.BoltedTo[0]

	// boltedTo stays mutable on a builtin, unlike name/svg
	rec := carvingRequest(t, router, "POST", "/1/carving/carvings/"+seed.Id+"/bolt", `{"spot":"`+spot+`"}`, token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 re-bolting a builtin's own spot, got %d: %s", rec.Code, rec.Body.String())
	}
}
