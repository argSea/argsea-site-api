package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type NoteRepo interface {
	List(publishedOnly bool, limit int64) (domain.Notes, error)
	Get(id string) domain.Note
	Add(note domain.Note) (string, error)
	Set(note domain.Note) error
	Remove(id string) error
}
