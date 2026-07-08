package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type doodleMongoAdapter struct {
	store *stores.Mordor
}

func NewDoodleMongoAdapter(store *stores.Mordor) out_port.DoodleRepo {
	return doodleMongoAdapter{
		store: store,
	}
}

func (d doodleMongoAdapter) List() (domain.Doodles, error) {
	var doodles domain.Doodles
	_, err := d.store.GetAll(0, 0, nil, &doodles)

	return doodles, err
}

func (d doodleMongoAdapter) Get(id string) domain.Doodle {
	var doodle domain.Doodle
	err := d.store.Get("_id", id, &doodle)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Doodle{}
	}

	return doodle
}

func (d doodleMongoAdapter) Add(doodle domain.Doodle) (string, error) {
	doodle.Id = ""
	return d.store.Write(doodle)
}

func (d doodleMongoAdapter) Set(doodle domain.Doodle) error {
	key := doodle.Id
	doodle.Id = ""
	return d.store.Replace(key, doodle)
}

func (d doodleMongoAdapter) Remove(id string) error {
	return d.store.Delete(id)
}
