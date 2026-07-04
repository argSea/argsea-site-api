package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type SuggestionService interface {
	List() (domain.Suggestions, error)
	Add(value string) (domain.Suggestion, error)
	Delete(id string) error
}
