package data_objects

//general - outward
// type BaseResume struct {
// 	Id            string      `json:"resumeID"`
// 	UserID        string      `json:"userID"`
// 	About         string      `json:"about"`
// 	Experiences   interface{} `json:"experiences"`
// 	Education     interface{} `json:"education"`
// 	ExtraCourses  interface{} `json:"extraCourses"`
// 	SkillSections interface{} `json:"skills"`
// }

//general - inward

//web
////general
type ErroredResponseObject struct {
	Status  string      `json:"status"`
	Code    int64       `json:"code"`
	Message interface{} `json:"message"`
}

type ItemLessResponseObject struct {
	Status string `json:"status"`
	Code   int64  `json:"code"`
}

////resume
type ResumeResponseObject struct {
	Status  string        `json:"status"`
	Code    int64         `json:"code"`
	Count   int64         `json:"count"`
	Resumes []interface{} `json:"resumes"`
}

type NewResumeResponseObject struct {
	Status   string `json:"status"`
	Code     int64  `json:"code"`
	ResumeID string `json:"resumeID"`
}

////user
type UserResponseObject struct {
	Status string        `json:"status"`
	Code   int64         `json:"code"`
	Count  int64         `json:"count"`
	Users  []interface{} `json:"users"`
}

type LoginResponseObject struct {
	Status   string `json:"status"`
	Code     int64  `json:"code"`
	UserName string `json:"userName"`
	UserID   string `json:"userID"`
	Token    string `json:"token"`
}

type NewUserResponseObject struct {
	Status string `json:"status"`
	Code   int64  `json:"code"`
	UserID string `json:"userID"`
}

//projects
type ProjectResponseObject struct {
	Status   string        `json:"status"`
	Code     int64         `json:"code"`
	Count    int64         `json:"count"`
	Projects []interface{} `json:"projects"`
}

type NewProjectResponseObject struct {
	Status    string `json:"status"`
	Code      int64  `json:"code"`
	ProjectID string `json:"projectID"`
}

type AuthValidationResponseObject struct {
	Valid  bool   `json:"valid"`
	Role   string `json:"roles"`
	UserID string `json:"userID"`
}

// //skills
type SkillResponseObject struct {
	Status string        `json:"status"`
	Code   int64         `json:"code"`
	Count  int64         `json:"count"`
	Skills []interface{} `json:"skills"`
}

type NewSkillResponseObject struct {
	Status  string `json:"status"`
	Code    int64  `json:"code"`
	SkillID string `json:"skillID"`
}
