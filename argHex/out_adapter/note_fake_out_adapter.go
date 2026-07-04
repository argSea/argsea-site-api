package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// noteFakeOutAdapter is an in-memory NoteRepo for tests.
type noteFakeOutAdapter struct {
	notes *map[string]domain.Note
	seq   *int
}

func NewNoteFakeOutAdapter() out_port.NoteRepo {
	return noteFakeOutAdapter{
		notes: &map[string]domain.Note{},
		seq:   new(int),
	}
}

func (n noteFakeOutAdapter) List(publishedOnly bool, limit int64) (domain.Notes, error) {
	var out domain.Notes

	for _, note := range *n.notes {
		if publishedOnly && domain.StatusPublished != note.Status {
			continue
		}

		out = append(out, note)
	}

	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (n noteFakeOutAdapter) Get(id string) domain.Note {
	return (*n.notes)[id]
}

func (n noteFakeOutAdapter) Add(note domain.Note) (string, error) {
	*n.seq++
	id := fmt.Sprintf("note-%d", *n.seq)
	note.Id = id
	(*n.notes)[id] = note

	return id, nil
}

func (n noteFakeOutAdapter) Set(note domain.Note) error {
	(*n.notes)[note.Id] = note

	return nil
}

func (n noteFakeOutAdapter) Remove(id string) error {
	delete(*n.notes, id)

	return nil
}
