package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// ResumeRepo is the metadata half of the resume shelf: the mongo documents that
// describe each stored cut. The PDFs themselves land through MediaRepo, the
// same file half the darkroom writes to.
type ResumeRepo interface {
	List() (domain.Resumes, error)
	Get(id string) domain.Resume
	Add(resume domain.Resume) (string, error)
	Set(resume domain.Resume) error
	Remove(id string) error
}
