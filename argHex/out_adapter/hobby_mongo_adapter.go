package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

// hobbyRenameMap moves the four prose fields from their postcard-era names to
// their ship's-log names. $rename preserves each value and leaves a doc missing
// a source field untouched, so it is safe over every old-shape doc.
var hobbyRenameMap = bson.M{
	"log":    "bearing",
	"cause":  "offCourse",
	"found":  "floats",
	"return": "odds",
}

type hobbyMongoAdapter struct {
	store *stores.Mordor
}

func NewHobbyMongoAdapter(store *stores.Mordor) out_port.HobbyRepo {
	return hobbyMongoAdapter{
		store: store,
	}
}

// List returns hobbies in manual sort order, optionally only the moored ones:
// the ships at their berth, the moved-over reading of the old active flag.
func (h hobbyMongoAdapter) List(activeOnly bool) (domain.Hobbies, error) {
	var hobbies domain.Hobbies
	sort := bson.D{{Key: "order", Value: 1}}
	var err error

	if activeOnly {
		_, err = h.store.GetMany("state", domain.StateMoored, 0, 0, sort, &hobbies)
	} else {
		_, err = h.store.GetAll(0, 0, sort, &hobbies)
	}

	return hobbies, err
}

func (h hobbyMongoAdapter) Get(id string) domain.Hobby {
	var hobby domain.Hobby
	err := h.store.Get("_id", id, &hobby)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Hobby{}
	}

	return hobby
}

func (h hobbyMongoAdapter) Add(hobby domain.Hobby) (string, error) {
	hobby.Id = ""
	return h.store.Write(hobby)
}

func (h hobbyMongoAdapter) Set(hobby domain.Hobby) error {
	key := hobby.Id
	hobby.Id = ""
	return h.store.Replace(key, hobby)
}

func (h hobbyMongoAdapter) Remove(id string) error {
	return h.store.Delete(id)
}

// Migrate lifts every old-shape hobby doc to the ship's-log shape once, at boot,
// before the routes mount. It is idempotent: only docs still missing state are
// touched, so a reboot moves nothing. Returns the count migrated for the log.
func (h hobbyMongoAdapter) Migrate() (int, error) {
	var migrated int64

	for _, pass := range hobbyMigrationPasses() {
		n, err := h.store.UpdateManyRaw(pass.filter, pass.update)

		if nil != err {
			return int(migrated), err
		}

		migrated += n
	}

	return int(migrated), nil
}

// hobbyMigrationPass is one idempotent update over the old-shape docs: it matches
// only docs still missing state and derives their state from the legacy active
// flag.
type hobbyMigrationPass struct {
	filter bson.M
	update bson.M
}

// hobbyMigrationPasses builds the boot migration to the ship's-log shape. Docs
// missing state get their four prose fields renamed and state/coord/from/seasons
// set, with state derived from active (true is moored, else adrift). The legacy
// fields (active, disposition, marker, char, wear) are left as dead data; nothing
// $unsets them.
func hobbyMigrationPasses() []hobbyMigrationPass {
	return []hobbyMigrationPass{
		{
			filter: bson.M{"state": bson.M{"$exists": false}, "active": true},
			update: hobbyMigrationUpdate(domain.StateMoored),
		},
		{
			filter: bson.M{"state": bson.M{"$exists": false}, "active": bson.M{"$ne": true}},
			update: hobbyMigrationUpdate(domain.StateAdrift),
		},
	}
}

func hobbyMigrationUpdate(state string) bson.M {
	return bson.M{
		"$rename": hobbyRenameMap,
		"$set": bson.M{
			"state":   state,
			"coord":   nil,
			"from":    nil,
			"seasons": "",
		},
	}
}
