package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type CarvingRepo interface {
	List() (domain.Carvings, error)
	Get(id string) domain.Carving
	Add(carving domain.Carving) (string, error)
	Set(carving domain.Carving) error
	Remove(id string) error
}
