package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type lanternMongoAdapter struct {
	store *stores.Mordor
}

// NewLanternMongoAdapter persists the lantern's singleton state document, the
// same upsert-the-first-doc pattern the site copy uses.
func NewLanternMongoAdapter(store *stores.Mordor) out_port.LanternStateRepo {
	return lanternMongoAdapter{
		store: store,
	}
}

// LastHoistedAt returns the persisted stamp, or empty when no hoist has ever
// succeeded (the collection starts empty).
func (l lanternMongoAdapter) LastHoistedAt() (string, error) {
	var states []domain.LanternState
	_, err := l.store.GetAll(1, 0, nil, &states)

	if nil != err || 0 == len(states) {
		return "", err
	}

	return states[0].LastHoistedAt, nil
}

// SaveLastHoistedAt upserts the singleton: update the existing document if one
// exists, otherwise write the first.
func (l lanternMongoAdapter) SaveLastHoistedAt(stamp string) error {
	var states []domain.LanternState
	_, err := l.store.GetAll(1, 0, nil, &states)

	if nil != err {
		return err
	}

	if 0 == len(states) {
		_, writeErr := l.store.Write(domain.LanternState{LastHoistedAt: stamp})

		return writeErr
	}

	return l.store.Update(states[0].Id, domain.LanternState{LastHoistedAt: stamp})
}
