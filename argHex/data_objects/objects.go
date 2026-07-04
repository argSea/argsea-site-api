package data_objects

// web
// //general
type ErroredResponseObject struct {
	Status  string      `json:"status"`
	Code    int64       `json:"code"`
	Message interface{} `json:"message"`
}

type ItemLessResponseObject struct {
	Status string `json:"status"`
	Code   int64  `json:"code"`
}

// //user
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
