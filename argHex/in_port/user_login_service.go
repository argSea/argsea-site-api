package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// user auth interface
type UserLoginService interface {
	Login(user domain.User) (domain.User, error)
	Signup(user domain.User) (string, error)
}
