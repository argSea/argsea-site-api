package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type RevisionRepo interface {
	Add(revision domain.Revision) (string, error)
	Get(id string) domain.Revision
	List(entityID string, limit int64) (domain.Revisions, error)
	// ClearCurrentExcept unsets the current flag on every revision of the
	// entity other than revisionID. Running it after the new current revision
	// is inserted means a partial failure leaves two currents (self-healing on
	// the next snapshot) instead of zero.
	ClearCurrentExcept(entityID string, revisionID string) error
}
