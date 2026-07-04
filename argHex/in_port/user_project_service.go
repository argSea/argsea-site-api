package in_port

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
)

type UserProjectService interface {
	GetProjects(userID string) (domain.Projects, int64)
}
