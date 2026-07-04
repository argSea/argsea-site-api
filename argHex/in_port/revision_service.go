package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// RevisionService is the generic, content-agnostic history seam. Projects and
// notes both depend on it: they snapshot their full document on every write and
// read back the last few printings for the rollback UI.
type RevisionService interface {
	// Snapshot records snapshot as the new current revision for the entity and
	// clears the current flag on any prior revision. Returns the new id.
	Snapshot(entityType string, entityID string, snapshot string, summary string) (string, error)
	// List returns an entity's revisions newest-first, capped at limit.
	List(entityType string, entityID string, limit int64) (domain.Revisions, error)
	// Get returns a single revision by id (empty if not found).
	Get(id string) domain.Revision
}
