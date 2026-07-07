package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type figureheadService struct {
	repo     out_port.CatDesignRepo
	activity in_port.ActivityService
}

func NewFigureheadService(repo out_port.CatDesignRepo, activity in_port.ActivityService) in_port.FigureheadService {
	return figureheadService{
		repo:     repo,
		activity: activity,
	}
}

// Published returns the design on the bow for each pose. It tolerates the
// crash window inside Publish (two published for one pose, never zero) by
// picking the most recently updated — the fixed-width stamp makes the string
// compare chronological.
func (f figureheadService) Published() (domain.CatDesigns, error) {
	designs, err := f.repo.List()

	if nil != err {
		return nil, err
	}

	current := map[string]domain.CatDesign{}

	for _, design := range designs {
		if !design.Published {
			continue
		}

		if held, ok := current[design.Pose]; !ok || design.UpdatedAt > held.UpdatedAt {
			current[design.Pose] = design
		}
	}

	var out domain.CatDesigns

	for _, pose := range []string{domain.PosePerched, domain.PoseLying} {
		if design, ok := current[pose]; ok {
			out = append(out, design)
		}
	}

	return out, nil
}

func (f figureheadService) List() (domain.CatDesigns, error) {
	return f.repo.List()
}

// Create stores a new draft. Lifecycle flags never arrive through the door:
// published only moves through Publish, and nobody gets to mint a seed.
func (f figureheadService) Create(design domain.CatDesign) (domain.CatDesign, error) {
	if err := validateDesign(design); nil != err {
		return domain.CatDesign{}, err
	}

	now := nowStamp()

	design.Id = ""
	design.Published = false
	design.Seed = false
	design.CreatedAt = now
	design.UpdatedAt = now

	id, err := f.repo.Add(design)

	if nil != err {
		return domain.CatDesign{}, err
	}

	saved := f.repo.Get(id)

	f.record("figurehead design \""+saved.Label+"\" ("+saved.Pose+") created", saved.Id)

	return saved, nil
}

// Update writes new label/viewBox/shapes but leaves the lifecycle alone —
// published only moves through Publish, and a design never changes stance
// after it is carved. The seeded v1s are immutable outright: an editable seed
// would defeat the always-a-v1-to-go-back-to guarantee.
func (f figureheadService) Update(design domain.CatDesign) (domain.CatDesign, error) {
	existing := f.repo.Get(design.Id)

	if "" == existing.Id {
		return domain.CatDesign{}, errors.New("design not found")
	}

	if existing.Seed {
		return domain.CatDesign{}, in_port.ErrDesignSeeded
	}

	design.Pose = existing.Pose

	if err := validateDesign(design); nil != err {
		return domain.CatDesign{}, err
	}

	design.Published = existing.Published
	design.Seed = existing.Seed
	design.CreatedAt = existing.CreatedAt
	design.UpdatedAt = nowStamp()

	if err := f.repo.Set(design); nil != err {
		return domain.CatDesign{}, err
	}

	saved := f.repo.Get(design.Id)

	f.record("figurehead design \""+saved.Label+"\" ("+saved.Pose+") edited", saved.Id)

	return saved, nil
}

func (f figureheadService) Delete(id string) error {
	existing := f.repo.Get(id)

	if "" == existing.Id {
		return errors.New("design not found")
	}

	if existing.Seed {
		return in_port.ErrDesignSeeded
	}

	if existing.Published {
		return in_port.ErrDesignPublished
	}

	if err := f.repo.Remove(id); nil != err {
		return err
	}

	f.record("figurehead design \""+existing.Label+"\" ("+existing.Pose+") deleted", id)

	return nil
}

// Publish hoists the design as its pose's cat and lowers whatever flew there
// before. Hoist first, lower after: a crash between the writes leaves the pose
// with two published designs (Published picks the newer) rather than none —
// the site must never build without a cat.
func (f figureheadService) Publish(id string) (domain.CatDesign, error) {
	design := f.repo.Get(id)

	if "" == design.Id {
		return domain.CatDesign{}, errors.New("design not found")
	}

	others, err := f.repo.List()

	if nil != err {
		return domain.CatDesign{}, err
	}

	now := nowStamp()
	design.Published = true
	design.UpdatedAt = now

	if err := f.repo.Set(design); nil != err {
		return domain.CatDesign{}, err
	}

	for _, other := range others {
		if other.Id == design.Id || other.Pose != design.Pose || !other.Published {
			continue
		}

		other.Published = false
		other.UpdatedAt = now

		if err := f.repo.Set(other); nil != err {
			return domain.CatDesign{}, err
		}
	}

	f.record("figurehead design \""+design.Label+"\" published as the "+design.Pose+" cat", id)

	return f.repo.Get(id), nil
}

// Seed plants the two shipped v1 cats into an empty collection, published, at
// boot. Anything already in the collection means a keeper has been here — the
// seed never runs twice and never touches existing designs.
func (f figureheadService) Seed() error {
	existing, err := f.repo.List()

	if nil != err {
		return err
	}

	if 0 != len(existing) {
		return nil
	}

	now := nowStamp()

	for _, design := range []domain.CatDesign{seedPerchedV1(), seedLyingV1()} {
		design.Published = true
		design.Seed = true
		design.CreatedAt = now
		design.UpdatedAt = now

		id, err := f.repo.Add(design)

		if nil != err {
			return err
		}

		f.record("figurehead design \""+design.Label+"\" ("+design.Pose+") seeded", id)
	}

	return nil
}

// validateDesign is the vocabulary gate: pose and shape type are closed enums
// so the renderers only ever meet primitives they know. Role is deliberately
// not validated — the contract stores it opaquely.
func validateDesign(design domain.CatDesign) error {
	if domain.PosePerched != design.Pose && domain.PoseLying != design.Pose {
		return errors.New("pose must be perched or lying")
	}

	for _, shape := range design.Shapes {
		switch shape.Type {
		case "path", "ellipse", "rect", "line":
		default:
			return errors.New("shape type must be one of path, ellipse, rect, line")
		}
	}

	return nil
}

func (f figureheadService) record(message string, id string) {
	if err := f.activity.Record(message, domain.EntityFigurehead, id); nil != err {
		log.Printf("activity record failed for figurehead design %v: %v\n", id, err)
	}
}
