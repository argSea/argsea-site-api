package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// ActivityService is the keeper's log seam. Every content mutation records an
// entry; the dashboard reads the most recent ones.
type ActivityService interface {
	Record(message string, entityType string, entityID string) error
	Recent(limit int64) (domain.ActivityLogs, error)
}
