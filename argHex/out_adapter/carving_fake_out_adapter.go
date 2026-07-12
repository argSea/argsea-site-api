package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// carvingFakeOutAdapter is an in-memory CarvingRepo for tests.
type carvingFakeOutAdapter struct {
	carvings *map[string]domain.Carving
	seq      *int
}

func NewCarvingFakeOutAdapter() out_port.CarvingRepo {
	return carvingFakeOutAdapter{
		carvings: &map[string]domain.Carving{},
		seq:      new(int),
	}
}

func (c carvingFakeOutAdapter) List() (domain.Carvings, error) {
	var out domain.Carvings

	for _, carving := range *c.carvings {
		out = append(out, carving)
	}

	return out, nil
}

func (c carvingFakeOutAdapter) Get(id string) domain.Carving {
	return (*c.carvings)[id]
}

func (c carvingFakeOutAdapter) Add(carving domain.Carving) (string, error) {
	*c.seq++
	id := fmt.Sprintf("carving-%d", *c.seq)
	carving.Id = id
	(*c.carvings)[id] = carving

	return id, nil
}

func (c carvingFakeOutAdapter) Set(carving domain.Carving) error {
	(*c.carvings)[carving.Id] = carving

	return nil
}

func (c carvingFakeOutAdapter) Remove(id string) error {
	delete(*c.carvings, id)

	return nil
}
