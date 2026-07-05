package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type mediaMetaMongoAdapter struct {
	store *stores.Mordor
}

func NewMediaMetaMongoAdapter(store *stores.Mordor) out_port.MediaMetaRepo {
	return mediaMetaMongoAdapter{
		store: store,
	}
}

func (m mediaMetaMongoAdapter) List() (domain.MediaList, error) {
	var media domain.MediaList
	_, err := m.store.GetAll(0, 0, nil, &media)

	return media, err
}

func (m mediaMetaMongoAdapter) Get(id string) domain.Media {
	var media domain.Media
	err := m.store.Get("_id", id, &media)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Media{}
	}

	return media
}

func (m mediaMetaMongoAdapter) Add(media domain.Media) (string, error) {
	media.Id = "" // make sure it wasn't set
	return m.store.Write(media)
}

func (m mediaMetaMongoAdapter) Remove(id string) error {
	return m.store.Delete(id)
}
