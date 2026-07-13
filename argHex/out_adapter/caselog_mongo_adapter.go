package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type caseLogMongoAdapter struct {
	store *stores.Mordor
}

func NewCaseLogMongoAdapter(store *stores.Mordor) out_port.CaseLogRepo {
	return caseLogMongoAdapter{
		store: store,
	}
}

// List returns logs, optionally narrowed to published ones (what the Astro build
// consumes). A limit of 0 means "no limit". The service re-sorts by createdAt,
// which is a stable no-op over this query order.
func (c caseLogMongoAdapter) List(publishedOnly bool, limit int64) (domain.CaseLogs, error) {
	var logs domain.CaseLogs
	var err error

	sort := bson.D{{Key: "createdAt", Value: 1}}

	if publishedOnly {
		_, err = c.store.GetMany("status", domain.StatusPublished, limit, 0, sort, &logs)
	} else {
		_, err = c.store.GetAll(limit, 0, sort, &logs)
	}

	return logs, err
}

func (c caseLogMongoAdapter) Get(id string) domain.CaseLog {
	var log domain.CaseLog
	err := c.store.Get("_id", id, &log)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.CaseLog{}
	}

	return log
}

func (c caseLogMongoAdapter) Add(log domain.CaseLog) (string, error) {
	log.Id = "" // make sure it wasn't set
	return c.store.Write(log)
}

func (c caseLogMongoAdapter) Set(log domain.CaseLog) error {
	key := log.Id
	log.Id = "" // replacement doc must not carry _id; mongo keeps the existing one
	return c.store.Replace(key, log)
}

func (c caseLogMongoAdapter) Remove(id string) error {
	return c.store.Delete(id)
}
