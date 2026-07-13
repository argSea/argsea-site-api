package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type CaseLogCRUDService interface {
	List(publishedOnly bool, limit int64) (domain.CaseLogs, error)
	Create(log domain.CaseLog) (domain.CaseLog, error)
	Read(id string) domain.CaseLog
	Update(log domain.CaseLog) (domain.CaseLog, error)
	Delete(id string) error
	Publish(id string) (domain.CaseLog, error)
	Unpublish(id string) (domain.CaseLog, error)
	Revisions(id string, limit int64) (domain.Revisions, error)
	Restore(id string, revisionID string) (domain.CaseLog, error)
}
