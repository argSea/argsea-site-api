package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// blockSetFakeOutAdapter is an in-memory BlockSetRepo for tests.
type blockSetFakeOutAdapter struct {
	sets *map[string]domain.BlockSet
	seq  *int
}

func NewBlockSetFakeOutAdapter() out_port.BlockSetRepo {
	return blockSetFakeOutAdapter{
		sets: &map[string]domain.BlockSet{},
		seq:  new(int),
	}
}

func (b blockSetFakeOutAdapter) List() (domain.BlockSets, error) {
	var out domain.BlockSets

	for _, set := range *b.sets {
		out = append(out, set)
	}

	return out, nil
}

func (b blockSetFakeOutAdapter) Get(id string) domain.BlockSet {
	return (*b.sets)[id]
}

func (b blockSetFakeOutAdapter) Add(set domain.BlockSet) (string, error) {
	*b.seq++
	id := fmt.Sprintf("bset-%d", *b.seq)
	set.Id = id
	(*b.sets)[id] = set

	return id, nil
}

func (b blockSetFakeOutAdapter) Remove(id string) error {
	delete(*b.sets, id)

	return nil
}
