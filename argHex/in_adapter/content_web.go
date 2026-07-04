package in_adapter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
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

// requireAuth gates a write endpoint. It writes the appropriate error response
// and returns false when the caller is not authorized, so a handler can early
// return on `if !requireAuth(w, r) { return }`.
func requireAuth(w http.ResponseWriter, r *http.Request) bool {
	authorized, err := checkAuth(r)

	if nil != err {
		writeError(w, 500, err.Error())
		return false
	}

	if !authorized {
		writeError(w, 401, "Unauthorized")
		return false
	}

	return true
}

// checkAuth validates the caller's session cookie against the auth service's
// validate endpoint. This mirrors the existing project/user adapters — the
// validate route owns JWT verification, adapters just ask it.
func checkAuth(r *http.Request) (bool, error) {
	validate_endpoint := "https://api.argsea.com/1/auth/validate/"

	cookies := r.Cookies()
	cookie_string := ""

	for i := 0; i < len(cookies); i++ {
		cookie_string += cookies[i].Name + "=" + cookies[i].Value + ";"
	}

	req, req_err := http.NewRequest("GET", validate_endpoint, nil)

	if nil != req_err {
		return false, req_err
	}

	req.Header.Add("Cookie", cookie_string)

	val_res, val_err := http.DefaultClient.Do(req)

	if nil != val_err {
		return false, val_err
	}

	defer val_res.Body.Close()

	val_body, val_body_err := ioutil.ReadAll(val_res.Body)

	if nil != val_body_err {
		return false, val_body_err
	}

	var val_data map[string]interface{}
	json.Unmarshal(val_body, &val_data)

	if "ok" != val_data["status"] {
		return false, nil
	}

	return true, nil
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
