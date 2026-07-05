package in_adapter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/in_port"
)

// Shared plumbing for the content in-adapters (projects, notes, hobbies, copy,
// suggestions, activity). Keeps each handler focused on its resource.

// writeJSON writes a status code and a JSON body. Content-Type is already set by
// the base middleware.
func writeJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// writeError writes the standard errored response envelope.
func writeError(w http.ResponseWriter, code int64, message interface{}) {
	w.WriteHeader(int(code))
	json.NewEncoder(w).Encode(data_objects.ErroredResponseObject{
		Status:  "error",
		Code:    code,
		Message: message,
	})
}

// requireAuth gates a write endpoint through the shared in-process validator.
// It writes the 401 response and returns false when the caller is not
// authorized, so a handler can early return on `if !requireAuth(auth, w, r)`.
func requireAuth(auth *WebAuth, w http.ResponseWriter, r *http.Request) bool {
	if !auth.Authorized(r) {
		writeError(w, 401, "Unauthorized")
		return false
	}

	return true
}

// requireAdmin gates an admin-only endpoint: 401 when the request carries no
// valid token at all, 403 when the token is valid but not admin-role. Returns
// false when the caller may not proceed, mirroring requireAuth.
func requireAdmin(auth *WebAuth, w http.ResponseWriter, r *http.Request) bool {
	role, authorized := auth.Role(r)

	if !authorized {
		writeError(w, 401, "Unauthorized")
		return false
	}

	if in_port.PERM_ADMIN != role {
		writeError(w, 403, "Forbidden")
		return false
	}

	return true
}

// requireSelfOrAdmin gates a per-user write: 401 when the request carries no
// valid token, 403 when the token is valid but belongs to neither the target
// user nor an admin. Returns false when the caller may not proceed, mirroring
// requireAuth.
func requireSelfOrAdmin(auth *WebAuth, w http.ResponseWriter, r *http.Request, id string) bool {
	claims, authorized := auth.Claims(r)

	if !authorized {
		writeError(w, 401, "Unauthorized")
		return false
	}

	if claims.UserID != id && in_port.PERM_ADMIN != claims.Role {
		writeError(w, 403, "Forbidden")
		return false
	}

	return true
}

// queryLimit reads an optional ?limit= integer, defaulting to fallback when it
// is absent or unparseable.
func queryLimit(r *http.Request, fallback int64) int64 {
	raw := r.URL.Query().Get("limit")

	if "" == raw {
		return fallback
	}

	limit, err := strconv.ParseInt(raw, 10, 64)

	if nil != err {
		return fallback
	}

	return limit
}

// queryFlag reports whether a query parameter is present and set to "true".
func queryFlag(r *http.Request, key string) bool {
	return "true" == r.URL.Query().Get(key)
}
