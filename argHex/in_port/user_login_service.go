package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrLoginBarred refuses a hail from a barred client IP. It carries the
// keeper-voiced line the console shows; it leaks nothing an attacker can act on,
// so showing it is fine. A direct hail never sees it: the adapter sends that
// adrift instead.
var ErrLoginBarred = errors.New("the door is barred for the night. come back with the tide.")

// ErrBadCredentials refuses a hail for a wrong username or password. The message
// stays generic so it never tells the two apart, and a barred refusal stays
// indistinguishable from it under the same 400 the console gets.
var ErrBadCredentials = errors.New("incorrect credentials or user does not exist")

// user auth interface
type UserLoginService interface {
	Login(user domain.User, ip string) (domain.User, error)
	Signup(user domain.User) (string, error)
}
