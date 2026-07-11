package domain

type ActivityLogs []ActivityLog

// ActivityLog is one line in the keeper's log: an append-only record written by
// every content mutation. Timestamp is RFC3339 UTC so it sorts chronologically
// as a plain string. EntityType/EntityId tie the entry back to what changed.
type ActivityLog struct {
	Id         string `json:"id" bson:"_id,omitempty"`
	Timestamp  string `json:"timestamp" bson:"timestamp,omitempty"`
	Message    string `json:"message" bson:"message,omitempty"`
	EntityType string `json:"entityType" bson:"entityType,omitempty"`
	EntityId   string `json:"entityId" bson:"entityId,omitempty"`
}
