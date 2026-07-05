package in_adapter

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/data_objects"
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	auth "github.com/argSea/argsea-site-api/argHex/utility"
	"github.com/gorilla/mux"
)

// FROM USER TO APP
type userMuxAdapter struct {
	user  in_port.UserCRUDService
	media in_port.MediaService
	auth  *WebAuth
}

func NewUserMuxAdapter(u in_port.UserCRUDService, m in_port.MediaService, auth *WebAuth, router *mux.Router) {
	adapter := &userMuxAdapter{
		user:  u,
		media: m,
		auth:  auth,
	}

	//user service
	router.HandleFunc("", adapter.GetAll).Methods("GET")
	router.HandleFunc("", adapter.Create).Methods("POST")
	router.HandleFunc("/{id}", adapter.Get).Methods("GET")
	router.HandleFunc("/{id}", adapter.Update).Methods("PUT")
	router.HandleFunc("/{id}", adapter.Delete).Methods("DELETE")

	router.HandleFunc("/", adapter.GetAll).Methods("GET")
	router.HandleFunc("/", adapter.Create).Methods("POST")
	router.HandleFunc("/{id}/", adapter.Get).Methods("GET")
	router.HandleFunc("/{id}/", adapter.Update).Methods("PUT")
	router.HandleFunc("/{id}/", adapter.Delete).Methods("DELETE")

	// public keeper profile — the one unauthenticated user read; the site build
	// and the admin greeting both consume it
	router.HandleFunc("/{id}/profile", adapter.Profile).Methods("GET")
	router.HandleFunc("/{id}/profile/", adapter.Profile).Methods("GET")
}

// Profile serves the bare public keeper subset — the nine profile fields only,
// never username, password, or role. Unknown users 404.
func (u userMuxAdapter) Profile(w http.ResponseWriter, r *http.Request) {
	user_data := u.user.Read(mux.Vars(r)["id"])

	if "" == user_data.Id {
		writeError(w, 404, "Not found")
		return
	}

	writeJSON(w, http.StatusOK, user_data.Profile())
}

func (u userMuxAdapter) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	// the full user documents are keeper-only — the public read is /profile
	if !requireAuth(u.auth, w, r) {
		return
	}

	limit := int64(0)
	offset := int64(0)
	sort := ""

	if nil != r.URL.Query()["limit"] {
		// convert string to int64
		i, ierr := strconv.ParseInt(r.URL.Query()["limit"][0], 10, 64)

		if nil != ierr {
			// do nothing
		} else {
			limit = i
		}
	}

	if nil != r.URL.Query()["offset"] {
		// convert string to int64
		i, ierr := strconv.ParseInt(r.URL.Query()["offset"][0], 10, 64)

		if nil != ierr {
			// do nothing
		} else {
			offset = i
		}
	}

	if nil != r.URL.Query()["sort"] {
		sort = r.URL.Query()["sort"][0]

		if "" == sort {
			sort = "nil"
		}
	}

	// if limit and offset are 0, check for range query string
	if 0 == limit && 0 == offset {
		if nil != r.URL.Query()["range"] {
			// convert [0, 10] to limit = 10, offset = 0
			range_str := r.URL.Query()["range"][0]
			range_str = strings.Replace(range_str, "[", "", -1)
			range_str = strings.Replace(range_str, "]", "", -1)

			range_arr := strings.Split(range_str, ",")
			limit, _ = strconv.ParseInt(range_arr[1], 10, 64)
			offset, _ = strconv.ParseInt(range_arr[0], 10, 64)
		}
	}

	users := u.user.ReadAll(limit, offset, sort)

	response := data_objects.UserResponseObject{
		Status: "ok",
		Code:   200,
	}

	for i := 0; i < len(users); i++ {
		response.Users = append(response.Users, users[i])
	}

	// set Content-Range header with limit, offset, and total
	total := len(response.Users)
	// response.Count = int64(total)

	// w.Header().Add("Content-Range", "users "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+limit, 10)+"/"+strconv.FormatInt(int64(total), 10))
	// w.Header().Add("range", "users "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+limit, 10)+"/"+strconv.FormatInt(int64(total), 10))
	w.Header().Add("X-Total-Count", strconv.FormatInt(int64(total), 10))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.Users)
}

func (u userMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	if !requireAuth(u.auth, w, r) {
		return
	}

	var user domain.User
	json.NewDecoder(r.Body).Decode(&user)

	// role never comes from the request body — admin is granted only by a
	// direct DB update
	user.Role = ""

	new_id, err := u.user.Create(user)

	// get user by new_id
	user = u.user.Read(new_id)
	var resp interface{}

	if nil != err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp = data_objects.NewUserResponseObject{
			Status: "ok",
			Code:   200,
			UserID: new_id,
		}

		resp = user

		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

func (u userMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			json.NewEncoder(w).Encode(response)
		}
	}()

	// the full user document stays gated — anonymous callers get the bare
	// profile subset through /profile, never userName or role
	if !requireAuth(u.auth, w, r) {
		return
	}

	id := mux.Vars(r)["id"]
	user_data := u.user.Read(id)

	response := data_objects.UserResponseObject{
		Status: "ok",
		Code:   200,
	}

	response.Users = append(response.Users, user_data)

	json.NewEncoder(w).Encode(response.Users[0])
}

func (u userMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	// check for json errors in r.body
	body, body_err := ioutil.ReadAll(r.Body)

	if nil != body_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: body_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	// check for empty body
	if "" == string(body) {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: "Empty body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	log.Println(string(body))

	// parse body into user
	user := domain.User{}
	json_err := json.Unmarshal(body, &user)

	if nil != json_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: json_err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	log.Println(user)

	id := mux.Vars(r)["id"]
	user.Id = id

	// strip any role in the body: with role empty (bson omitempty) the $set
	// update leaves the stored role untouched, so a PUT cannot self-grant admin
	user.Role = ""

	// a valid token alone is not enough — it must belong to the user being
	// rewritten, or carry the admin role
	if !requireSelfOrAdmin(u.auth, w, r, id) {
		return
	}

	if "" != user.Password {
		// hash password
		hashed_pass, pass_err := auth.HashPassword(string(user.Password))

		if nil != pass_err {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: pass_err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)

			return
		}

		user.Password = domain.Password(hashed_pass)
	}

	// upload all user.Pictures
	for i := 0; i < len(user.Pictures); i++ {
		this_picture := user.Pictures[i].Image
		if "" == this_picture.Source {
			continue
		}
		// check if icon is file data or url
		if "data:" == this_picture.Source[:5] {
			// upload file
			mime_type := this_picture.Source[5:strings.Index(this_picture.Source, ";")]
			encoded_data := this_picture.Source[strings.Index(this_picture.Source, ",")+1:]

			decoded_data, decode_err := base64.StdEncoding.DecodeString(encoded_data)

			if nil != decode_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: decode_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			// upload file
			upload_res, upload_err := u.media.UploadMedia(mime_type, decoded_data)

			if nil != upload_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: upload_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			this_picture.Source = upload_res
		}

		user.Pictures[i].Image = this_picture
	}

	for i := 0; i < len(user.Contacts); i++ {
		// check if icon is file data or url
		if "" == user.Contacts[i].Icon.Source {
			continue
		}

		if "data:" == user.Contacts[i].Icon.Source[:5] {
			// upload file
			mime_type := user.Contacts[i].Icon.Source[5:strings.Index(user.Contacts[i].Icon.Source, ";")]
			encoded_data := user.Contacts[i].Icon.Source[strings.Index(user.Contacts[i].Icon.Source, ",")+1:]

			decoded_data, decode_err := base64.StdEncoding.DecodeString(encoded_data)

			if nil != decode_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: decode_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			// upload file
			upload_res, upload_err := u.media.UploadMedia(mime_type, decoded_data)

			if nil != upload_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: upload_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			user.Contacts[i].Icon.Source = upload_res
		}
	}

	// do the same for user.Interests
	for i := 0; i < len(user.TechInterests); i++ {
		// check if icon is file data or url
		if "" == user.TechInterests[i].Icon.Source {
			continue
		}

		if "data:" == user.TechInterests[i].Icon.Source[:5] {
			// upload file
			mime_type := user.TechInterests[i].Icon.Source[5:strings.Index(user.TechInterests[i].Icon.Source, ";")]
			encoded_data := user.TechInterests[i].Icon.Source[strings.Index(user.TechInterests[i].Icon.Source, ",")+1:]

			decoded_data, decode_err := base64.StdEncoding.DecodeString(encoded_data)

			if nil != decode_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: decode_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			// upload file
			upload_res, upload_err := u.media.UploadMedia(mime_type, decoded_data)

			if nil != upload_err {
				response := data_objects.ErroredResponseObject{
					Status:  "error",
					Code:    500,
					Message: upload_err.Error(),
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)

				return
			}

			user.TechInterests[i].Icon.Source = upload_res
		}
	}

	updated_err := u.user.Update(user)

	// get updated user
	user = u.user.Read(id)

	var resp interface{}

	if nil != updated_err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: updated_err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp = data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}

		resp = user
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

func (u userMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			response := data_objects.ErroredResponseObject{
				Status:  "error",
				Code:    500,
				Message: err,
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
		}
	}()

	user := domain.User{}

	id := mux.Vars(r)["id"]
	user.Id = id

	// same identity rule as Update: only the user themself or an admin may
	// delete a user document
	if !requireSelfOrAdmin(u.auth, w, r, id) {
		return
	}

	deleted_err := u.user.Delete(user)

	var resp interface{}

	if nil != deleted_err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: deleted_err,
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp = data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}
