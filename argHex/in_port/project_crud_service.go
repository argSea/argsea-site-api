package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type ProjectCRUDService interface {
	List(publishedOnly bool, limit int64) (domain.Projects, error)
	Create(project domain.Project) (domain.Project, error)
	Read(id string) domain.Project
	Update(project domain.Project) (domain.Project, error)
	Delete(id string) error
	Publish(id string) (domain.Project, error)
	Unpublish(id string) (domain.Project, error)
	Revisions(id string, limit int64) (domain.Revisions, error)
	Restore(id string, revisionID string) (domain.Project, error)
}
