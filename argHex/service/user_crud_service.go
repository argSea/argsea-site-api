package service

import (
	"log"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type userCRUDService struct {
	repo out_port.UserRepo
}

func NewUserCRUDService(repo out_port.UserRepo) in_port.UserCRUDService {
	return userCRUDService{
		repo: repo,
	}
}

func (u userCRUDService) Create(user domain.User) (string, error) {
	user_id, err := u.repo.Add(user)

	if nil == err {
		log.Printf("User created with ID: %v\n", user_id)
	} else {
		log.Printf("User not created. err: %v", err)
	}

	return user_id, err
}

func (u userCRUDService) Read(id string) domain.User {
	userI := u.repo.Get(id)

	return userI
}

func (u userCRUDService) ReadAll(limit int64, offset int64, sort interface{}) []domain.User {
	users := u.repo.GetAll(limit, offset, sort)

	return users
}

func (u userCRUDService) Update(user domain.User) error {
	err := u.repo.Set(user)

	if nil == err {
		log.Printf("User updated, user: %v\n", user)
	} else {
		log.Printf("User not updated, error: %v\n", err)
	}

	return err
}

func (u userCRUDService) Delete(user domain.User) error {
	err := u.repo.Remove(user)

	if nil == err {
		log.Printf("User with ID %v deleted successfully\n", user.Id)
	} else {
		log.Printf("User with ID %v could not be deleted, possible user doesn't exist? err: %v\n", user.Id, err)
	}

	return err
}
