package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type WatchRepo interface {
	Get() domain.Watch
	Save(watch domain.Watch) (domain.Watch, error)
}
