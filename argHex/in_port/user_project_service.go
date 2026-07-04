package in_port

import (
	"github.com/argSea/portfolio_blog_api/argHex/domain"
)

type UserProjectService interface {
	GetProjects(userID string) (domain.Projects, int64)
}
