package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type siteCopyMongoAdapter struct {
	store *stores.Mordor
}

func NewSiteCopyMongoAdapter(store *stores.Mordor) out_port.SiteCopyRepo {
	return siteCopyMongoAdapter{
		store: store,
	}
}

// Get returns the single site-copy document, or a zero value when none has been
// saved yet (the new database starts empty).
func (s siteCopyMongoAdapter) Get() domain.SiteCopy {
	var copies []domain.SiteCopy
	_, err := s.store.GetAll(1, 0, nil, &copies)

	if nil != err || 0 == len(copies) {
		return domain.SiteCopy{}
	}

	return copies[0]
}

// Save upserts the singleton: it updates the existing document if one exists,
// otherwise it writes the first.
func (s siteCopyMongoAdapter) Save(copy domain.SiteCopy) (domain.SiteCopy, error) {
	existing := s.Get()

	if "" == existing.Id {
		copy.Id = ""
		id, err := s.store.Write(copy)

		if nil != err {
			return domain.SiteCopy{}, err
		}

		copy.Id = id
		return copy, nil
	}

	copy.Id = ""

	if err := s.store.Update(existing.Id, copy); nil != err {
		return domain.SiteCopy{}, err
	}

	return s.Get(), nil
}
