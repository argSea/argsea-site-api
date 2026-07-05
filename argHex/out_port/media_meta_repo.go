package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// MediaMetaRepo is the metadata half of the darkroom: the mongo documents that
// describe what lives on disk.
type MediaMetaRepo interface {
	List() (domain.MediaList, error)
	Get(id string) domain.Media
	Add(media domain.Media) (string, error)
	Remove(id string) error
}
