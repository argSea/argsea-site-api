package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrCarvingBuiltin guards the seven seeded v1 carvings; the adapter maps it
// to a 409. Name and svg are frozen so every spot always has a v1 to bolt
// back to, and a builtin can never be deleted outright; only the bolt moves.
var ErrCarvingBuiltin = errors.New("the seeded v1 carvings are permanent: name and svg are frozen, and a builtin carving cannot be deleted")

// CarvingService is the carving shop counter: CRUD over the raw-svg carvings
// plus the one-carving-per-spot Bolt swap. List is the public read the site
// builds against; Seed plants the seven shipped v1 carvings into an empty
// collection at boot and is a no-op forever after.
type CarvingService interface {
	List() (domain.Carvings, error)
	Create(carving domain.Carving) (domain.Carving, error)
	Update(carving domain.Carving) (domain.Carving, error)
	Delete(id string) error
	Bolt(id string, spot string) (domain.Carving, error)
	Seed() error
}
