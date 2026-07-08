package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// doodleFakeOutAdapter is an in-memory DoodleRepo for tests.
type doodleFakeOutAdapter struct {
	doodles *map[string]domain.Doodle
	seq     *int
}

func NewDoodleFakeOutAdapter() out_port.DoodleRepo {
	return doodleFakeOutAdapter{
		doodles: &map[string]domain.Doodle{},
		seq:     new(int),
	}
}

func (d doodleFakeOutAdapter) List() (domain.Doodles, error) {
	var out domain.Doodles

	for _, doodle := range *d.doodles {
		out = append(out, doodle)
	}

	return out, nil
}

func (d doodleFakeOutAdapter) Get(id string) domain.Doodle {
	return (*d.doodles)[id]
}

func (d doodleFakeOutAdapter) Add(doodle domain.Doodle) (string, error) {
	*d.seq++
	id := fmt.Sprintf("doodle-%d", *d.seq)
	doodle.Id = id
	(*d.doodles)[id] = doodle

	return id, nil
}

func (d doodleFakeOutAdapter) Set(doodle domain.Doodle) error {
	(*d.doodles)[doodle.Id] = doodle

	return nil
}

func (d doodleFakeOutAdapter) Remove(id string) error {
	delete(*d.doodles, id)

	return nil
}
