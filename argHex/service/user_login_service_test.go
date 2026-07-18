package service_test

import (
	"errors"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
	"golang.org/x/crypto/bcrypt"
)

const keeperPassword = "passphrase"

// spyUserRepo is a single-user UserRepo that answers only for the stored
// username and counts its lookups, so a test can prove a barred hail never
// reaches the user store (and so never runs bcrypt).
type spyUserRepo struct {
	stored  domain.User
	lookups int
}

func (s *spyUserRepo) GetAll(limit int64, offset int64, sort interface{}) domain.Users {
	return domain.Users{s.stored}
}

func (s *spyUserRepo) Get(id string) domain.User {
	return s.stored
}

func (s *spyUserRepo) GetByUserName(username string) domain.User {
	s.lookups++

	if username == s.stored.UserName {
		return s.stored
	}

	return domain.User{}
}

func (s *spyUserRepo) Set(user domain.User) error {
	return nil
}

func (s *spyUserRepo) Add(user domain.User) (string, error) {
	return "new-user", nil
}

func (s *spyUserRepo) Remove(user domain.User) error {
	return nil
}

// newLoginService wires the real login service over a spy user repo holding one
// keeper account and an in-memory lock ledger, returning a read handle on the
// ledger so a test can inspect a client's standing.
func newLoginService(t *testing.T) (in_port.UserLoginService, *spyUserRepo, out_port.LoginLockReadRepo) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(keeperPassword), bcrypt.MinCost)

	if nil != err {
		t.Fatalf("could not hash password: %v", err)
	}

	repo := &spyUserRepo{
		stored: domain.User{
			Id:       "keeper",
			UserName: "meo",
			Password: domain.Password(hash),
		},
	}

	locks := out_adapter.NewLoginLockFakeOutAdapter()

	return service.NewUserLoginService(repo, locks, locks), repo, locks
}

// hail posts one login attempt from ip with the given password.
func hail(svc in_port.UserLoginService, ip string, password string) (domain.User, error) {
	return svc.Login(domain.User{UserName: "meo", Password: domain.Password(password)}, ip)
}

func TestBarAfterSixMisses(t *testing.T) {
	svc, _, locks := newLoginService(t)
	ip := "203.0.113.10"

	for miss := 1; miss <= 5; miss++ {
		if _, err := hail(svc, ip, "wrong"); !errors.Is(err, in_port.ErrBadCredentials) {
			t.Fatalf("miss %d should read as bad credentials, not a bar, got %v", miss, err)
		}

		if locks.GetByIP(ip).IsBarred() {
			t.Fatalf("the door must not bar before the sixth miss, barred at %d", miss)
		}
	}

	if _, err := hail(svc, ip, "wrong"); !errors.Is(err, in_port.ErrLoginBarred) {
		t.Fatalf("the sixth miss must bar the door, got %v", err)
	}

	if !locks.GetByIP(ip).IsBarred() {
		t.Fatalf("the IP must be barred after its sixth miss")
	}
}

func TestBarredShortCircuitsBeforeBcrypt(t *testing.T) {
	svc, repo, _ := newLoginService(t)
	ip := "203.0.113.10"

	for miss := 1; miss <= 6; miss++ {
		hail(svc, ip, "wrong")
	}

	lookupsBeforeBarredHail := repo.lookups

	// the correct password must still be refused while the IP is barred
	if _, err := hail(svc, ip, keeperPassword); !errors.Is(err, in_port.ErrLoginBarred) {
		t.Fatalf("a barred IP must refuse even the correct password, got %v", err)
	}

	if repo.lookups != lookupsBeforeBarredHail {
		t.Fatalf("a barred hail must short-circuit before the user lookup and bcrypt, but the store was read")
	}
}

func TestSecondIPUnaffectedByAnotherBar(t *testing.T) {
	svc, _, _ := newLoginService(t)
	barredIP := "203.0.113.10"
	freshIP := "203.0.113.20"

	for miss := 1; miss <= 6; miss++ {
		hail(svc, barredIP, "wrong")
	}

	if _, err := hail(svc, barredIP, keeperPassword); !errors.Is(err, in_port.ErrLoginBarred) {
		t.Fatalf("the barred IP should stay barred, got %v", err)
	}

	// the whole point of per-IP: one client's misses never lock another
	user, err := hail(svc, freshIP, keeperPassword)

	if nil != err {
		t.Fatalf("a second IP must be unaffected by another IP's bar, got %v", err)
	}

	if "keeper" != user.Id {
		t.Fatalf("the second IP should log in as the keeper, got %q", user.Id)
	}
}

func TestGoodHailClearsTheCounter(t *testing.T) {
	svc, _, locks := newLoginService(t)
	ip := "203.0.113.10"

	for miss := 1; miss <= 3; miss++ {
		hail(svc, ip, "wrong")
	}

	if 3 != locks.GetByIP(ip).Misses {
		t.Fatalf("expected three misses recorded, got %d", locks.GetByIP(ip).Misses)
	}

	if _, err := hail(svc, ip, keeperPassword); nil != err {
		t.Fatalf("a good hail below the bar must log in, got %v", err)
	}

	if 0 != locks.GetByIP(ip).Misses {
		t.Fatalf("a good hail must wipe the client's slate, got %d misses", locks.GetByIP(ip).Misses)
	}
}

func TestAbsentUserReadsAsBadCredentials(t *testing.T) {
	svc, _, _ := newLoginService(t)

	// a username with no document must fail the same generic way a wrong
	// password does, never a distinct fast path that leaks which names are real
	if _, err := svc.Login(domain.User{UserName: "nobody", Password: domain.Password("whatever")}, "203.0.113.30"); !errors.Is(err, in_port.ErrBadCredentials) {
		t.Fatalf("an absent user must read as bad credentials, got %v", err)
	}
}

func TestAbsentUserNeverAuthenticatesOnDummyMatch(t *testing.T) {
	svc, _, _ := newLoginService(t)

	// even a password that hashes to the dummy comparison must never log in an
	// absent user: with no credential on file the hail is always refused
	if _, err := svc.Login(domain.User{UserName: "nobody", Password: domain.Password("argsea-no-such-keeper")}, "203.0.113.30"); nil == err {
		t.Fatalf("an absent user must never authenticate, even on a dummy-hash match")
	}
}
