package in_adapter

import (
	"encoding/base64"
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

type projectMuxAdatper struct {
	project in_port.ProjectCRUDService
	media   in_port.MediaService
}

func NewProjectMuxAdapter(proj in_port.ProjectCRUDService, m *mux.Router, media in_port.MediaService) *projectMuxAdatper {
	p := projectMuxAdatper{
		project: proj,
		media:   media,
	}

	m.HandleFunc("", p.GetAll).Methods("GET")
	m.HandleFunc("", p.Create).Methods("POST")
	m.HandleFunc("/{id}", p.Get).Methods("GET")
	m.HandleFunc("/{id}", p.Update).Methods("PUT")
	m.HandleFunc("/{id}", p.Delete).Methods("DELETE")

	m.HandleFunc("/", p.GetAll).Methods("GET")
	m.HandleFunc("/", p.Create).Methods("POST")
	m.HandleFunc("/{id}/", p.Get).Methods("GET")
	m.HandleFunc("/{id}/", p.Update).Methods("PUT")
	m.HandleFunc("/{id}/", p.Delete).Methods("DELETE")

	return &p
}

func (p projectMuxAdatper) Create(w http.ResponseWriter, r *http.Request) {
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

	var project domain.Project
	json.NewDecoder(r.Body).Decode(&project)

	// check auth
	authorized, auth_err := p.checkAuth(r)

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

	// upload project.icon.src
	if "" != project.Icon.Source {
		// check if icon is file data or url
		if "data:" == project.Icon.Source[:5] {
			// upload file
			mime_type := project.Icon.Source[5:strings.Index(project.Icon.Source, ";")]
			encoded_data := project.Icon.Source[strings.Index(project.Icon.Source, ",")+1:]

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
			upload_res, upload_err := p.media.UploadMedia(mime_type, decoded_data)

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

			project.Icon.Source = upload_res
		}
	}

	// do the same with project.Images
	for i := 0; i < len(project.Images); i++ {
		if "" != project.Images[i].Image.Source {
			// check if icon is file data or url
			if "data:" == project.Images[i].Image.Source[:5] {
				// upload file
				mime_type := project.Images[i].Image.Source[5:strings.Index(project.Images[i].Image.Source, ";")]
				encoded_data := project.Images[i].Image.Source[strings.Index(project.Images[i].Image.Source, ",")+1:]

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
				upload_res, upload_err := p.media.UploadMedia(mime_type, decoded_data)

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

				project.Images[i].Image.Source = upload_res
			}
		}
	}

	new_id, err := p.project.Create(project)
	new_project := p.project.Read(new_id)
	var resp interface{}

	if nil != err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: err,
		}
	} else {
		resp = data_objects.NewProjectResponseObject{
			Status:    "ok",
			Code:      200,
			ProjectID: new_id,
		}

		resp = new_project
	}

	json.NewEncoder(w).Encode(resp)
}

func (p projectMuxAdatper) Get(w http.ResponseWriter, r *http.Request) {
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

	id := mux.Vars(r)["id"]
	project_data := p.project.Read(id)

	response := data_objects.ProjectResponseObject{
		Status: "ok",
		Code:   200,
	}

	response.Projects = append(response.Projects, project_data)

	json.NewEncoder(w).Encode(response.Projects[0])
}

func (p projectMuxAdatper) GetAll(w http.ResponseWriter, r *http.Request) {
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

	// look for filter param and decode it
	query := r.URL.Query().Get("filter")

	// log query
	log.Println(query)

	if query == "" {
		// throw 404
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    404,
			Message: "No filter provided",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var filter map[string]string
	json.Unmarshal([]byte(query), &filter)

	// check for userID in filter
	userID, ok := filter["userID"]

	if !ok {
		// throw 404
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    404,
			Message: "No userID provided",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	projects, count, err := p.project.GetByUserID(userID)

	if nil != err {
		// throw 404
		response := data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    404,
			Message: err,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := data_objects.ProjectResponseObject{
		Status: "ok",
		Code:   200,
		Count:  count,
	}

	for i := 0; i < len(projects); i++ {
		response.Projects = append(response.Projects, projects[i])
	}

	total := len(response.Projects)

	// w.Header().Add("Content-Range", "users "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+limit, 10)+"/"+strconv.FormatInt(int64(total), 10))
	// w.Header().Add("range", "users "+strconv.FormatInt(offset, 10)+"-"+strconv.FormatInt(offset+limit, 10)+"/"+strconv.FormatInt(int64(total), 10))
	w.Header().Add("X-Total-Count", strconv.FormatInt(int64(total), 10))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.Projects)

}

func (p projectMuxAdatper) Update(w http.ResponseWriter, r *http.Request) {
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

	var project domain.Project
	json.NewDecoder(r.Body).Decode(&project)

	id := mux.Vars(r)["id"]
	project.Id = id

	// check auth
	authorized, auth_err := p.checkAuth(r)

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

	// upload project.icon.src
	if "" != project.Icon.Source {
		// check if icon is file data or url
		if "data:" == project.Icon.Source[:5] {
			// upload file
			mime_type := project.Icon.Source[5:strings.Index(project.Icon.Source, ";")]
			encoded_data := project.Icon.Source[strings.Index(project.Icon.Source, ",")+1:]

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
			upload_res, upload_err := p.media.UploadMedia(mime_type, decoded_data)

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

			project.Icon.Source = upload_res
		}
	}

	// do the same with project.Images
	for i := 0; i < len(project.Images); i++ {
		if "" != project.Images[i].Image.Source {
			// check if icon is file data or url
			if "data:" == project.Images[i].Image.Source[:5] {
				// upload file
				mime_type := project.Images[i].Image.Source[5:strings.Index(project.Images[i].Image.Source, ";")]
				encoded_data := project.Images[i].Image.Source[strings.Index(project.Images[i].Image.Source, ",")+1:]

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
				upload_res, upload_err := p.media.UploadMedia(mime_type, decoded_data)

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

				project.Images[i].Image.Source = upload_res
			}
		}
	}

	updated_err := p.project.Update(project)

	// get updated project
	project = p.project.Read(id)

	var resp interface{}

	if nil != updated_err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: updated_err,
		}

		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp = data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}
		resp = project

		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

func (p projectMuxAdatper) Delete(w http.ResponseWriter, r *http.Request) {
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

	project := domain.Project{}

	authorized, auth_err := p.checkAuth(r)

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

	id := mux.Vars(r)["id"]
	project.Id = id
	deleted_err := p.project.Delete(project)

	var resp interface{}

	if nil != deleted_err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: deleted_err,
		}
	} else {
		resp = data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (p projectMuxAdatper) checkAuth(r *http.Request) (bool, error) {
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
