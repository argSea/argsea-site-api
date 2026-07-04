package in_adapter

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/argSea/portfolio_blog_api/argHex/data_objects"
	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/gorilla/mux"
)

type skillMuxAdapter struct {
	skillService in_port.SkillCRUDService
}

func NewSkillMuxAdapter(s in_port.SkillCRUDService, r *mux.Router) {
	adapter := skillMuxAdapter{
		skillService: s,
	}

	// create skill
	r.HandleFunc("", adapter.Create).Methods("POST")
	r.HandleFunc("/", adapter.Create).Methods("POST")

	// get all skills
	r.HandleFunc("", adapter.GetAll).Methods("GET")
	r.HandleFunc("/", adapter.GetAll).Methods("GET")

	// get skill by id
	r.HandleFunc("/{id}", adapter.Get).Methods("GET")
	r.HandleFunc("/{id}/", adapter.Get).Methods("GET")

	// update skill
	r.HandleFunc("/{id}", adapter.Update).Methods("PUT")
	r.HandleFunc("/{id}/", adapter.Update).Methods("PUT")

	// delete skill
	r.HandleFunc("/{id}", adapter.Delete).Methods("DELETE")
	r.HandleFunc("/{id}/", adapter.Delete).Methods("DELETE")
}

func (s skillMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
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

	authorized, auth_err := s.checkAuth(r)

	if nil != auth_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: auth_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	if !authorized {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)

		return
	}

	var skill domain.Skill
	json.NewDecoder(r.Body).Decode(&skill)

	skill_id, err := s.skillService.Create(skill)

	if nil != err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	// get skill by id
	skill = s.skillService.Read(skill_id)
	var resp interface{}

	// check if skill is valid
	if "" == skill.Id {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: "Skill not found",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	resp = data_objects.NewSkillResponseObject{
		Status:  "ok",
		Code:    200,
		SkillID: skill.Id,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s skillMuxAdapter) GetAll(w http.ResponseWriter, r *http.Request) {
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

	limit := int64(0)
	offset := int64(0)
	sort := ""

	// no auth for GET requests
	// authorized, auth_err := s.checkAuth(r)

	// if nil != auth_err {
	// 	response := data_objects.ErroredResponseObject{
	// 		Status:  "error",
	// 		Code:    500,
	// 		Message: auth_err.Error(),
	// 	}
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	json.NewEncoder(w).Encode(response)

	// 	return
	// }

	// if !authorized {
	// 	response := data_objects.ErroredResponseObject{
	// 		Status:  "error",
	// 		Code:    401,
	// 		Message: "Unauthorized",
	// 	}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(response)

	// 	return
	// }

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

	skills := s.skillService.ReadAll(limit, offset, sort)

	if 0 == len(skills) {
		response := data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		return
	}

	resp := data_objects.SkillResponseObject{
		Status: "ok",
		Code:   200,
		Count:  int64(len(skills)),
	}

	for i := 0; i < len(skills); i++ {
		resp.Skills = append(resp.Skills, skills[i])
	}

	total := len(resp.Skills)

	w.Header().Add("X-Total-Count", strconv.FormatInt(int64(total), 10))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp.Skills)
}

func (s skillMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
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

	// no auth for GET requests
	// skip auth for GET requests

	var skill domain.Skill
	id := mux.Vars(r)["id"]

	skill = s.skillService.Read(id)

	if "" == skill.Id {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: "Skill not found",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	resp := data_objects.SkillResponseObject{
		Status: "ok",
		Code:   200,
		Count:  1,
	}

	resp.Skills = append(resp.Skills, skill)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp.Skills[0])
}

func (s skillMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
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

	// check for json errors in request body
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
			Code:    500,
			Message: "Empty body",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	log.Println(string(body))

	// parse body into skill
	skill := domain.Skill{}
	json_err := json.Unmarshal(body, &skill)

	if nil != json_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: json_err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	log.Println(skill)

	id := mux.Vars(r)["id"]
	skill.Id = id

	// check auth
	authorized, auth_err := s.checkAuth(r)

	if nil != auth_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: auth_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	if !authorized {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)

		return
	}

	err := s.skillService.Update(skill)

	if nil != err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	response := data_objects.ItemLessResponseObject{
		Status: "ok",
		Code:   200,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s skillMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
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

	// check auth
	authorized, auth_err := s.checkAuth(r)

	if nil != auth_err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: auth_err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)

		return
	}

	if !authorized {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    401,
			Message: "Unauthorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)

		return
	}

	skill := domain.Skill{}
	id := mux.Vars(r)["id"]
	skill.Id = id

	err := s.skillService.Delete(skill)

	if nil != err {
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    500,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}

	response := data_objects.ItemLessResponseObject{
		Status: "ok",
		Code:   200,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s skillMuxAdapter) checkAuth(r *http.Request) (bool, error) {
	// check auth
	validate_endpoint := "https://api.argsea.com/1/auth/validate/"

	// pass along all cookies
	cookies := r.Cookies()
	cookie_string := ""

	for i := 0; i < len(cookies); i++ {
		cookie_string += cookies[i].Name + "=" + cookies[i].Value + ";"
	}

	log.Println(cookie_string)

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
