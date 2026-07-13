package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type blockSetMongoAdapter struct {
	store *stores.Mordor
}

func NewBlockSetMongoAdapter(store *stores.Mordor) out_port.BlockSetRepo {
	return blockSetMongoAdapter{
		store: store,
	}
}

func (b blockSetMongoAdapter) List() (domain.BlockSets, error) {
	var sets domain.BlockSets
	sort := bson.D{{Key: "name", Value: 1}}
	_, err := b.store.GetAll(0, 0, sort, &sets)

	return sets, err
}

func (b blockSetMongoAdapter) Get(id string) domain.BlockSet {
	var set domain.BlockSet
	err := b.store.Get("_id", id, &set)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.BlockSet{}
	}

	return set
}

func (b blockSetMongoAdapter) Add(set domain.BlockSet) (string, error) {
	set.Id = ""
	return b.store.Write(set)
}

func (b blockSetMongoAdapter) Remove(id string) error {
	return b.store.Delete(id)
}
