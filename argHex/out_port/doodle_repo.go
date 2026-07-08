package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type DoodleRepo interface {
	List() (domain.Doodles, error)
	Get(id string) domain.Doodle
	Add(doodle domain.Doodle) (string, error)
	Set(doodle domain.Doodle) error
	Remove(id string) error
}
