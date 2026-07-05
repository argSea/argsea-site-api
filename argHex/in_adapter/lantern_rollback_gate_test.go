package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newRollbackRouter mounts the lantern adapter over a scriptable fake release
// store, exposing the store so a test can control what Previous returns.
func newRollbackRouter(t *testing.T, previous string, gate chan struct{}) (in_port.AuthService, *out_adapter.LanternFakeReleaseStore, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	releases := &out_adapter.LanternFakeReleaseStore{Prev: previous}
	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		&out_adapter.LanternFakeRunner{Gate: gate},
		releases,
		&out_adapter.LanternFakeStateRepo{},
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewLanternMuxAdapter(lantern, webAuth, router.PathPrefix("/1/lantern").Subrouter())

	return authService, releases, router
}

func TestRollbackIsAdminGated(t *testing.T) {
	authService, _, router := newRollbackRouter(t, "gen-older", nil)

	if rec := lanternRequest(t, router, "POST", "/1/lantern/rollback/", ""); http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for an anonymous rollback, got %d", rec.Code)
	}

	token := mintRoleToken(t, authService, in_port.PERM_USER)

	if rec := lanternRequest(t, router, "POST", "/1/lantern/rollback/", token); http.StatusForbidden != rec.Code {
		t.Fatalf("expected 403 for a plain-user rollback, got %d", rec.Code)
	}
}

func TestRollbackReturns200WithStatus(t *testing.T) {
	authService, releases, router := newRollbackRouter(t, "gen-older", nil)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	// both slash spellings work, same handling as hoist
	for _, path := range []string{"/1/lantern/rollback", "/1/lantern/rollback/"} {
		rec := lanternRequest(t, router, "POST", path, token)

		if http.StatusOK != rec.Code {
			t.Fatalf("expected 200 from %s, got %d (%s)", path, rec.Code, rec.Body.String())
		}
	}

	var status domain.LanternStatus

	json.Unmarshal(lanternRequest(t, router, "POST", "/1/lantern/rollback/", token).Body.Bytes(), &status)

	if domain.LanternIdle != status.State {
		t.Fatalf("expected the idle status in the body, got %+v", status)
	}

	if 0 == len(releases.Swapped) || "gen-older" != releases.Swapped[0] {
		t.Fatalf("expected the live link swapped to gen-older, got %v", releases.Swapped)
	}
}

func TestRollbackWithoutPreviousBuildIs409(t *testing.T) {
	authService, _, router := newRollbackRouter(t, "", nil)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := lanternRequest(t, router, "POST", "/1/lantern/rollback/", token)

	if http.StatusConflict != rec.Code {
		t.Fatalf("expected 409 with no previous build, got %d", rec.Code)
	}

	// both 409 shapes carry the LanternStatus body, mirroring hoist — here the
	// state is idle, which is how the admin tells "nothing kept" from "in flight"
	var status domain.LanternStatus
	json.Unmarshal(rec.Body.Bytes(), &status)

	if domain.LanternIdle != status.State {
		t.Fatalf("expected the idle status in the 409 body, got %s", rec.Body.String())
	}
}

func TestRollbackDuringHoistIs409(t *testing.T) {
	gate := make(chan struct{})
	authService, _, router := newRollbackRouter(t, "gen-older", gate)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	defer close(gate)

	if rec := lanternRequest(t, router, "POST", "/1/lantern/hoist/", token); http.StatusAccepted != rec.Code {
		t.Fatalf("hoist failed to start: %d", rec.Code)
	}

	rec := lanternRequest(t, router, "POST", "/1/lantern/rollback/", token)

	var status domain.LanternStatus
	json.Unmarshal(rec.Body.Bytes(), &status)

	if http.StatusConflict != rec.Code || domain.LanternBuilding != status.State {
		t.Fatalf("expected a 409 carrying the running status, got %d %+v", rec.Code, status)
	}
}
