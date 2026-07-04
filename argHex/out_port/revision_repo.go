package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type RevisionRepo interface {
	Add(revision domain.Revision) (string, error)
	Get(id string) domain.Revision
	List(entityID string, limit int64) (domain.Revisions, error)
	// ClearCurrent unsets the current flag on every revision of the entity.
	ClearCurrent(entityID string) error
}
