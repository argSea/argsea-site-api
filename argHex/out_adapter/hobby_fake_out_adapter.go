package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// hobbyFakeOutAdapter is an in-memory HobbyRepo for tests.
type hobbyFakeOutAdapter struct {
	hobbies *map[string]domain.Hobby
	seq     *int
}

func NewHobbyFakeOutAdapter() out_port.HobbyRepo {
	return hobbyFakeOutAdapter{
		hobbies: &map[string]domain.Hobby{},
		seq:     new(int),
	}
}

func (h hobbyFakeOutAdapter) List(activeOnly bool) (domain.Hobbies, error) {
	var out domain.Hobbies

	for _, hobby := range *h.hobbies {
		if activeOnly && domain.StateMoored != hobby.State {
			continue
		}

		out = append(out, hobby)
	}

	return out, nil
}

func (h hobbyFakeOutAdapter) Get(id string) domain.Hobby {
	return (*h.hobbies)[id]
}

func (h hobbyFakeOutAdapter) Add(hobby domain.Hobby) (string, error) {
	*h.seq++
	id := fmt.Sprintf("hobby-%d", *h.seq)
	hobby.Id = id
	(*h.hobbies)[id] = hobby

	return id, nil
}

func (h hobbyFakeOutAdapter) Set(hobby domain.Hobby) error {
	(*h.hobbies)[hobby.Id] = hobby

	return nil
}

func (h hobbyFakeOutAdapter) Remove(id string) error {
	delete(*h.hobbies, id)

	return nil
}

// Migrate is a no-op for the in-memory fake: it never holds old-shape docs. The
// boot migration is a mongo $rename/$set concern, exercised over raw docs in the
// adapter's own test.
func (h hobbyFakeOutAdapter) Migrate() (int, error) {
	return 0, nil
}
