package out_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

//User repo to connect to a store
type UserRepo interface {
	GetAll(limit int64, offset int64, sort interface{}) domain.Users
	Get(id string) domain.User
	GetByUserName(username string) domain.User
	Set(user domain.User) error
	Add(user domain.User) (string, error)
	Remove(user domain.User) error
}
