package service

import (
	"log"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type userResumeService struct {
	repo out_port.ResumeRepo
}

func NewUserResumeService(repo out_port.ResumeRepo) in_port.UserResumeService {
	resu := userResumeService{
		repo: repo,
	}

	return resu
}

func (res userResumeService) GetResumes(userID string) (domain.Resumes, int64) {
	resumes, count, err := res.repo.GetByUserID(userID)

	if nil != err {
		log.Println(err)
	}

	return resumes, count

}
