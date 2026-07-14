package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrLoginStruck refuses a hail from a struck client IP. It carries the
// keeper-voiced line the console shows; it leaks nothing an attacker can act on,
// so showing it is fine. A direct hail never sees it: the adapter sends that
// adrift instead.
var ErrLoginStruck = errors.New("the light will not answer. it has been struck for the night.")

// ErrBadCredentials refuses a hail for a wrong username or password. The message
// stays generic so it never tells the two apart, and a struck refusal stays
// indistinguishable from it under the same 400 the console gets.
var ErrBadCredentials = errors.New("incorrect credentials or user does not exist")

// user auth interface
type UserLoginService interface {
	Login(user domain.User, ip string) (domain.User, error)
	Signup(user domain.User) (string, error)
}
