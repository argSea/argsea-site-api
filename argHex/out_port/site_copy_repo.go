package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type SiteCopyRepo interface {
	Get() domain.SiteCopy
	Save(copy domain.SiteCopy) (domain.SiteCopy, error)
}
