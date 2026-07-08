package domain

type Revisions []Revision

// Revision is one printing in a document's append-only history. Snapshot holds
// the full document as a JSON string at the moment it was recorded. Exactly one
// revision per entity carries IsCurrent; that flag is the "current pointer".
// A rollback copies an older snapshot forward as a new current revision, so the
// pointer move is itself an auditable entry in the log.
type Revision struct {
	Id         string `json:"id" bson:"_id,omitempty"`
	EntityType string `json:"entityType" bson:"entityType,omitempty"`
	EntityId   string `json:"entityId" bson:"entityId,omitempty"`
	Snapshot   string `json:"snapshot" bson:"snapshot,omitempty"` // full document as JSON
	Summary    string `json:"summary" bson:"summary,omitempty"`
	IsCurrent  bool   `json:"isCurrent" bson:"isCurrent"`
	CreatedAt  string `json:"createdAt" bson:"createdAt,omitempty"` // RFC3339 UTC
}
