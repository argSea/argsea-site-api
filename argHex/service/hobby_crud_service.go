package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// validateState gates a hobby's state against the closed vocabulary the same
// way the light's kind is gated: the value drives the ship's glyph on the public
// chart, so this is the injection boundary for state. Empty is not a state; a
// ship always stands somewhere.
func validateState(state string) error {
	if !domain.ValidHobbyState(state) {
		return errors.New("state must be one of moored, port, adrift, marooned, inkspill")
	}

	return nil
}

// clampBearings snaps a hobby's coord and wake origin into the chart window
// before the write reaches the store, so an off-window bearing lands at the
// nearest visible edge rather than off the chart. A null bearing is uncharted
// and rides through untouched. The admin editor mirrors this clamp; here it is
// the data-level truth that protects every client.
func clampBearings(hobby *domain.Hobby) {
	domain.ClampCoord(hobby.Coord)
	domain.ClampCoord(hobby.From)
}

type hobbyCRUDService struct {
	repo     out_port.HobbyRepo
	activity in_port.ActivityService
}

func NewHobbyCRUDService(repo out_port.HobbyRepo, activity in_port.ActivityService) in_port.HobbyCRUDService {
	return hobbyCRUDService{
		repo:     repo,
		activity: activity,
	}
}

func (h hobbyCRUDService) List(activeOnly bool) (domain.Hobbies, error) {
	return h.repo.List(activeOnly)
}

func (h hobbyCRUDService) Read(id string) domain.Hobby {
	return h.repo.Get(id)
}

func (h hobbyCRUDService) Create(hobby domain.Hobby) (domain.Hobby, error) {
	// an invalid state never reaches the store; reject before anything is written
	if err := validateState(hobby.State); nil != err {
		return domain.Hobby{}, err
	}

	// an off-window bearing snaps into the chart window before anything is stored
	clampBearings(&hobby)

	now := nowStamp()

	hobby.Id = ""
	hobby.CreatedAt = now
	hobby.UpdatedAt = now

	order, orderErr := h.nextOrder()

	if nil != orderErr {
		return domain.Hobby{}, orderErr
	}

	hobby.Order = order

	id, err := h.repo.Add(hobby)

	if nil != err {
		return domain.Hobby{}, err
	}

	saved := h.repo.Get(id)
	h.record("hobby \""+saved.Name+"\" picked up", saved.Id)

	return saved, nil
}

// nextOrder places a new hobby last in the log: max(order)+1 across every ship,
// whatever its state. A failed list fails the create; silently defaulting would
// collide at the head of the log.
func (h hobbyCRUDService) nextOrder() (int, error) {
	hobbies, err := h.repo.List(false)

	if nil != err {
		return 0, err
	}

	max := 0

	for _, hobby := range hobbies {
		if hobby.Order > max {
			max = hobby.Order
		}
	}

	return max + 1, nil
}

func (h hobbyCRUDService) Update(hobby domain.Hobby) (domain.Hobby, error) {
	existing := h.repo.Get(hobby.Id)

	if "" == existing.Id {
		return domain.Hobby{}, errors.New("hobby not found")
	}

	if err := validateState(hobby.State); nil != err {
		return domain.Hobby{}, err
	}

	clampBearings(&hobby)

	hobby.CreatedAt = existing.CreatedAt
	hobby.UpdatedAt = nowStamp()

	if err := h.repo.Set(hobby); nil != err {
		return domain.Hobby{}, err
	}

	saved := h.repo.Get(hobby.Id)
	h.record("hobby \""+saved.Name+"\" updated", saved.Id)

	return saved, nil
}

func (h hobbyCRUDService) Delete(id string) error {
	existing := h.repo.Get(id)

	if err := h.repo.Remove(id); nil != err {
		return err
	}

	h.record("hobby \""+existing.Name+"\" removed", id)

	return nil
}

func (h hobbyCRUDService) record(message string, id string) {
	if err := h.activity.Record(message, domain.EntityHobby, id); nil != err {
		log.Printf("activity record failed for hobby %v: %v\n", id, err)
	}
}
