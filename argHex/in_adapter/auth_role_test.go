package in_adapter_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"encoding/json"
	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// stubUserRepo is a single-user UserRepo whose stored document (including its
// role) the test controls, and which captures what Add and Set receive so the
// role-stripping on signup and update can be asserted.
type stubUserRepo struct {
	stored  domain.User
	added   *domain.User
	updated *domain.User
}

func (s *stubUserRepo) GetAll(limit int64, offset int64, sort interface{}) domain.Users {
	return domain.Users{s.stored}
}

func (s *stubUserRepo) Get(id string) domain.User {
	return s.stored
}

func (s *stubUserRepo) GetByUserName(username string) domain.User {
	return s.stored
}

func (s *stubUserRepo) Set(user domain.User) error {
	if nil != s.updated {
		*s.updated = user
	}

	return nil
}

func (s *stubUserRepo) Add(user domain.User) (string, error) {
	*s.added = user
	return "new-user", nil
}

func (s *stubUserRepo) Remove(user domain.User) error {
	return nil
}

// newLoginRouter wires the real auth adapter + login service over a stored
// user document with the given role, so the login → token path is exercised
// exactly as production runs it.
func newLoginRouter(t *testing.T, storedRole string) (in_port.AuthService, *mux.Router) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("passphrase"), bcrypt.MinCost)

	if nil != err {
		t.Fatalf("could not hash password: %v", err)
	}

	repo := &stubUserRepo{
		stored: domain.User{
			Id:       "keeper",
			UserName: "meo",
			Password: domain.Password(hash),
			Role:     storedRole,
		},
		added: &domain.User{},
	}

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")
	loginService := service.NewUserLoginService(repo)

	router := mux.NewRouter()
	in_adapter.NewAuthMuxAdapter(authService, loginService, webAuth, router.PathPrefix("/1/auth").Subrouter())

	return authService, router
}

// login posts credentials (plus any extra body fields) and returns the minted
// token's validated role claim.
func loginRole(t *testing.T, authService in_port.AuthService, router *mux.Router, body string) string {
	t.Helper()

	req := httptest.NewRequest("POST", "/1/auth/login/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 from login, got %d", rec.Code)
	}

	var response data_objects.LoginResponseObject
	json.Unmarshal(rec.Body.Bytes(), &response)

	validation, err := authService.Validate(response.Token)

	if nil != err || !validation.Valid {
		t.Fatalf("login returned an invalid token: %v", err)
	}

	return validation.Role
}

func TestLoginMintsStoredAdminRole(t *testing.T) {
	authService, router := newLoginRouter(t, in_port.PERM_ADMIN)

	role := loginRole(t, authService, router, `{"userName":"meo","password":"passphrase"}`)

	if in_port.PERM_ADMIN != role {
		t.Fatalf("expected an admin-role token for a stored admin, got %q", role)
	}
}

func TestLoginDefaultsToUserRole(t *testing.T) {
	authService, router := newLoginRouter(t, "")

	role := loginRole(t, authService, router, `{"userName":"meo","password":"passphrase"}`)

	if in_port.PERM_USER != role {
		t.Fatalf("expected a plain-user token for a role-less document, got %q", role)
	}
}

func TestLoginIgnoresRoleInRequestBody(t *testing.T) {
	authService, router := newLoginRouter(t, "")

	// a plain user claiming admin in the login body must still get a user token
	role := loginRole(t, authService, router, `{"userName":"meo","password":"passphrase","role":"admin"}`)

	if in_port.PERM_USER != role {
		t.Fatalf("expected the body role to be ignored, got %q", role)
	}
}

func TestSignupStripsRole(t *testing.T) {
	repo := &stubUserRepo{added: &domain.User{}}

	var loginService in_port.UserLoginService = service.NewUserLoginService(repo)
	var _ out_port.UserRepo = repo

	if _, err := loginService.Signup(domain.User{UserName: "sneak", Role: in_port.PERM_ADMIN}); nil != err {
		t.Fatalf("signup failed: %v", err)
	}

	if "" != repo.added.Role {
		t.Fatalf("signup must strip the role, stored %q", repo.added.Role)
	}
}
