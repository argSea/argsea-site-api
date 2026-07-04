package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type revisionMongoAdapter struct {
	store *stores.Mordor
}

func NewRevisionMongoAdapter(store *stores.Mordor) out_port.RevisionRepo {
	return revisionMongoAdapter{
		store: store,
	}
}

func (r revisionMongoAdapter) Add(revision domain.Revision) (string, error) {
	revision.Id = ""
	return r.store.Write(revision)
}

func (r revisionMongoAdapter) Get(id string) domain.Revision {
	var revision domain.Revision
	err := r.store.Get("_id", id, &revision)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Revision{}
	}

	return revision
}

// List returns an entity's revisions newest-first. Storage is unbounded; the
// limit (last 5 in the admin UI) is applied at read time.
func (r revisionMongoAdapter) List(entityID string, limit int64) (domain.Revisions, error) {
	var revisions domain.Revisions
	sort := bson.D{{Key: "createdAt", Value: -1}}
	_, err := r.store.GetMany("entityId", entityID, limit, 0, sort, &revisions)

	return revisions, err
}

func (r revisionMongoAdapter) ClearCurrent(entityID string) error {
	return r.store.UpdateManyByField("entityId", entityID, bson.M{"isCurrent": false})
}
