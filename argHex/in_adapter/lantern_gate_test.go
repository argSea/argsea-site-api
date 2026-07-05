package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newLanternRouter mounts the lantern adapter over a gated fake runner, so the
// hoist stays observably "building" for the duration of a test.
func newLanternRouter(t *testing.T) (in_port.AuthService, chan struct{}, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	gate := make(chan struct{})
	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		&out_adapter.LanternFakeRunner{Gate: gate},
		&out_adapter.LanternFakeReleaseStore{},
		&out_adapter.LanternFakeStateRepo{},
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewLanternMuxAdapter(lantern, webAuth, router.PathPrefix("/1/lantern").Subrouter())

	// release the gated build no matter how the test exits
	t.Cleanup(func() {
		select {
		case <-gate:
		default:
			close(gate)
		}
	})

	return authService, gate, router
}

// mintRoleToken mints a valid token carrying the given role.
func mintRoleToken(t *testing.T, authService in_port.AuthService, role string) string {
	t.Helper()

	token, err := authService.Generate("keeper", time.Now().Add(time.Hour), []string{role})

	if nil != err {
		t.Fatalf("could not mint token: %v", err)
	}

	return token
}

func lanternRequest(t *testing.T, router *mux.Router, method string, path string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestLanternWithoutTokenIs401(t *testing.T) {
	_, _, router := newLanternRouter(t)

	if rec := lanternRequest(t, router, "POST", "/1/lantern/hoist/", ""); http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for an anonymous hoist, got %d", rec.Code)
	}

	if rec := lanternRequest(t, router, "GET", "/1/lantern/", ""); http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for an anonymous status read, got %d", rec.Code)
	}
}

func TestLanternWithPlainUserTokenIs403(t *testing.T) {
	authService, _, router := newLanternRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	if rec := lanternRequest(t, router, "POST", "/1/lantern/hoist/", token); http.StatusForbidden != rec.Code {
		t.Fatalf("expected 403 for a plain-user hoist, got %d", rec.Code)
	}

	if rec := lanternRequest(t, router, "GET", "/1/lantern/", token); http.StatusForbidden != rec.Code {
		t.Fatalf("expected 403 for a plain-user status read, got %d", rec.Code)
	}
}

func TestLanternHoistFlow202Then409(t *testing.T) {
	authService, _, router := newLanternRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	// idle status first
	rec := lanternRequest(t, router, "GET", "/1/lantern/", token)

	var status domain.LanternStatus
	json.Unmarshal(rec.Body.Bytes(), &status)

	if http.StatusOK != rec.Code || domain.LanternIdle != status.State {
		t.Fatalf("expected an idle 200 status, got %d %+v", rec.Code, status)
	}

	// first hoist: accepted, reports building
	rec = lanternRequest(t, router, "POST", "/1/lantern/hoist/", token)
	json.Unmarshal(rec.Body.Bytes(), &status)

	if http.StatusAccepted != rec.Code || domain.LanternBuilding != status.State {
		t.Fatalf("expected a 202 building status, got %d %+v", rec.Code, status)
	}

	// second hoist while the build is gated: conflict, carries the running status
	rec = lanternRequest(t, router, "POST", "/1/lantern/hoist/", token)
	json.Unmarshal(rec.Body.Bytes(), &status)

	if http.StatusConflict != rec.Code || domain.LanternBuilding != status.State {
		t.Fatalf("expected a 409 carrying the running status, got %d %+v", rec.Code, status)
	}
}
