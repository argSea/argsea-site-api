package service

import (
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is a valid bcrypt hash at the production cost, compared against when
// the named user does not exist so a wrong username costs the same bcrypt work
// as a wrong password. Without it a missing account returns fast and leaks which
// usernames are real by timing. No practical plaintext hashes to it.
var dummyHash = mustDummyHash()

type userLoginService struct {
	repo      out_port.UserRepo
	lockRead  out_port.LoginLockReadRepo
	lockWrite out_port.LoginLockWriteRepo
}

func NewUserLoginService(repo out_port.UserRepo, lockRead out_port.LoginLockReadRepo, lockWrite out_port.LoginLockWriteRepo) in_port.UserLoginService {
	return userLoginService{
		repo:      repo,
		lockRead:  lockRead,
		lockWrite: lockWrite,
	}
}

// Login authenticates a hail from client ip. A struck IP is refused before any
// user lookup or bcrypt, so a hammerer cannot probe and cannot cost the light
// work. A bad hail records a miss against that IP and strikes it on the sixth; a
// good hail wipes that IP's slate. The error is typed so the adapter can tell a
// struck refusal from bad credentials.
func (u userLoginService) Login(user domain.User, ip string) (domain.User, error) {
	lock := u.lockRead.GetByIP(ip)

	if lock.IsStruck() {
		log.Printf("Login refused: %s is struck", ip)
		return domain.User{}, in_port.ErrLoginStruck
	}

	logged_in_user := u.repo.GetByUserName(user.UserName)

	// compare against the stored hash, or the dummy hash when the user has no
	// usable credential on file, so both failing paths run the same bcrypt work
	// and a missing username never gives itself away by timing
	hash := string(logged_in_user.Password)
	present := "" != hash

	if !present {
		hash = dummyHash
	}

	compare_err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))

	// an absent user never authenticates even on a compare that happens to
	// match the dummy hash: with no credential on file the hail is always refused
	if !present || nil != compare_err {
		log.Printf("User not logged in from %s", ip)
		return domain.User{}, u.strike(ip, lock)
	}

	// a good hail wipes this client's slate; only bother when it had misses
	if 0 < lock.Misses {
		if err := u.lockWrite.ClearByIP(ip); nil != err {
			log.Printf("could not clear login lock for %s: %v", ip, err)
		}
	}

	log.Printf("User logged in with ID: %v\n", logged_in_user.Id)
	return logged_in_user, nil
}

// strike records one more bad hail from ip and returns the refusal to show. The
// sixth miss strikes the light for that IP; a strike write that fails is logged,
// not surfaced, since the refusal still stands.
func (u userLoginService) strike(ip string, lock domain.LoginLock) error {
	lock.IP = ip
	struck := lock.Missed()

	if err := u.lockWrite.Save(struck); nil != err {
		log.Printf("could not record login miss for %s: %v", ip, err)
	}

	if struck.IsStruck() {
		return in_port.ErrLoginStruck
	}

	return in_port.ErrBadCredentials
}

func (u userLoginService) Signup(user domain.User) (string, error) {
	// a signup can never carry a role; admin is granted only by a direct DB
	// update on the user document
	user.Role = ""

	user_id, err := u.repo.Add(user)

	if nil == err {
		log.Printf("User created with ID: %v\n", user_id)
	} else {
		log.Printf("User not created. err: %v", err)
	}

	return user_id, err
}

func mustDummyHash() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("argsea-no-such-keeper"), bcrypt.DefaultCost)

	if nil != err {
		// bcrypt only errors on an out-of-range cost, which DefaultCost is not; a
		// light that cannot build its dummy hash would leak usernames, so fail loud
		panic(err)
	}

	return string(hash)
}
