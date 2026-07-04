package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
)

type suggestionMongoAdapter struct {
	store *stores.Mordor
}

func NewSuggestionMongoAdapter(store *stores.Mordor) out_port.SuggestionRepo {
	return suggestionMongoAdapter{
		store: store,
	}
}

func (s suggestionMongoAdapter) List() (domain.Suggestions, error) {
	var suggestions domain.Suggestions
	_, err := s.store.GetAll(0, 0, bson.D{{Key: "order", Value: 1}}, &suggestions)

	return suggestions, err
}

func (s suggestionMongoAdapter) Add(suggestion domain.Suggestion) (string, error) {
	suggestion.Id = ""
	return s.store.Write(suggestion)
}

func (s suggestionMongoAdapter) Remove(id string) error {
	return s.store.Delete(id)
}
