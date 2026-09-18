package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ResumeFakeOutAdapter is an in-memory ResumeRepo for tests. Writes records
// every Set in the order it landed, as "<id>=<published>", because the publish
// transition's exclusivity rests on clearing the old cut before setting the new
// one and the finished state reads the same either way round.
type ResumeFakeOutAdapter struct {
	resumes map[string]domain.Resume
	seq     int
	Writes  []string
}

func NewResumeFakeOutAdapter() *ResumeFakeOutAdapter {
	return &ResumeFakeOutAdapter{
		resumes: map[string]domain.Resume{},
	}
}

func (r *ResumeFakeOutAdapter) List() (domain.Resumes, error) {
	var out domain.Resumes

	for _, resume := range r.resumes {
		out = append(out, resume)
	}

	return out, nil
}

func (r *ResumeFakeOutAdapter) Get(id string) domain.Resume {
	return r.resumes[id]
}

func (r *ResumeFakeOutAdapter) Add(resume domain.Resume) (string, error) {
	r.seq++
	id := fmt.Sprintf("resume-%d", r.seq)
	resume.Id = id
	r.resumes[id] = resume

	return id, nil
}

func (r *ResumeFakeOutAdapter) Set(resume domain.Resume) error {
	r.Writes = append(r.Writes, fmt.Sprintf("%s=%t", resume.Id, resume.Published))
	r.resumes[resume.Id] = resume

	return nil
}

func (r *ResumeFakeOutAdapter) Remove(id string) error {
	delete(r.resumes, id)

	return nil
}
