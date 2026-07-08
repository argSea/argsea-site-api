package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type catDesignMongoAdapter struct {
	store *stores.Mordor
}

func NewCatDesignMongoAdapter(store *stores.Mordor) out_port.CatDesignRepo {
	return catDesignMongoAdapter{
		store: store,
	}
}

// List returns every design; the wardrobe is a handful of documents, so the
// published filtering stays in the service.
func (c catDesignMongoAdapter) List() (domain.CatDesigns, error) {
	var designs domain.CatDesigns
	_, err := c.store.GetAll(0, 0, nil, &designs)

	return designs, err
}

func (c catDesignMongoAdapter) Get(id string) domain.CatDesign {
	var design domain.CatDesign
	err := c.store.Get("_id", id, &design)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.CatDesign{}
	}

	return design
}

func (c catDesignMongoAdapter) Add(design domain.CatDesign) (string, error) {
	design.Id = ""
	return c.store.Write(design)
}

func (c catDesignMongoAdapter) Set(design domain.CatDesign) error {
	key := design.Id
	design.Id = ""
	return c.store.Replace(key, design)
}

func (c catDesignMongoAdapter) Remove(id string) error {
	return c.store.Delete(id)
}
