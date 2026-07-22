package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type SightingRepo interface {
	Add(sighting domain.Sighting) (string, error)
	Window(since string) (domain.Sightings, error)
	Flares() (domain.Sightings, error)
	EnsureIndexes() error
}
