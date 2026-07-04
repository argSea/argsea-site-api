package out_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

type ResumeRepo interface {
	GetByUserID(id string) (domain.Resumes, int64, error)
	Get(id string) domain.Resume
	Set(resume domain.Resume) error
	Add(resume domain.Resume) (string, error)
	Remove(resume domain.Resume) error
}
