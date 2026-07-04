package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// revisionFakeOutAdapter is an in-memory RevisionRepo for tests. It keeps the
// append-only log in insertion order and honours the current-flag contract, so
// the revision service's snapshot/restore semantics can be exercised without a
// database.
type revisionFakeOutAdapter struct {
	revisions *[]domain.Revision
	seq       *int
}

func NewRevisionFakeOutAdapter() out_port.RevisionRepo {
	return revisionFakeOutAdapter{
		revisions: &[]domain.Revision{},
		seq:       new(int),
	}
}

func (r revisionFakeOutAdapter) Add(revision domain.Revision) (string, error) {
	*r.seq++
	revision.Id = fmt.Sprintf("rev-%d", *r.seq)
	*r.revisions = append(*r.revisions, revision)

	return revision.Id, nil
}

func (r revisionFakeOutAdapter) Get(id string) domain.Revision {
	for _, rev := range *r.revisions {
		if rev.Id == id {
			return rev
		}
	}

	return domain.Revision{}
}

func (r revisionFakeOutAdapter) List(entityID string, limit int64) (domain.Revisions, error) {
	var out domain.Revisions

	// newest-first: walk the append-only log backwards
	for i := len(*r.revisions) - 1; i >= 0; i-- {
		rev := (*r.revisions)[i]

		if rev.EntityId == entityID {
			out = append(out, rev)
		}
	}

	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (r revisionFakeOutAdapter) ClearCurrentExcept(entityID string, revisionID string) error {
	for i := range *r.revisions {
		if (*r.revisions)[i].EntityId == entityID && (*r.revisions)[i].Id != revisionID {
			(*r.revisions)[i].IsCurrent = false
		}
	}

	return nil
}
