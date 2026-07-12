package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newSightingRouter wires the sighting adapter over the in-memory fake behind a
// real JWT service, so the public/authed split is exercised end-to-end.
func newSightingRouter(t *testing.T) (in_port.AuthService, *mux.Router) {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")
	sightings := service.NewSightingService(out_adapter.NewSightingFakeOutAdapter(), "gate-test-salt")

	router := mux.NewRouter()
	in_adapter.NewSightingMuxAdapter(sightings, webAuth, router.PathPrefix("/1/sighting").Subrouter())

	return authService, router
}

// sightingRequest fires a request carrying a human user agent (so the ingest
// does not read it as a bot) unless the caller overrides it.
func sightingRequest(t *testing.T, router *mux.Router, method string, path string, body string, contentType string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/128.0")

	if "" != contentType {
		req.Header.Set("Content-Type", contentType)
	}

	if "" != token {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestSightingIngestIsPublicAndTakesTextPlain(t *testing.T) {
	_, router := newSightingRouter(t)

	// navigator.sendBeacon posts a bare text/plain body with no token
	rec := sightingRequest(t, router, "POST", "/1/sighting/", `{"kind":"sail","path":"/projects/foo","ref":"https://www.google.com/"}`, "text/plain;charset=UTF-8", "")

	if http.StatusNoContent != rec.Code {
		t.Fatalf("expected 204 for an anonymous text/plain beacon, got %d: %s", rec.Code, rec.Body.String())
	}

	if 0 != rec.Body.Len() {
		t.Fatalf("ingest must never echo a body, got %q", rec.Body.String())
	}
}

func TestSightingIngestTakesApplicationJson(t *testing.T) {
	_, router := newSightingRouter(t)

	rec := sightingRequest(t, router, "POST", "/1/sighting/", `{"kind":"flip","path":"/projects/foo","subject":"cat-cascade"}`, "application/json", "")

	if http.StatusNoContent != rec.Code {
		t.Fatalf("expected 204 for a json beacon, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSightingIngestRejectsJunk(t *testing.T) {
	_, router := newSightingRouter(t)

	junk := []struct {
		name string
		body string
	}{
		{"unknown kind", `{"kind":"click","path":"/"}`},
		{"junk path", `{"kind":"sail","path":"nope"}`},
		{"malformed json", `not json at all`},
		{"empty body", ``},
	}

	for _, j := range junk {
		if rec := sightingRequest(t, router, "POST", "/1/sighting/", j.body, "text/plain", ""); http.StatusBadRequest != rec.Code {
			t.Fatalf("%s: expected 400, got %d", j.name, rec.Code)
		}
	}
}

func TestSightingIngestDropsBotsWithoutError(t *testing.T) {
	_, router := newSightingRouter(t)

	req := httptest.NewRequest("POST", "/1/sighting/", strings.NewReader(`{"kind":"sail","path":"/"}`))
	req.Header.Set("User-Agent", "Googlebot/2.1 (+http://www.google.com/bot.html)")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if http.StatusNoContent != rec.Code {
		t.Fatalf("a bot must be dropped with 204, not errored, got %d", rec.Code)
	}
}

func TestSightingIngestRejectsOversizedBody(t *testing.T) {
	_, router := newSightingRouter(t)

	big := `{"kind":"sail","path":"/","subject":"` + strings.Repeat("a", 2<<10) + `"}`

	if rec := sightingRequest(t, router, "POST", "/1/sighting/", big, "text/plain", ""); http.StatusBadRequest != rec.Code {
		t.Fatalf("expected 400 for an oversized body, got %d", rec.Code)
	}
}

func TestSightingTrafficRequiresAuth(t *testing.T) {
	_, router := newSightingRouter(t)

	if rec := sightingRequest(t, router, "GET", "/1/sighting/traffic", "", "", ""); http.StatusUnauthorized != rec.Code {
		t.Fatalf("expected 401 for an anonymous traffic read, got %d", rec.Code)
	}
}

func TestSightingTrafficReadsWithAuth(t *testing.T) {
	authService, router := newSightingRouter(t)
	token := mintToken(t, authService, time.Now().Add(time.Hour))

	rec := sightingRequest(t, router, "GET", "/1/sighting/traffic?days=7", "", "", token)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected 200 for an authed traffic read, got %d: %s", rec.Code, rec.Body.String())
	}

	var report domain.TrafficReport
	json.Unmarshal(rec.Body.Bytes(), &report)

	if 7 != len(report.Days) {
		t.Fatalf("expected a seven-day series, got %d", len(report.Days))
	}
}
