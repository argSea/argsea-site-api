package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type noteMongoAdapter struct {
	store *stores.Mordor
}

func NewNoteMongoAdapter(store *stores.Mordor) out_port.NoteRepo {
	return noteMongoAdapter{
		store: store,
	}
}

func (n noteMongoAdapter) List(publishedOnly bool, limit int64) (domain.Notes, error) {
	var notes domain.Notes
	var err error

	if publishedOnly {
		_, err = n.store.GetMany("status", domain.StatusPublished, limit, 0, nil, &notes)
	} else {
		_, err = n.store.GetAll(limit, 0, nil, &notes)
	}

	return notes, err
}

func (n noteMongoAdapter) Get(id string) domain.Note {
	var note domain.Note
	err := n.store.Get("_id", id, &note)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Note{}
	}

	return note
}

func (n noteMongoAdapter) Add(note domain.Note) (string, error) {
	note.Id = ""
	return n.store.Write(note)
}

func (n noteMongoAdapter) Set(note domain.Note) error {
	key := note.Id
	note.Id = ""
	return n.store.Update(key, note)
}

func (n noteMongoAdapter) Remove(id string) error {
	return n.store.Delete(id)
}
