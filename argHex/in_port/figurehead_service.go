package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrDesignPublished is returned by Delete for a design that is currently on
// the bow — the adapter maps it to a 409. Publish another design for the pose
// first; a published cat is superseded, never torn off.
var ErrDesignPublished = errors.New("a published design cannot be deleted — publish another design for its pose first")

// ErrDesignSeeded is returned by Delete and Update for the seeded v1 designs —
// the adapter maps it to a 409. The v1 cats are permanent so "go back to v1"
// is always possible; supersede them by publishing something else.
var ErrDesignSeeded = errors.New("the seeded v1 designs are permanent — supersede them by publishing another design")

// FigureheadService is the Figurehead Shop counter: CRUD over the cat designs
// plus the one-published-per-pose Publish swap. Published is the public read
// the site builds against; Seed plants the shipped v1 cats into an empty
// collection at boot and is a no-op forever after.
type FigureheadService interface {
	Published() (domain.CatDesigns, error)
	List() (domain.CatDesigns, error)
	Create(design domain.CatDesign) (domain.CatDesign, error)
	Update(design domain.CatDesign) (domain.CatDesign, error)
	Delete(id string) error
	Publish(id string) (domain.CatDesign, error)
	Seed() error
}
