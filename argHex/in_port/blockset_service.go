package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type BlockSetService interface {
	List() (domain.BlockSets, error)
	Create(set domain.BlockSet) (domain.BlockSet, error)
	Delete(id string) error
	Seed() error
}
