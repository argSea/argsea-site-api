package out_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

type ProjectRepo interface {
	GetProjectsByUserID(id string) (domain.Projects, int64, error)
	Get(id string) domain.Project
	Set(project domain.Project) error
	Add(project domain.Project) (string, error)
	Remove(project domain.Project) error
}
