package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// mediaMetaFakeOutAdapter is an in-memory MediaMetaRepo for tests.
type mediaMetaFakeOutAdapter struct {
	media *map[string]domain.Media
	seq   *int
}

func NewMediaMetaFakeOutAdapter() out_port.MediaMetaRepo {
	return mediaMetaFakeOutAdapter{
		media: &map[string]domain.Media{},
		seq:   new(int),
	}
}

func (m mediaMetaFakeOutAdapter) List() (domain.MediaList, error) {
	var out domain.MediaList

	for _, item := range *m.media {
		out = append(out, item)
	}

	return out, nil
}

func (m mediaMetaFakeOutAdapter) Get(id string) domain.Media {
	return (*m.media)[id]
}

func (m mediaMetaFakeOutAdapter) Add(media domain.Media) (string, error) {
	*m.seq++
	id := fmt.Sprintf("media-%d", *m.seq)
	media.Id = id
	(*m.media)[id] = media

	return id, nil
}

func (m mediaMetaFakeOutAdapter) Remove(id string) error {
	delete(*m.media, id)

	return nil
}
