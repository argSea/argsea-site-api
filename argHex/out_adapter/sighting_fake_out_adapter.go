package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// sightingFakeOutAdapter is an in-memory SightingRepo for tests.
type sightingFakeOutAdapter struct {
	sightings *[]domain.Sighting
	seq       *int
}

func NewSightingFakeOutAdapter() out_port.SightingRepo {
	return sightingFakeOutAdapter{
		sightings: &[]domain.Sighting{},
		seq:       new(int),
	}
}

func (s sightingFakeOutAdapter) Add(sighting domain.Sighting) (string, error) {
	*s.seq++
	sighting.Id = fmt.Sprintf("sighting-%d", *s.seq)
	*s.sightings = append(*s.sightings, sighting)

	return sighting.Id, nil
}

func (s sightingFakeOutAdapter) Window(since string) (domain.Sightings, error) {
	var out domain.Sightings

	for _, sighting := range *s.sightings {
		if sighting.Day >= since {
			out = append(out, sighting)
		}
	}

	return out, nil
}

func (s sightingFakeOutAdapter) Flares() (domain.Sightings, error) {
	var out domain.Sightings

	for _, sighting := range *s.sightings {
		if domain.SightingFlare == sighting.Kind {
			out = append(out, sighting)
		}
	}

	return out, nil
}

func (s sightingFakeOutAdapter) EnsureIndexes() error {
	return nil
}
