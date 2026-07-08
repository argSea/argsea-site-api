package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// DoodleService is CRUD over the marginalia doodles; no publish/seed/pose
// lifecycle, just structured shapes in and out.
type DoodleService interface {
	List() (domain.Doodles, error)
	Get(id string) domain.Doodle
	Create(doodle domain.Doodle) (domain.Doodle, error)
	Update(doodle domain.Doodle) (domain.Doodle, error)
	Delete(id string) error
}
