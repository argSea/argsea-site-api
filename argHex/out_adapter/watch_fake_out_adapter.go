package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// watchFakeOutAdapter is an in-memory WatchRepo for tests. It mirrors the mongo
// adapter's singleton semantics: the caller's Id is ignored, the first Save
// mints one and every later Save replaces the document under it.
type watchFakeOutAdapter struct {
	watch *domain.Watch
	seq   *int
}

func NewWatchFakeOutAdapter() out_port.WatchRepo {
	return watchFakeOutAdapter{
		watch: &domain.Watch{},
		seq:   new(int),
	}
}

func (w watchFakeOutAdapter) Get() domain.Watch {
	return *w.watch
}

func (w watchFakeOutAdapter) Save(watch domain.Watch) (domain.Watch, error) {
	if "" == w.watch.Id {
		*w.seq++
		watch.Id = fmt.Sprintf("watch-%d", *w.seq)
		*w.watch = watch

		return watch, nil
	}

	watch.Id = w.watch.Id
	*w.watch = watch

	return watch, nil
}
