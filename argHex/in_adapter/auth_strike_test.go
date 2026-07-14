package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
)

const adriftPath = "/1/auth/adrift/"

// postLogin fires a login from a given client IP (forwarded as nginx would),
// optionally carrying the console marker the admin SPA sets on its own request.
func postLogin(t *testing.T, router *mux.Router, body string, ip string, console bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest("POST", "/1/auth/login/", strings.NewReader(body))

	if "" != ip {
		req.Header.Set("X-Forwarded-For", ip)
	}

	if console {
		req.Header.Set("X-Argsea-Console", "1")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func errorMessage(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var response data_objects.ErroredResponseObject

	if err := json.Unmarshal(rec.Body.Bytes(), &response); nil != err {
		t.Fatalf("expected a JSON error envelope, got %q", rec.Body.String())
	}

	message, ok := response.Message.(string)

	if !ok {
		t.Fatalf("expected a string message, got %v", response.Message)
	}

	return message
}

func TestFailedConsoleHailGets400JSON(t *testing.T) {
	_, router := newLoginRouter(t, in_port.PERM_ADMIN)

	rec := postLogin(t, router, `{"userName":"meo","password":"nope"}`, "203.0.113.40", true)

	if http.StatusBadRequest != rec.Code {
		t.Fatalf("a failed console hail must get a 400, got %d", rec.Code)
	}

	if in_port.ErrBadCredentials.Error() != errorMessage(t, rec) {
		t.Fatalf("a bad-credentials console hail must show the generic line, got %q", errorMessage(t, rec))
	}
}

func TestFailedDirectHailIsSentAdrift(t *testing.T) {
	_, router := newLoginRouter(t, in_port.PERM_ADMIN)

	rec := postLogin(t, router, `{"userName":"meo","password":"nope"}`, "203.0.113.40", false)

	if http.StatusFound != rec.Code {
		t.Fatalf("a failed direct hail must get a 302, got %d", rec.Code)
	}

	if adriftPath != rec.Header().Get("Location") {
		t.Fatalf("a failed direct hail must point at the adrift trap, got %q", rec.Header().Get("Location"))
	}

	if 0 != rec.Body.Len() && strings.Contains(rec.Body.String(), "credentials") {
		t.Fatalf("a direct hail must not leak a JSON credentials body, got %q", rec.Body.String())
	}
}

func TestAdriftRedirectsToItself(t *testing.T) {
	_, router := newLoginRouter(t, in_port.PERM_ADMIN)

	req := httptest.NewRequest("GET", adriftPath, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusFound != rec.Code {
		t.Fatalf("the adrift trap must answer a 302, got %d", rec.Code)
	}

	if adriftPath != rec.Header().Get("Location") {
		t.Fatalf("the adrift trap must redirect to itself to loop, got %q", rec.Header().Get("Location"))
	}
}

func TestStruckConsoleHailShowsTheKeeperLine(t *testing.T) {
	_, router := newLoginRouter(t, in_port.PERM_ADMIN)
	ip := "203.0.113.50"

	// six bad hails from one IP strike the light; the sixth already reads struck
	var last *httptest.ResponseRecorder

	for miss := 1; miss <= 6; miss++ {
		last = postLogin(t, router, `{"userName":"meo","password":"nope"}`, ip, true)
	}

	if http.StatusBadRequest != last.Code {
		t.Fatalf("a struck console hail stays a 400, got %d", last.Code)
	}

	if in_port.ErrLoginStruck.Error() != errorMessage(t, last) {
		t.Fatalf("a struck console hail must show the keeper struck line, got %q", errorMessage(t, last))
	}

	// even the correct password is refused while struck, still a 400 to the console
	struck := postLogin(t, router, `{"userName":"meo","password":"passphrase"}`, ip, true)

	if http.StatusBadRequest != struck.Code || in_port.ErrLoginStruck.Error() != errorMessage(t, struck) {
		t.Fatalf("a struck IP must refuse even the right password with the struck line, got %d %q", struck.Code, errorMessage(t, struck))
	}
}
