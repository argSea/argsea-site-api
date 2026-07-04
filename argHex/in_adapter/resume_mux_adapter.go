package in_adapter

import (
	"encoding/json"
	"net/http"

	"github.com/argSea/portfolio_blog_api/argHex/data_objects"
	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/gorilla/mux"
)

//FROM USER TO APP
type resumeMuxAdapter struct {
	resume in_port.ResumeCRUDService
}

func NewResumeMuxAdapter(resume in_port.ResumeCRUDService, m *mux.Router) *resumeMuxAdapter {
	u := resumeMuxAdapter{
		resume: resume,
	}

	m.HandleFunc("/", u.Create).Methods("POST")
	m.HandleFunc("/{id}/", u.Get).Methods("GET")
	m.HandleFunc("/{id}/", u.Update).Methods("PUT")
	m.HandleFunc("/{id}/", u.Delete).Methods("DELETE")

	return &u
}

func (res resumeMuxAdapter) Create(w http.ResponseWriter, r *http.Request) {
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

	var resume domain.Resume
	json.NewDecoder(r.Body).Decode(&resume)

	new_id, err := res.resume.Create(resume)
	var resp interface{}

	if nil != err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: err,
		}
	} else {
		resp = data_objects.NewResumeResponseObject{
			Status:   "ok",
			Code:     200,
			ResumeID: new_id,
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (res resumeMuxAdapter) Get(w http.ResponseWriter, r *http.Request) {
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
	resume_data := res.resume.Read(id)

	response := data_objects.ResumeResponseObject{
		Status: "ok",
		Code:   200,
	}

	response.Resumes = append(response.Resumes, resume_data)

	json.NewEncoder(w).Encode(response)
}

func (res resumeMuxAdapter) Update(w http.ResponseWriter, r *http.Request) {
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

	var resume domain.Resume
	json.NewDecoder(r.Body).Decode(&resume)

	id := mux.Vars(r)["id"]
	resume.Id = id
	updated_err := res.resume.Update(resume)

	var resp interface{}

	if nil != updated_err {
		resp = data_objects.ErroredResponseObject{
			Status:  "error",
			Code:    400,
			Message: updated_err,
		}
	} else {
		resp = data_objects.ItemLessResponseObject{
			Status: "ok",
			Code:   200,
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (res resumeMuxAdapter) Delete(w http.ResponseWriter, r *http.Request) {
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

	resume := domain.Resume{}

	id := mux.Vars(r)["id"]
	resume.Id = id
	deleted_err := res.resume.Delete(resume)

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
