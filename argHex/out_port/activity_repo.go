package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type ActivityRepo interface {
	Add(entry domain.ActivityLog) (string, error)
	Recent(limit int64) (domain.ActivityLogs, error)
}
