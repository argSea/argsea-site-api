package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type HobbyRepo interface {
	List(activeOnly bool) (domain.Hobbies, error)
	Get(id string) domain.Hobby
	Add(hobby domain.Hobby) (string, error)
	Set(hobby domain.Hobby) error
	Remove(id string) error
	Migrate() (int, error)
}
