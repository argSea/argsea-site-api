package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// sightingTTLSeconds keeps a sighting for roughly four hundred days, long
// enough for a year-over-year read, after which the TTL sweeps it. The ledger
// stays bounded without any manual pruning.
const sightingTTLSeconds int32 = 400 * 24 * 60 * 60

// sightingMongoAdapter splits its writes across two collections: store keeps
// the TTL and holds everything but flares, drawer holds nothing but flares and
// carries no TTL. Retention is the adapter's call, not the domain's: a flare
// is kept forever by operator ruling, and all it stores is a day-salted
// visitor hash, already unlinkable across days, so keeping it forever never
// stores more than a per-day pseudonym.
type sightingMongoAdapter struct {
	store  *stores.Mordor
	drawer *stores.Mordor
}

func NewSightingMongoAdapter(store *stores.Mordor, drawer *stores.Mordor) out_port.SightingRepo {
	return sightingMongoAdapter{
		store:  store,
		drawer: drawer,
	}
}

// Add routes a flare to the TTL-free drawer and every other kind to sightings,
// unchanged from before the drawer existed.
func (s sightingMongoAdapter) Add(sighting domain.Sighting) (string, error) {
	sighting.Id = ""

	if domain.SightingFlare == sighting.Kind {
		return s.drawer.Write(sighting)
	}

	return s.store.Write(sighting)
}

// Window returns every sighting on or after since (a UTC YYYY-MM-DD string),
// oldest first. Day sorts lexically the same as chronologically, so the range
// match and the sort both ride the day string. The service folds the raw rows
// into the aggregate; keeping the pipeline out of here holds the port boundary.
func (s sightingMongoAdapter) Window(since string) (domain.Sightings, error) {
	var sightings domain.Sightings
	filter := bson.M{"day": bson.M{"$gte": since}}
	sort := bson.D{{Key: "day", Value: 1}}
	_, err := s.store.Find(filter, 0, 0, sort, &sightings)

	return sightings, err
}

// Flares returns every flare sighting ever recorded, oldest first. It reads
// the drawer, which carries no TTL, so the roll call counts forever and never
// rides the window or the sightings TTL.
func (s sightingMongoAdapter) Flares() (domain.Sightings, error) {
	var sightings domain.Sightings
	sort := bson.D{{Key: "day", Value: 1}}
	_, err := s.drawer.GetAll(0, 0, sort, &sightings)

	return sightings, err
}

// EnsureIndexes lands the TTL and the day+kind index on sightings, a plain day
// index (no TTL) on the drawer, and then runs the drawer backfill. Called once
// at boot; mongo ignores any index that already exists.
func (s sightingMongoAdapter) EnsureIndexes() error {
	if err := s.store.CreateIndexes([]mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(sightingTTLSeconds),
		},
		{
			Keys: bson.D{{Key: "day", Value: 1}, {Key: "kind", Value: 1}},
		},
	}); nil != err {
		return err
	}

	if err := s.drawer.CreateIndexes([]mongo.IndexModel{
		{
			Keys: bson.D{{Key: "day", Value: 1}},
		},
	}); nil != err {
		return err
	}

	return s.backfillDrawer()
}

// backfillDrawer is a one-time, idempotent migration riding along in
// EnsureIndexes rather than its own boot step, since it is small enough not to
// earn the hobby migration's separate Migrate pass: it lifts any flare that
// landed in sightings before the drawer existed, upserted by id so a repeat
// boot touches nothing. Those rows stay in sightings too, harmless until the
// TTL sweeps them; the drawer copy is what the roll call reads from here on.
func (s sightingMongoAdapter) backfillDrawer() error {
	var stray domain.Sightings
	filter := bson.M{"kind": domain.SightingFlare}
	sort := bson.D{{Key: "day", Value: 1}}
	_, err := s.store.Find(filter, 0, 0, sort, &stray)

	if nil != err {
		return err
	}

	for _, flare := range stray {
		upsertFilter, update, ok := backfillUpsert(flare)

		if !ok {
			continue
		}

		if err := s.drawer.Upsert(upsertFilter, update); nil != err {
			return err
		}
	}

	return nil
}

// backfillUpsert maps one stray flare row to the filter+update its upsert
// into the drawer rides, keyed on the row's own id so a repeat boot lands the
// same doc rather than a duplicate. ok is false when the row's id does not
// parse as an ObjectID, the guard backfillDrawer skips the row on. Kept as its
// own function since it is the one pure part of the migration, and the one a
// mistake in would fail silently: a wrong filter finds nothing to update,
// upserts a stray doc instead, and either way returns a nil error.
func backfillUpsert(flare domain.Sighting) (bson.M, bson.D, bool) {
	id, idErr := primitive.ObjectIDFromHex(flare.Id)

	if nil != idErr {
		return nil, nil, false
	}

	filter := bson.M{"_id": id}
	update := bson.D{{Key: "$set", Value: bson.M{
		"kind":    flare.Kind,
		"day":     flare.Day,
		"path":    flare.Path,
		"subject": flare.Subject,
		"port":    flare.Port,
		"visitor": flare.Visitor,
		"at":      flare.At,
	}}}

	return filter, update, true
}
