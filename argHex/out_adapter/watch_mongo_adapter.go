package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type watchMongoAdapter struct {
	store *stores.Mordor
}

func NewWatchMongoAdapter(store *stores.Mordor) out_port.WatchRepo {
	return watchMongoAdapter{
		store: store,
	}
}

// Get returns the single watch document, or a zero value when none has been
// kept yet (a never-kept watch is just the empty default).
func (w watchMongoAdapter) Get() domain.Watch {
	var watches []domain.Watch
	_, err := w.store.GetAll(1, 0, nil, &watches)

	if nil != err || 0 == len(watches) {
		return domain.Watch{}
	}

	return watches[0]
}

// Save upserts the singleton: it updates the existing document if one exists,
// otherwise it writes the first.
func (w watchMongoAdapter) Save(watch domain.Watch) (domain.Watch, error) {
	existing := w.Get()

	if "" == existing.Id {
		watch.Id = ""
		id, err := w.store.Write(watch)

		if nil != err {
			return domain.Watch{}, err
		}

		watch.Id = id
		return watch, nil
	}

	watch.Id = ""

	if err := w.store.Replace(existing.Id, watch); nil != err {
		return domain.Watch{}, err
	}

	return w.Get(), nil
}
