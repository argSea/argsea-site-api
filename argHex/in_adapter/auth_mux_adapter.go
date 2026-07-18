package in_adapter

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// consoleHeader is the marker the admin console sets on its own login request.
// Its presence is how a console hail is told apart from a direct one: present
// gets the JSON error its SPA reads to launch the flood, absent gets sent adrift.
const consoleHeader = "X-Argsea-Console"

// adriftPath is where a refused direct hail is sent: a route that 302s back to
// itself, an endless drift for anything that follows redirects. It is the mount
// point of the trap route under the /1/auth prefix.
const adriftPath = "/1/auth/adrift/"

type authMuxAdapter struct {
	authService  in_port.AuthService
	loginService in_port.UserLoginService
	webAuth      *WebAuth
	store        *sessions.CookieStore
}

// NewAuthMuxAdapter wires the login/logout/validate routes. It issues sessions
// through the shared WebAuth cookie store, so the tokens it writes are read
// back by the same mechanism every other adapter authenticates with.
func NewAuthMuxAdapter(a in_port.AuthService, l in_port.UserLoginService, webAuth *WebAuth, r *mux.Router) {
	adapter := authMuxAdapter{
		authService:  a,
		loginService: l,
		webAuth:      webAuth,
		store:        webAuth.Store(),
	}

	//user auth service
	r.HandleFunc("/login/", adapter.Login).Methods("POST")
	r.HandleFunc("/logout/", adapter.Logout).Methods("GET")
	r.HandleFunc("/validate/", adapter.Validate).Methods("GET")
	r.HandleFunc("/adrift/", adapter.Adrift).Methods("GET")

}

func (a authMuxAdapter) Logout(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			// set code 500
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	// check if auth-token cookie exists
	session, session_err := a.store.Get(r, "auth-token")

	if nil != session_err {
		log.Println("Error getting session: ", session_err)
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: session_err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("Session data: ", session)

	if session.IsNew {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("Session data: ", session.Options)
	// delete session
	session.Options.MaxAge = 1
	session.Values = nil
	s_err := session.Save(r, w)

	if nil != s_err {
		log.Println("Error saving session: ", s_err)
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: s_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("Session data: ", session.Options)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Status  string `json:"status"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Status:  "ok",
		Code:    200,
		Message: "User logged out",
	})
}

func (a authMuxAdapter) Login(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			// set code 500
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	var user domain.User
	json.NewDecoder(r.Body).Decode(&user)

	user, err := a.loginService.Login(user, clientIP(r))

	if nil != err {
		a.refuseHail(w, r, err)
		return
	}

	token, token_error := a.setSession(user, w, r)

	if nil != token_error {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: token_error.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data_objects.LoginResponseObject{
		Status:   "ok",
		Code:     200,
		UserName: user.UserName,
		UserID:   user.Id,
		Token:    token,
	})
}

// refuseHail answers a failed login. The console marks its request, so its SPA
// gets the 400 JSON it reads to launch the flood (the barred line when barred,
// the generic credentials line otherwise, same status either way). A direct hail
// carries no marker and gets no answer worth reading: a 302 into a drift that
// loops back on itself forever.
func (a authMuxAdapter) refuseHail(w http.ResponseWriter, r *http.Request, err error) {
	if "" != r.Header.Get(consoleHeader) {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Redirect(w, r, adriftPath, http.StatusFound)
}

// Adrift answers the trap route with a 302 back to itself, an endless drift for
// anything that follows redirects. curl -L loops to its cap; a browser reports a
// redirect loop. There is no body worth reading.
func (a authMuxAdapter) Adrift(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, adriftPath, http.StatusFound)
}

func (a authMuxAdapter) Validate(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			// set code 500
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	// pull the token off the request; session cookie or bearer header, the
	// same extraction every other adapter uses
	token := a.webAuth.Token(r)

	if "" == token {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// check auth
	v_response, v_err := a.authService.Validate(token)

	if nil != v_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: v_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !v_response.Valid {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Println("User is authorized! " + v_response.UserID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Status  string `json:"status"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Status:  "ok",
		Code:    200,
		Message: "User is authorized",
	})
}

func (a authMuxAdapter) setSession(user domain.User, w http.ResponseWriter, r *http.Request) (string, error) {
	expires := time.Now().Add(time.Hour * 24)

	// mint the role stored on the user document; the doc is trustworthy because
	// role never enters through a request body, only a direct DB update
	role := in_port.PERM_USER

	if in_port.PERM_ADMIN == user.Role {
		role = in_port.PERM_ADMIN
	}

	token, auth_error := a.authService.Generate(user.Id, expires, []string{role})

	if nil != auth_error {
		return "", auth_error
	}

	session, session_err := a.store.Get(r, "auth-token")

	if nil != session_err {
		return "", session_err
	}

	session.Values["token"] = token
	session.Values["iat"] = time.Now().Unix()
	session.Save(r, w)
	log.Println("Cookie set: ", session)

	return token, nil
}
