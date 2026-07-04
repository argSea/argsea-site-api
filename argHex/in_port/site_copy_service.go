package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type SiteCopyService interface {
	Get() domain.SiteCopy
	Save(copy domain.SiteCopy) (domain.SiteCopy, error)
}
