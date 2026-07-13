package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type BlockSetRepo interface {
	List() (domain.BlockSets, error)
	Get(id string) domain.BlockSet
	Add(set domain.BlockSet) (string, error)
	Remove(id string) error
}
