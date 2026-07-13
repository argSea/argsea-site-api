package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type CaseLogRepo interface {
	List(publishedOnly bool, limit int64) (domain.CaseLogs, error)
	Get(id string) domain.CaseLog
	Add(log domain.CaseLog) (string, error)
	Set(log domain.CaseLog) error
	Remove(id string) error
}
