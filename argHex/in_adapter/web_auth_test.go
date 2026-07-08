package in_adapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

var testSecret = []byte("web-auth-test-secret")

// newAuthedRouter wires the activity adapter (an auth-gated endpoint) behind a
// real JWT service and the shared WebAuth, so requests exercise the exact
// extraction + validation path production uses.
func newAuthedRouter(t *testing.T) (in_port.AuthService, *in_adapter.WebAuth, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	activityService := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	router := mux.NewRouter()
	in_adapter.NewActivityMuxAdapter(activityService, webAuth, router.PathPrefix("/1/activity").Subrouter())

	return authService, webAuth, router
}

func mintToken(t *testing.T, authService in_port.AuthService, expires time.Time) string {
	t.Helper()

	token, err := authService.Generate("keeper", expires, []string{in_port.PERM_ADMIN})

	if nil != err {
		t.Fatalf("could not mint token: %v", err)
	}

	return token
}

func TestBearerTokenAuthorizes(t *testing.T) {
	authService, _, router := newAuthedRouter(t)
	token := mintToken(t, authService, time.Now().Add(time.Hour))

	req := httptest.NewRequest("GET", "/1/activity", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 with a valid bearer token, got %d", rec.Code)
	}
}

func TestSessionCookieAuthorizes(t *testing.T) {
	authService, webAuth, router := newAuthedRouter(t)
	token := mintToken(t, authService, time.Now().Add(time.Hour))

	// issue the cookie exactly the way the login handler does: through the
	// shared store
	seed := httptest.NewRequest("GET", "/1/activity", nil)
	issuer := httptest.NewRecorder()
	session, _ := webAuth.Store().Get(seed, "auth-token")
	session.Values["token"] = token

	if err := session.Save(seed, issuer); nil != err {
		t.Fatalf("could not issue session cookie: %v", err)
	}

	req := httptest.NewRequest("GET", "/1/activity", nil)

	for _, cookie := range issuer.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 with a valid session cookie, got %d", rec.Code)
	}
}

func TestCookieDomainIsConfigurable(t *testing.T) {
	authService := service.NewJWTAuthService(testSecret)

	// the domain baked into the session cookie comes from the constructor;
	// main.go feeds it auth.cookie_domain from the config
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "cookies.example")

	seed := httptest.NewRequest("GET", "/1/activity", nil)
	issuer := httptest.NewRecorder()
	session, _ := webAuth.Store().Get(seed, "auth-token")
	session.Values["token"] = "whatever"

	if err := session.Save(seed, issuer); nil != err {
		t.Fatalf("could not issue session cookie: %v", err)
	}

	cookies := issuer.Result().Cookies()

	if 1 != len(cookies) || "cookies.example" != cookies[0].Domain {
		t.Fatalf("expected the configured cookie domain, got %+v", cookies)
	}
}

func TestMissingTokenIsRejected(t *testing.T) {
	_, _, router := newAuthedRouter(t)

	req := httptest.NewRequest("GET", "/1/activity", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 without a token, got %d", rec.Code)
	}
}

func TestGarbageTokenIsRejected(t *testing.T) {
	_, _, router := newAuthedRouter(t)

	req := httptest.NewRequest("GET", "/1/activity", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for a garbage token, got %d", rec.Code)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	authService, _, router := newAuthedRouter(t)
	token := mintToken(t, authService, time.Now().Add(-time.Hour))

	req := httptest.NewRequest("GET", "/1/activity", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for an expired token, got %d", rec.Code)
	}
}

func TestTokenSignedWithWrongSecretIsRejected(t *testing.T) {
	_, _, router := newAuthedRouter(t)

	// a structurally valid JWT minted under a different secret must not pass
	foreign := service.NewJWTAuthService([]byte("some-other-secret"))
	token := mintToken(t, foreign, time.Now().Add(time.Hour))

	req := httptest.NewRequest("GET", "/1/activity", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for a foreign-signed token, got %d", rec.Code)
	}
}
