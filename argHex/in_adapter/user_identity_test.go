package in_adapter_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_adapter"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// newIdentityRouter mounts the user adapter over a stored user document, so
// the PUT/DELETE identity gate can be exercised with tokens of both roles.
// Every minted token belongs to userID "keeper".
func newIdentityRouter(t *testing.T, stored domain.User) (*stubUserRepo, in_port.AuthService, *mux.Router) {
	t.Helper()

	repo := &stubUserRepo{
		stored:  stored,
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

	return repo, authService, router
}

func TestUserWriteIdentityMatrix(t *testing.T) {
	// self ok, another user's document 403, admin anywhere ok, anonymous 401 —
	// for both PUT and DELETE
	cases := []struct {
		name   string
		method string
		path   string
		role   string // "" means no token at all
		want   int
	}{
		{"put self", "PUT", "/1/user/keeper", in_port.PERM_USER, http.StatusOK},
		{"put other as user", "PUT", "/1/user/other", in_port.PERM_USER, http.StatusForbidden},
		{"put other as admin", "PUT", "/1/user/other", in_port.PERM_ADMIN, http.StatusOK},
		{"put anonymous", "PUT", "/1/user/keeper", "", http.StatusUnauthorized},
		{"delete self", "DELETE", "/1/user/keeper", in_port.PERM_USER, http.StatusOK},
		{"delete other as user", "DELETE", "/1/user/other", in_port.PERM_USER, http.StatusForbidden},
		{"delete other as admin", "DELETE", "/1/user/other", in_port.PERM_ADMIN, http.StatusOK},
		{"delete anonymous", "DELETE", "/1/user/keeper", "", http.StatusUnauthorized},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo, authService, router := newIdentityRouter(t, domain.User{Id: "keeper", UserName: "meo"})

			token := ""

			if "" != c.role {
				token = mintRoleToken(t, authService, c.role)
			}

			body := ""

			if "PUT" == c.method {
				body = `{"userName":"rewritten"}`
			}

			rec := userRequest(t, router, c.method, c.path, token, body)

			if c.want != rec.Code {
				t.Fatalf("expected %d, got %d (%s)", c.want, rec.Code, rec.Body.String())
			}

			// a refused write must never reach the repo
			if http.StatusOK != c.want && "PUT" == c.method && "" != repo.updated.UserName {
				t.Fatalf("a refused update still reached the repo: %+v", repo.updated)
			}
		})
	}
}

func TestUserReadsAreAuthGated(t *testing.T) {
	_, authService, router := newIdentityRouter(t, domain.User{Id: "keeper", UserName: "meo"})
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	// the full user documents never leave without a token — only /profile is
	// public; any valid token (self or not) may read
	cases := []string{"/1/user", "/1/user/", "/1/user/keeper"}

	for _, path := range cases {
		if rec := userRequest(t, router, "GET", path, "", ""); http.StatusUnauthorized != rec.Code {
			t.Fatalf("expected 401 for anonymous GET %s, got %d", path, rec.Code)
		}

		if rec := userRequest(t, router, "GET", path, token, ""); http.StatusOK != rec.Code {
			t.Fatalf("expected 200 for authed GET %s, got %d", path, rec.Code)
		}
	}
}

func TestUserPutPersistsClearedProfileFields(t *testing.T) {
	repo, authService, router := newIdentityRouter(t, domain.User{
		Id:       "keeper",
		UserName: "meo",
		Password: domain.Password("stored-hash"),
		Role:     in_port.PERM_ADMIN,
		Pronouns: "he/him",
		Linkedin: "in/argsea",
	})

	token := mintRoleToken(t, authService, in_port.PERM_ADMIN)

	// pronouns emptied explicitly, linkedin omitted entirely — both are clears
	// under the PUT's full-replace shape for profile fields
	rec := userRequest(t, router, "PUT", "/1/user/keeper", token, `{"userName":"meo","name":"Justin","pronouns":""}`)

	if http.StatusOK != rec.Code {
		t.Fatalf("expected the update to go through, got %d", rec.Code)
	}

	// the profile keys have no bson omitempty, so the $set document must carry
	// them even when empty — that is what makes the clear persist in mongo
	raw, err := bson.Marshal(*repo.updated)

	if nil != err {
		t.Fatalf("could not bson-marshal the updated user: %v", err)
	}

	var doc bson.M

	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("could not decode the marshaled user: %v", err)
	}

	for _, key := range []string{"pronouns", "linkedin"} {
		value, present := doc[key]

		if !present || "" != value {
			t.Fatalf("expected %q carried as an empty string in the $set document, got %v (present: %v)", key, value, present)
		}
	}

	if "Justin" != doc["name"] {
		t.Fatalf("expected the new name in the $set document, got %v", doc["name"])
	}

	// the credential/role protections must survive the tag change: a body
	// without a password keeps its omitempty, so the stored hash stays put,
	// and the stripped role still vanishes from the $set document
	for _, key := range []string{"password", "role"} {
		if _, present := doc[key]; present {
			t.Fatalf("field %q must stay absent from the $set document, got %+v", key, doc)
		}
	}
}

func TestUserProfilePictureRejectsSvg(t *testing.T) {
	repo, authService, router := newIdentityRouter(t, domain.User{Id: "keeper", UserName: "meo"})

	// a non-admin can self-PUT, and the picture is served static from
	// /media/images/ on the PUBLIC profile page — so a data:image/svg+xml
	// picture is the reachable svg-execution hole and must be refused
	token := mintRoleToken(t, authService, in_port.PERM_USER)

	rec := userRequest(t, router, "PUT", "/1/user/keeper", token,
		`{"userName":"meo","pictures":[{"image":{"src":"data:image/svg+xml;base64,PHN2Zz48L3N2Zz4="}}]}`)

	if http.StatusOK == rec.Code {
		t.Fatalf("expected an svg profile picture to be refused, got 200")
	}

	if "" != repo.updated.UserName {
		t.Fatalf("a refused svg upload must not persist the user: %+v", repo.updated)
	}
}

func TestUserProfileIsPublicAndBare(t *testing.T) {
	_, _, router := newIdentityRouter(t, domain.User{
		Id:       "keeper",
		UserName: "meo",
		Password: domain.Password("hashed-secret"),
		Role:     in_port.PERM_ADMIN,
		Name:     "Justin",
		Pronouns: "he/him",
		Location: "the harbor",
		Title:    "keeper of the light",
		Bio:      "keeps the lantern lit",
		Email:    "keeper@argsea.com",
		Github:   "argSea",
		Linkedin: "in/argsea",
		Signoff:  "— J",
	})

	// no token at all — the profile read is the one public user endpoint
	rec := userRequest(t, router, "GET", "/1/user/keeper/profile", "", "")

	if http.StatusOK != rec.Code {
		t.Fatalf("expected a public 200, got %d", rec.Code)
	}

	var payload map[string]interface{}

	if err := json.Unmarshal(rec.Body.Bytes(), &payload); nil != err {
		t.Fatalf("could not parse the profile: %v", err)
	}

	// exactly the nine profile fields — credentials and role must never leak
	if 9 != len(payload) {
		t.Fatalf("expected exactly 9 profile fields, got %d: %+v", len(payload), payload)
	}

	for _, key := range []string{"name", "pronouns", "location", "title", "bio", "email", "github", "linkedin", "signoff"} {
		if _, present := payload[key]; !present {
			t.Fatalf("expected profile field %q, got %+v", key, payload)
		}
	}

	for _, key := range []string{"userName", "password", "role", "id"} {
		if _, present := payload[key]; present {
			t.Fatalf("field %q must never leak through the public profile", key)
		}
	}

	if "Justin" != payload["name"] || "keeper of the light" != payload["title"] {
		t.Fatalf("profile values did not round-trip: %+v", payload)
	}
}

func TestUserProfileUnknownUserIs404(t *testing.T) {
	_, _, router := newIdentityRouter(t, domain.User{})

	if rec := userRequest(t, router, "GET", "/1/user/ghost/profile", "", ""); http.StatusNotFound != rec.Code {
		t.Fatalf("expected 404 for an unknown user, got %d", rec.Code)
	}
}
