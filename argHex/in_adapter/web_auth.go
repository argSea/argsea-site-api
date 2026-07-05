package in_adapter

import (
	"net/http"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/sessions"
)

const authCookieName = "auth-token"

// WebAuth is the single request-authentication mechanism for every adapter:
// it extracts the JWT from the request and validates it in-process through
// in_port.AuthService — no HTTP round-trips to a validate endpoint. The auth
// adapter shares the same cookie store for issuing sessions on login, so there
// is exactly one place that knows how a token travels.
type WebAuth struct {
	auth  in_port.AuthService
	store *sessions.CookieStore
}

func NewWebAuth(auth in_port.AuthService, secret []byte, cookieDomain string) *WebAuth {
	store := sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Domain:   cookieDomain,
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
	}

	return &WebAuth{
		auth:  auth,
		store: store,
	}
}

// Store exposes the shared cookie store so the auth adapter can save and clear
// sessions with the same encoding the extractor reads.
func (w *WebAuth) Store() *sessions.CookieStore {
	return w.store
}

// Token pulls the JWT off the request: the auth-token session cookie first,
// falling back to an Authorization: Bearer header (handy for curl and any
// client that doesn't hold the cookie). Empty string means no token.
func (w *WebAuth) Token(r *http.Request) string {
	session, err := w.store.Get(r, authCookieName)

	if nil == err && !session.IsNew {
		if token, ok := session.Values["token"].(string); ok && "" != token {
			return token
		}
	}

	header := r.Header.Get("Authorization")

	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}

	return ""
}

// Authorized reports whether the request carries a valid JWT.
func (w *WebAuth) Authorized(r *http.Request) bool {
	_, authorized := w.Role(r)

	return authorized
}

// Role returns the role claim of the request's validated JWT. authorized is
// false when the request carries no token or the token doesn't validate — in
// that case the role is meaningless and returned empty.
func (w *WebAuth) Role(r *http.Request) (string, bool) {
	claims, authorized := w.Claims(r)

	return claims.Role, authorized
}

// Claims returns the validated claims of the request's JWT (role + userID),
// for gates that need to know who the token belongs to, not just that it is
// valid. authorized is false when the request carries no token or the token
// doesn't validate — the claims are meaningless then.
func (w *WebAuth) Claims(r *http.Request) (data_objects.AuthValidationResponseObject, bool) {
	token := w.Token(r)

	if "" == token {
		return data_objects.AuthValidationResponseObject{}, false
	}

	validation, err := w.auth.Validate(token)

	if nil != err || !validation.Valid {
		return data_objects.AuthValidationResponseObject{}, false
	}

	return validation, true
}
