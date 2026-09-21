package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrResumePublished is returned by Delete for the cut that is currently on the
// hoist; the adapter maps it to a 409. Publish another cut first. The refusal
// is what keeps a published record from outliving its own PDF: a delete clears
// the file before the record, so anything that got past this could leave the
// shelf claiming a live resume whose file is already gone.
var ErrResumePublished = errors.New("a published resume cannot be deleted; publish another cut first")

// ResumeValidationError marks a rejection of the request itself: a content type
// the shelf does not take, a missing title, an unknown id, as opposed to an
// infrastructure failure. The adapter maps it to a 400 and everything else to a
// 500 carrying the real cause, so a chmod on the media directory reads back as
// a chmod rather than as a generic failure.
type ResumeValidationError struct {
	Message string
}

func (e ResumeValidationError) Error() string {
	return e.Message
}

// ResumeService is the resume shelf: CRUD over the stored cuts plus the one
// transition they have. Exactly one cut is published at a time, so Publish
// clears whichever held it, and Published is both what the hoist asks before it
// builds anything and what the public read hands out.
type ResumeService interface {
	List() (domain.Resumes, error)
	Read(id string) domain.Resume
	Create(resume domain.Resume, mime_type string, bytes []byte) (domain.Resume, error)
	Update(resume domain.Resume) (domain.Resume, error)
	Publish(id string) (domain.Resume, error)
	Published() (domain.Resume, error)
	Delete(id string) error
}
