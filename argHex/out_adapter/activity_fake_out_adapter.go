package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// activityFakeOutAdapter is an in-memory ActivityRepo for tests.
type activityFakeOutAdapter struct {
	entries *[]domain.ActivityLog
	seq     *int
}

func NewActivityFakeOutAdapter() out_port.ActivityRepo {
	return activityFakeOutAdapter{
		entries: &[]domain.ActivityLog{},
		seq:     new(int),
	}
}

func (a activityFakeOutAdapter) Add(entry domain.ActivityLog) (string, error) {
	*a.seq++
	entry.Id = fmt.Sprintf("act-%d", *a.seq)
	*a.entries = append(*a.entries, entry)

	return entry.Id, nil
}

func (a activityFakeOutAdapter) Recent(limit int64) (domain.ActivityLogs, error) {
	var out domain.ActivityLogs

	for i := len(*a.entries) - 1; i >= 0; i-- {
		out = append(out, (*a.entries)[i])
	}

	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}

	return out, nil
}
