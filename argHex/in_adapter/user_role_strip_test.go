package in_adapter_test

import (
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
	"go.mongodb.org/mongo-driver/bson"
)

// newUserRouter mounts the real user adapter over the capturing stub repo, so
// the role-stripping is exercised at the HTTP seam — body in, repo call out.
func newUserRouter(t *testing.T) (*stubUserRepo, string, *mux.Router) {
	t.Helper()

	repo := &stubUserRepo{
		added:   &domain.User{},
		updated: &domain.User{},
	}

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	userService := service.NewUserCRUDService(repo)
	mediaService := service.NewMediaService(
		out_adapter.NewMediaWebstoreAdapter("", ""),
		out_adapter.NewMediaMetaFakeOutAdapter(),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewUserMuxAdapter(userService, mediaService, webAuth, router.PathPrefix("/1/user").Subrouter())

	token := mintRoleToken(t, authService, in_port.PERM_USER)

	return repo, token, router
}

func userRequest(t *testing.T, router *mux.Router, method string, path string, token string, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestUserCreateStripsBodyRole(t *testing.T) {
	repo, token, router := newUserRouter(t)

	rec := userRequest(t, router, "POST", "/1/user", token, `{"userName":"sneak","role":"admin"}`)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected the create to go through, got %d", rec.Code)
	}

	if "sneak" != repo.added.UserName {
		t.Fatalf("the create never reached the repo: %+v", repo.added)
	}

	if "" != repo.added.Role {
		t.Fatalf("signup must strip a body role before the repo sees it, stored %q", repo.added.Role)
	}
}

func TestUserUpdateCannotSelfGrantAdmin(t *testing.T) {
	repo, token, router := newUserRouter(t)

	// the token belongs to "keeper", so the self-update path is the one that
	// passes the identity gate
	rec := userRequest(t, router, "PUT", "/1/user/keeper", token, `{"userName":"meo","role":"admin"}`)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected the update to go through, got %d", rec.Code)
	}

	if "meo" != repo.updated.UserName {
		t.Fatalf("the update never reached the repo: %+v", repo.updated)
	}

	if "" != repo.updated.Role {
		t.Fatalf("an update must strip a body role before the repo sees it, stored %q", repo.updated.Role)
	}

	// the second half of the contract: with the role empty, its bson-omitempty
	// tag drops the key entirely, so the mongo adapter's $set update leaves the
	// stored role untouched — a PUT can never self-grant admin
	raw, err := bson.Marshal(*repo.updated)

	if nil != err {
		t.Fatalf("could not bson-marshal the updated user: %v", err)
	}

	var doc bson.M

	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("could not decode the marshaled user: %v", err)
	}

	if _, present := doc["role"]; present {
		t.Fatalf("an empty role must vanish from the $set document, got %+v", doc)
	}
}
