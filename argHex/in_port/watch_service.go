package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type WatchService interface {
	Get() domain.Watch
	Save(watch domain.Watch) (domain.Watch, error)
}
