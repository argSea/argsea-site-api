package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrCarvingBuiltin guards the seven seeded v1 carvings; the adapter maps it
// to a 409. Name and svg are frozen so every spot always has a v1 to bolt
// back to, and a builtin can never be deleted outright; only the bolt moves.
var ErrCarvingBuiltin = errors.New("the seeded v1 carvings are permanent: name and svg are frozen, and a builtin carving cannot be deleted")

// ErrCarvingBolted guards a carving that currently holds a spot; the adapter
// maps it to a 409. A live spot must never go dark or vanish at the next
// hoist, so a bolted carving cannot have its svg blanked and cannot be
// deleted; unbolt it (bolt another carving onto the spot) first.
var ErrCarvingBolted = errors.New("a bolted carving cannot have its svg blanked or be deleted; unbolt it first")

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
