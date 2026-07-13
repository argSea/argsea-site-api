package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// caseLogFakeOutAdapter is an in-memory CaseLogRepo for tests.
type caseLogFakeOutAdapter struct {
	logs *map[string]domain.CaseLog
	seq  *int
}

func NewCaseLogFakeOutAdapter() out_port.CaseLogRepo {
	return caseLogFakeOutAdapter{
		logs: &map[string]domain.CaseLog{},
		seq:  new(int),
	}
}

func (c caseLogFakeOutAdapter) List(publishedOnly bool, limit int64) (domain.CaseLogs, error) {
	var out domain.CaseLogs

	for _, log := range *c.logs {
		if publishedOnly && domain.StatusPublished != log.Status {
			continue
		}

		out = append(out, log)
	}

	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (c caseLogFakeOutAdapter) Get(id string) domain.CaseLog {
	return (*c.logs)[id]
}

func (c caseLogFakeOutAdapter) Add(log domain.CaseLog) (string, error) {
	*c.seq++
	id := fmt.Sprintf("log-%d", *c.seq)
	log.Id = id
	(*c.logs)[id] = log

	return id, nil
}

func (c caseLogFakeOutAdapter) Set(log domain.CaseLog) error {
	(*c.logs)[log.Id] = log

	return nil
}

func (c caseLogFakeOutAdapter) Remove(id string) error {
	delete(*c.logs, id)

	return nil
}
