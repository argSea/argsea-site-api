package service

import (
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type suggestionService struct {
	repo     out_port.SuggestionRepo
	activity in_port.ActivityService
}

func NewSuggestionService(repo out_port.SuggestionRepo, activity in_port.ActivityService) in_port.SuggestionService {
	return suggestionService{
		repo:     repo,
		activity: activity,
	}
}

func (s suggestionService) List() (domain.Suggestions, error) {
	return s.repo.List()
}

// Add appends a chip to the pool. Order is derived from the current pool size,
// so new chips land at the end in the sequence they were added.
func (s suggestionService) Add(value string) (domain.Suggestion, error) {
	existing, err := s.repo.List()

	if nil != err {
		return domain.Suggestion{}, err
	}

	suggestion := domain.Suggestion{
		Value: value,
		Order: len(existing),
	}

	id, err := s.repo.Add(suggestion)

	if nil != err {
		return domain.Suggestion{}, err
	}

	suggestion.Id = id
	s.record("suggestion \""+value+"\" added", id)

	return suggestion, nil
}

func (s suggestionService) Delete(id string) error {
	if err := s.repo.Remove(id); nil != err {
		return err
	}

	s.record("suggestion removed", id)

	return nil
}

func (s suggestionService) record(message string, id string) {
	if err := s.activity.Record(message, domain.EntitySuggestion, id); nil != err {
		log.Printf("activity record failed for suggestion %v: %v\n", id, err)
	}
}
