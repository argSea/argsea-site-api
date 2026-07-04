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

	user, err := a.loginService.Login(user)

	if nil != err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
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

	// pull the token off the request — session cookie or bearer header, the
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
	roles := []string{"user"}
	token, auth_error := a.authService.Generate(user.Id, expires, roles)

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
