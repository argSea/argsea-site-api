package in_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

type ProjectCRUDService interface {
	GetByUserID(userID string) ([]domain.Project, int64, error)
	Create(project domain.Project) (string, error)
	Read(id string) domain.Project
	Update(project domain.Project) error
	Delete(project domain.Project) error
}
