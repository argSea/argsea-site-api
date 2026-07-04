package in_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

//User service for CRUD
type UserCRUDService interface {
	Create(user domain.User) (string, error)
	Read(id string) domain.User
	ReadAll(limit int64, offset int64, sort interface{}) []domain.User
	Update(user domain.User) error
	Delete(user domain.User) error
}
