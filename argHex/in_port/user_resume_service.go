package in_port

import (
	"github.com/argSea/portfolio_blog_api/argHex/domain"
)

//User service for CRUD
type UserResumeService interface {
	GetResumes(userID string) (domain.Resumes, int64)
}
