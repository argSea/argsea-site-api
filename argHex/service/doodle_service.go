package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type doodleService struct {
	repo     out_port.DoodleRepo
	activity in_port.ActivityService
}

func NewDoodleService(repo out_port.DoodleRepo, activity in_port.ActivityService) in_port.DoodleService {
	return doodleService{
		repo:     repo,
		activity: activity,
	}
}

func (d doodleService) List() (domain.Doodles, error) {
	return d.repo.List()
}

func (d doodleService) Get(id string) domain.Doodle {
	return d.repo.Get(id)
}

func (d doodleService) Create(doodle domain.Doodle) (domain.Doodle, error) {
	if err := validateDoodle(doodle); nil != err {
		return domain.Doodle{}, err
	}

	now := nowStamp()

	doodle.Id = ""
	doodle.CreatedAt = now
	doodle.UpdatedAt = now

	id, err := d.repo.Add(doodle)

	if nil != err {
		return domain.Doodle{}, err
	}

	saved := d.repo.Get(id)

	d.record("doodle \""+saved.Name+"\" created", saved.Id)

	return saved, nil
}

// Update is a full replace preserving createdAt; the same shape as the
// figurehead edit path, minus the lifecycle fields doodles don't have.
func (d doodleService) Update(doodle domain.Doodle) (domain.Doodle, error) {
	existing := d.repo.Get(doodle.Id)

	if "" == existing.Id {
		return domain.Doodle{}, errors.New("doodle not found")
	}

	if err := validateDoodle(doodle); nil != err {
		return domain.Doodle{}, err
	}

	doodle.CreatedAt = existing.CreatedAt
	doodle.UpdatedAt = nowStamp()

	if err := d.repo.Set(doodle); nil != err {
		return domain.Doodle{}, err
	}

	saved := d.repo.Get(doodle.Id)

	d.record("doodle \""+saved.Name+"\" edited", saved.Id)

	return saved, nil
}

func (d doodleService) Delete(id string) error {
	existing := d.repo.Get(id)

	if "" == existing.Id {
		return errors.New("doodle not found")
	}

	if err := d.repo.Remove(id); nil != err {
		return err
	}

	d.record("doodle \""+existing.Name+"\" deleted", id)

	return nil
}

// validateDoodle is the vocabulary gate: shape type is a closed enum so the
// renderers only ever meet primitives they know.
func validateDoodle(doodle domain.Doodle) error {
	for _, shape := range doodle.Shapes {
		switch shape.Type {
		case "path", "ellipse", "rect", "line":
		default:
			return errors.New("shape type must be one of path, ellipse, rect, line")
		}
	}

	return nil
}

func (d doodleService) record(message string, id string) {
	if err := d.activity.Record(message, domain.EntityDoodle, id); nil != err {
		log.Printf("activity record failed for doodle %v: %v\n", id, err)
	}
}
