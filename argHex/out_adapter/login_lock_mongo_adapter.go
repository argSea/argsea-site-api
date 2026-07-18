package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// loginLockMongoAdapter is the login lockout ledger in mongo: one doc per client
// IP. It satisfies both the read and the write port; main.go hands the same
// value to each seam.
type loginLockMongoAdapter struct {
	store *stores.Mordor
}

func NewLoginLockMongoAdapter(store *stores.Mordor) loginLockMongoAdapter {
	return loginLockMongoAdapter{
		store: store,
	}
}

// GetByIP reads a client's standing, or a zero lock carrying the IP when none is
// on file: an IP that has never missed reads as no misses and unbarred.
func (l loginLockMongoAdapter) GetByIP(ip string) domain.LoginLock {
	var lock domain.LoginLock
	err := l.store.Get("ip", ip, &lock)

	if nil != err {
		return domain.LoginLock{IP: ip}
	}

	return lock
}

// Save upserts a client's standing keyed on its IP, so a first miss creates the
// doc and every later miss updates it. Only the mutable counters travel in the
// $set; the IP rides the filter onto the inserted doc.
func (l loginLockMongoAdapter) Save(lock domain.LoginLock) error {
	filter := bson.M{"ip": lock.IP}
	update := bson.D{{Key: "$set", Value: bson.M{"misses": lock.Misses, "barred": lock.Barred}}}

	return l.store.Upsert(filter, update)
}

// ClearByIP deletes a client's lock doc, wiping its slate after a good hail.
func (l loginLockMongoAdapter) ClearByIP(ip string) error {
	return l.store.DeleteBy("ip", ip)
}

// EnsureIndexes lands a unique index on ip so the ledger holds exactly one doc
// per client and the upsert never races two docs into being. Called once at
// boot; mongo ignores it if it already exists.
func (l loginLockMongoAdapter) EnsureIndexes() error {
	return l.store.CreateIndexes([]mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "ip", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
}
