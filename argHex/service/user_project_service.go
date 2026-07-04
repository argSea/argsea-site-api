package service

import (
	"fmt"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type userProjectService struct {
	repo out_port.ProjectRepo
}

func NewUserProjectService(repo out_port.ProjectRepo) in_port.UserProjectService {
	return userProjectService{
		repo: repo,
	}
}

func (u userProjectService) GetProjects(id string) (domain.Projects, int64) {
	projects, count, err := u.repo.GetProjectsByUserID(id)

	if nil != err {
		fmt.Println(err)
	}

	return projects, count
}
