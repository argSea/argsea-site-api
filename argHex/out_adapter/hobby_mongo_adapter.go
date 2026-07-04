package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type hobbyMongoAdapter struct {
	store *stores.Mordor
}

func NewHobbyMongoAdapter(store *stores.Mordor) out_port.HobbyRepo {
	return hobbyMongoAdapter{
		store: store,
	}
}

// List returns hobbies in manual sort order, optionally only the active
// ("currently learning") ones.
func (h hobbyMongoAdapter) List(activeOnly bool) (domain.Hobbies, error) {
	var hobbies domain.Hobbies
	sort := bson.D{{Key: "order", Value: 1}}
	var err error

	if activeOnly {
		_, err = h.store.GetMany("active", true, 0, 0, sort, &hobbies)
	} else {
		_, err = h.store.GetAll(0, 0, sort, &hobbies)
	}

	return hobbies, err
}

func (h hobbyMongoAdapter) Get(id string) domain.Hobby {
	var hobby domain.Hobby
	err := h.store.Get("_id", id, &hobby)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Hobby{}
	}

	return hobby
}

func (h hobbyMongoAdapter) Add(hobby domain.Hobby) (string, error) {
	hobby.Id = ""
	return h.store.Write(hobby)
}

func (h hobbyMongoAdapter) Set(hobby domain.Hobby) error {
	key := hobby.Id
	hobby.Id = ""
	return h.store.Update(key, hobby)
}

func (h hobbyMongoAdapter) Remove(id string) error {
	return h.store.Delete(id)
}
