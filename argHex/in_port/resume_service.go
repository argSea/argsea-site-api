package in_port

import "github.com/argSea/argsea-site-api/argHex/domain"

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
// clears whichever held it, and Published is what the hoist asks before it
// builds anything.
type ResumeService interface {
	List() (domain.Resumes, error)
	Read(id string) domain.Resume
	Create(resume domain.Resume, mime_type string, bytes []byte) (domain.Resume, error)
	Update(resume domain.Resume) (domain.Resume, error)
	Publish(id string) (domain.Resume, error)
	Published() (domain.Resume, error)
	Delete(id string) error
}
