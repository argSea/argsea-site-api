package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type HobbyCRUDService interface {
	List(activeOnly bool) (domain.Hobbies, error)
	Create(hobby domain.Hobby) (domain.Hobby, error)
	Read(id string) domain.Hobby
	Update(hobby domain.Hobby) (domain.Hobby, error)
	Delete(id string) error
}
