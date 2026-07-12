package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// carvingSvgMaxBytes caps a raw carving: hand-carved svg, not an asset dump.
const carvingSvgMaxBytes = 100 << 10

type carvingService struct {
	repo     out_port.CarvingRepo
	activity in_port.ActivityService
}

func NewCarvingService(repo out_port.CarvingRepo, activity in_port.ActivityService) in_port.CarvingService {
	return carvingService{
		repo:     repo,
		activity: activity,
	}
}

func (c carvingService) List() (domain.Carvings, error) {
	return c.repo.List()
}

// Create stores a new carving, unbolted. Builtin never arrives through the
// door; only Seed mints one, the same rule figurehead holds for its seeded
// v1 designs.
func (c carvingService) Create(carving domain.Carving) (domain.Carving, error) {
	if err := validateCarving(carving); nil != err {
		return domain.Carving{}, err
	}

	now := nowStamp()

	carving.Id = ""
	carving.Builtin = false
	carving.BoltedTo = nil
	carving.CreatedAt = now
	carving.UpdatedAt = now

	id, err := c.repo.Add(carving)

	if nil != err {
		return domain.Carving{}, err
	}

	saved := c.repo.Get(id)

	c.record("carving \""+saved.Name+"\" created", saved.Id)

	return saved, nil
}

// Update writes a new name/svg but leaves the bolt alone; a spot only moves
// through Bolt. The seeded v1s gate harder: name and svg are frozen outright,
// so a spot can always be re-bolted back to its v1 look. A bolted carving
// gates too: its svg cannot be blanked, builtin or not, or the spot goes dark
// at the next hoist; unbolt it (bolt another carving onto the spot) first.
func (c carvingService) Update(carving domain.Carving) (domain.Carving, error) {
	existing := c.repo.Get(carving.Id)

	if "" == existing.Id {
		return domain.Carving{}, errors.New("carving not found")
	}

	if existing.Builtin && (existing.Name != carving.Name || existing.Svg != carving.Svg) {
		return domain.Carving{}, in_port.ErrCarvingBuiltin
	}

	if 0 != len(existing.BoltedTo) && "" == carving.Svg {
		return domain.Carving{}, in_port.ErrCarvingBolted
	}

	if err := validateCarving(carving); nil != err {
		return domain.Carving{}, err
	}

	carving.Builtin = existing.Builtin
	carving.BoltedTo = existing.BoltedTo
	carving.CreatedAt = existing.CreatedAt
	carving.UpdatedAt = nowStamp()

	if err := c.repo.Set(carving); nil != err {
		return domain.Carving{}, err
	}

	saved := c.repo.Get(carving.Id)

	c.record("carving \""+saved.Name+"\" edited", saved.Id)

	return saved, nil
}

// Delete refuses a builtin outright: the seven v1 carvings are permanent so
// every spot always has a v1 to bolt back to. It also refuses any carving
// still bolted to a spot, the same guard figurehead's Delete holds for a
// published design: unbolt it (bolt another carving onto the spot) first, so
// exactly one carving holds a given spot at a time stays true.
func (c carvingService) Delete(id string) error {
	existing := c.repo.Get(id)

	if "" == existing.Id {
		return errors.New("carving not found")
	}

	if existing.Builtin {
		return in_port.ErrCarvingBuiltin
	}

	if 0 != len(existing.BoltedTo) {
		return in_port.ErrCarvingBolted
	}

	if err := c.repo.Remove(id); nil != err {
		return err
	}

	c.record("carving \""+existing.Name+"\" deleted", id)

	return nil
}

// Bolt hoists the carving onto spot and strips the spot from whoever held it
// before. Hoist first, strip after: a crash between the writes leaves the
// spot bolted twice, never bolted to nothing, the same crash-safety order
// figurehead's Publish uses for a pose.
func (c carvingService) Bolt(id string, spot string) (domain.Carving, error) {
	if !domain.CarvingSpots[spot] {
		return domain.Carving{}, errors.New("spot must be one of the seven carving spots")
	}

	carving := c.repo.Get(id)

	if "" == carving.Id {
		return domain.Carving{}, errors.New("carving not found")
	}

	if "" == carving.Svg {
		return domain.Carving{}, errors.New("a carving with no svg cannot be bolted")
	}

	others, err := c.repo.List()

	if nil != err {
		return domain.Carving{}, err
	}

	now := nowStamp()

	if !hasSpot(carving.BoltedTo, spot) {
		carving.BoltedTo = append(carving.BoltedTo, spot)
	}

	carving.UpdatedAt = now

	if err := c.repo.Set(carving); nil != err {
		return domain.Carving{}, err
	}

	for _, other := range others {
		if other.Id == carving.Id || !hasSpot(other.BoltedTo, spot) {
			continue
		}

		other.BoltedTo = withoutSpot(other.BoltedTo, spot)
		other.UpdatedAt = now

		if err := c.repo.Set(other); nil != err {
			return domain.Carving{}, err
		}
	}

	c.record("carving \""+carving.Name+"\" bolted onto "+spot, id)

	return c.repo.Get(id), nil
}

// Seed plants the seven shipped v1 carvings into an empty collection at
// boot, each pre-bolted to its own spot: the current look on the site IS the
// v1 bolt. Anything already in the collection means a keeper has been here;
// the seed never runs twice and never touches existing carvings.
func (c carvingService) Seed() error {
	existing, err := c.repo.List()

	if nil != err {
		return err
	}

	if 0 != len(existing) {
		return nil
	}

	now := nowStamp()

	for _, carving := range seedCarvings() {
		carving.Builtin = true
		carving.CreatedAt = now
		carving.UpdatedAt = now

		id, err := c.repo.Add(carving)

		if nil != err {
			return err
		}

		c.record("carving \""+carving.Name+"\" seeded", id)
	}

	return nil
}

// validateCarving is the shared door gate for create and update. BoltedTo is
// deliberately not checked here: create/update never carry a client-supplied
// spot list past the preserved existing value, only Bolt moves it.
func validateCarving(carving domain.Carving) error {
	if "" == carving.Name {
		return errors.New("carving name is required")
	}

	if carvingSvgMaxBytes < len(carving.Svg) {
		return errors.New("carving svg must be 100KB or smaller")
	}

	return nil
}

func hasSpot(spots []string, spot string) bool {
	for _, s := range spots {
		if spot == s {
			return true
		}
	}

	return false
}

func withoutSpot(spots []string, spot string) []string {
	var out []string

	for _, s := range spots {
		if spot != s {
			out = append(out, s)
		}
	}

	return out
}

func (c carvingService) record(message string, id string) {
	if err := c.activity.Record(message, domain.EntityCarving, id); nil != err {
		log.Printf("activity record failed for carving %v: %v\n", id, err)
	}
}
