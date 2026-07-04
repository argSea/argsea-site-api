package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type ProjectRepo interface {
	List(publishedOnly bool, limit int64) (domain.Projects, error)
	Get(id string) domain.Project
	Add(project domain.Project) (string, error)
	Set(project domain.Project) error
	Remove(id string) error
}
