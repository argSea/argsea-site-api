package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type SuggestionRepo interface {
	List() (domain.Suggestions, error)
	Add(suggestion domain.Suggestion) (string, error)
	Remove(id string) error
}
