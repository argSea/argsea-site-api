package in_port

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
)

//User service for CRUD
type UserResumeService interface {
	GetResumes(userID string) (domain.Resumes, int64)
}
