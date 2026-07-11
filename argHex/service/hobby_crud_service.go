package service

import (
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// The marker vocabulary is closed for the same reason as the light's kind:
// it selects a headstone graphic rendered on the public graveyard, so this
// enum gate is the injection boundary for marker data.
var hobbyMarkers = map[string]bool{"stone": true, "sticks": true, "driftwood": true, "cairn": true, "buoy": true, "lamp": true}

// validateMarker checks marker against the closed vocabulary. An empty marker
// is valid; the site falls back to its default headstone.
func validateMarker(marker string) error {
	if "" == marker {
		return nil
	}

	if !hobbyMarkers[marker] {
		return errors.New("marker must be one of stone, sticks, driftwood, cairn, buoy, lamp")
	}

	return nil
}

// validateWear checks wear is a fraction: 0 is a fresh stone, 1 is worn smooth.
func validateWear(wear float64) error {
	if 0 > wear || 1 < wear {
		return errors.New("wear must be between 0 and 1")
	}

	return nil
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
	// an invalid marker or wear never reaches the store; reject before
	// anything is written
	if err := validateMarker(hobby.Marker); nil != err {
		return domain.Hobby{}, err
	}

	if err := validateWear(hobby.Wear); nil != err {
		return domain.Hobby{}, err
	}

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

// nextOrder places a new hobby after everything already on the shelf:
// max(order)+1 across all hobbies, active or resting. A failed list fails the
// create; silently defaulting would collide at the front of the shelf.
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

	if err := validateMarker(hobby.Marker); nil != err {
		return domain.Hobby{}, err
	}

	if err := validateWear(hobby.Wear); nil != err {
		return domain.Hobby{}, err
	}

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
