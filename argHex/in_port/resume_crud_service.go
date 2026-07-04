package in_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

//User service for CRUD
type ResumeCRUDService interface {
	Create(res domain.Resume) (string, error)
	Read(id string) domain.Resume
	Update(res domain.Resume) error
	Delete(res domain.Resume) error
}
