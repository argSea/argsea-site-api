package in_adapter_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
)

// newMediaRouter mounts the media adapter over a temp-dir webstore and fake
// metadata, behind the real JWT gate.
func newMediaRouter(t *testing.T) *mux.Router {
	t.Helper()

	authService := service.NewJWTAuthService(testSecret)
	webAuth := in_adapter.NewWebAuth(authService, testSecret, "argsea.com")

	mediaService := service.NewMediaService(
		out_adapter.NewMediaWebstoreAdapter(t.TempDir()+string(filepath.Separator), "/media/images"),
		out_adapter.NewMediaMetaFakeOutAdapter(),
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	router := mux.NewRouter()
	in_adapter.NewMediaMuxAdapter(mediaService, webAuth, router.PathPrefix("/1/media").Subrouter())

	return router
}

func TestMediaRoutesAreAuthGated(t *testing.T) {
	router := newMediaRouter(t)

	// the whole darkroom is keeper-only; every route bounces an anonymous call
	cases := []struct {
		method string
		path   string
	}{
		{"GET", "/1/media"},
		{"GET", "/1/media/"},
		{"POST", "/1/media/"},
		{"DELETE", "/1/media/some-id"},
	}

	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if http.StatusUnauthorized != rec.Code {
			t.Fatalf("expected 401 for anonymous %s %s, got %d", c.method, c.path, rec.Code)
		}
	}
}
