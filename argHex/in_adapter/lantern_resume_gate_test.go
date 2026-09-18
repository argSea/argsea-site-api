package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newBareShelfLanternRouter mounts the lantern adapter over a resume shelf with
// nothing published, the state the hoist refuses on.
func newBareShelfLanternRouter(t *testing.T) (in_port.AuthService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	resumes := service.NewResumeService(
		out_adapter.NewResumeFakeOutAdapter(),
		out_adapter.NewMediaWebstoreAdapter("", "/media/images/"),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		&out_adapter.LanternFakeRunner{},
		&out_adapter.LanternFakeReleaseStore{},
		&out_adapter.LanternFakeStateRepo{},
		resumes,
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewLanternMuxAdapter(lantern, webAuth, router.PathPrefix("/1/lantern").Subrouter())

	return authService, router
}

// TestHoistWithNoPublishedResumeIs412WithTheReason pins the code as hard as the
// body. It must not be 409: the installed admin reads a 409 on this route as
// "a hoist is already running" without looking at the body, so refusing here
// with one would have the deploy button report a hoist that does not exist.
func TestHoistWithNoPublishedResumeIs412WithTheReason(t *testing.T) {
	authService, router := newBareShelfLanternRouter(t)
	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	rec := lanternRequest(t, router, "POST", "/1/lantern/hoist/", token)

	if http.StatusPreconditionFailed != rec.Code {
		t.Fatalf("expected 412 when nothing is published, got %d", rec.Code)
	}

	var body data_objects.ErroredResponseObject
	json.Unmarshal(rec.Body.Bytes(), &body)

	// the state fields say nothing about a bare shelf, so the refusal has to
	// carry its own reason
	if in_port.ErrNoPublishedResume.Error() != body.Message {
		t.Fatalf("expected the refusal's reason in the body, got %+v", body)
	}
}
