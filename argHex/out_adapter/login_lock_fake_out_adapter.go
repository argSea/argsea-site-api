package out_adapter

import "github.com/argSea/argsea-site-api/argHex/domain"

// loginLockFakeOutAdapter is an in-memory login lock ledger for tests. It
// satisfies both the read and the write port, so a test hands the one value to
// each seam and inspects the standing through GetByIP.
type loginLockFakeOutAdapter struct {
	locks map[string]domain.LoginLock
}

func NewLoginLockFakeOutAdapter() *loginLockFakeOutAdapter {
	return &loginLockFakeOutAdapter{
		locks: map[string]domain.LoginLock{},
	}
}

func (l *loginLockFakeOutAdapter) GetByIP(ip string) domain.LoginLock {
	lock, ok := l.locks[ip]

	if !ok {
		return domain.LoginLock{IP: ip}
	}

	return lock
}

func (l *loginLockFakeOutAdapter) Save(lock domain.LoginLock) error {
	l.locks[lock.IP] = lock

	return nil
}

func (l *loginLockFakeOutAdapter) ClearByIP(ip string) error {
	delete(l.locks, ip)

	return nil
}
