package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// sightingTTLSeconds keeps a sighting for roughly four hundred days, long
// enough for a year-over-year read, after which the TTL sweeps it. The ledger
// stays bounded without any manual pruning.
const sightingTTLSeconds int32 = 400 * 24 * 60 * 60

type sightingMongoAdapter struct {
	store *stores.Mordor
}

func NewSightingMongoAdapter(store *stores.Mordor) out_port.SightingRepo {
	return sightingMongoAdapter{
		store: store,
	}
}

func (s sightingMongoAdapter) Add(sighting domain.Sighting) (string, error) {
	sighting.Id = ""
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

// EnsureIndexes lands the TTL on at plus a day+kind index the window read and
// its groupings ride. Called once at boot; mongo ignores any that already
// exist.
func (s sightingMongoAdapter) EnsureIndexes() error {
	return s.store.CreateIndexes([]mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(sightingTTLSeconds),
		},
		{
			Keys: bson.D{{Key: "day", Value: 1}, {Key: "kind", Value: 1}},
		},
	})
}
