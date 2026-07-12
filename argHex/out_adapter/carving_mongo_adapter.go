package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type carvingMongoAdapter struct {
	store *stores.Mordor
}

func NewCarvingMongoAdapter(store *stores.Mordor) out_port.CarvingRepo {
	return carvingMongoAdapter{
		store: store,
	}
}

// List returns every carving; the bench holds a handful of documents, so
// spot resolution stays in the service.
func (c carvingMongoAdapter) List() (domain.Carvings, error) {
	var carvings domain.Carvings
	_, err := c.store.GetAll(0, 0, nil, &carvings)

	return carvings, err
}

func (c carvingMongoAdapter) Get(id string) domain.Carving {
	var carving domain.Carving
	err := c.store.Get("_id", id, &carving)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Carving{}
	}

	return carving
}

func (c carvingMongoAdapter) Add(carving domain.Carving) (string, error) {
	carving.Id = ""
	return c.store.Write(carving)
}

func (c carvingMongoAdapter) Set(carving domain.Carving) error {
	key := carving.Id
	carving.Id = ""
	return c.store.Replace(key, carving)
}

func (c carvingMongoAdapter) Remove(id string) error {
	return c.store.Delete(id)
}
