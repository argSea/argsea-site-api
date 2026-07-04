package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

type NoteCRUDService interface {
	List(publishedOnly bool, limit int64) (domain.Notes, error)
	Create(note domain.Note) (domain.Note, error)
	Read(id string) domain.Note
	Update(note domain.Note) (domain.Note, error)
	Delete(id string) error
	Publish(id string) (domain.Note, error)
	Unpublish(id string) (domain.Note, error)
	Revisions(id string, limit int64) (domain.Revisions, error)
	Restore(id string, revisionID string) (domain.Note, error)
}
