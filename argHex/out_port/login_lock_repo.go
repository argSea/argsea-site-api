package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// The login strike ledger, split read from write so the login path can check a
// client's standing without holding a mutation handle.

// LoginLockReadRepo reads one client IP's standing at the login.
type LoginLockReadRepo interface {
	GetByIP(ip string) domain.LoginLock
}

// LoginLockWriteRepo lands a client IP's standing: Save upserts its miss count
// and struck flag, ClearByIP wipes the slate when a good hail lands.
type LoginLockWriteRepo interface {
	Save(lock domain.LoginLock) error
	ClearByIP(ip string) error
}
