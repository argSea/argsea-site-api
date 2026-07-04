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

type AuthValidationResponseObject struct {
	Valid  bool   `json:"valid"`
	Role   string `json:"roles"`
	UserID string `json:"userID"`
}
