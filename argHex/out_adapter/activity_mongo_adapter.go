package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type activityMongoAdapter struct {
	store *stores.Mordor
}

func NewActivityMongoAdapter(store *stores.Mordor) out_port.ActivityRepo {
	return activityMongoAdapter{
		store: store,
	}
}

func (a activityMongoAdapter) Add(entry domain.ActivityLog) (string, error) {
	entry.Id = ""
	return a.store.Write(entry)
}

// Recent returns the newest entries first. Timestamps are RFC3339 UTC strings,
// so a descending string sort is a descending chronological sort.
func (a activityMongoAdapter) Recent(limit int64) (domain.ActivityLogs, error) {
	var entries domain.ActivityLogs
	sort := bson.D{{Key: "timestamp", Value: -1}}
	_, err := a.store.GetAll(limit, 0, sort, &entries)

	return entries, err
}
