package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// catDesignFakeOutAdapter is an in-memory CatDesignRepo for tests.
type catDesignFakeOutAdapter struct {
	designs *map[string]domain.CatDesign
	seq     *int
}

func NewCatDesignFakeOutAdapter() out_port.CatDesignRepo {
	return catDesignFakeOutAdapter{
		designs: &map[string]domain.CatDesign{},
		seq:     new(int),
	}
}

func (c catDesignFakeOutAdapter) List() (domain.CatDesigns, error) {
	var out domain.CatDesigns

	for _, design := range *c.designs {
		out = append(out, design)
	}

	return out, nil
}

func (c catDesignFakeOutAdapter) Get(id string) domain.CatDesign {
	return (*c.designs)[id]
}

func (c catDesignFakeOutAdapter) Add(design domain.CatDesign) (string, error) {
	*c.seq++
	id := fmt.Sprintf("design-%d", *c.seq)
	design.Id = id
	(*c.designs)[id] = design

	return id, nil
}

func (c catDesignFakeOutAdapter) Set(design domain.CatDesign) error {
	(*c.designs)[design.Id] = design

	return nil
}

func (c catDesignFakeOutAdapter) Remove(id string) error {
	delete(*c.designs, id)

	return nil
}
