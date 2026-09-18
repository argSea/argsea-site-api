package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// resumeFakeOutAdapter is an in-memory ResumeRepo for tests.
type resumeFakeOutAdapter struct {
	resumes *map[string]domain.Resume
	seq     *int
}

func NewResumeFakeOutAdapter() out_port.ResumeRepo {
	return resumeFakeOutAdapter{
		resumes: &map[string]domain.Resume{},
		seq:     new(int),
	}
}

func (r resumeFakeOutAdapter) List() (domain.Resumes, error) {
	var out domain.Resumes

	for _, resume := range *r.resumes {
		out = append(out, resume)
	}

	return out, nil
}

func (r resumeFakeOutAdapter) Get(id string) domain.Resume {
	return (*r.resumes)[id]
}

func (r resumeFakeOutAdapter) Add(resume domain.Resume) (string, error) {
	*r.seq++
	id := fmt.Sprintf("resume-%d", *r.seq)
	resume.Id = id
	(*r.resumes)[id] = resume

	return id, nil
}

func (r resumeFakeOutAdapter) Set(resume domain.Resume) error {
	(*r.resumes)[resume.Id] = resume

	return nil
}

func (r resumeFakeOutAdapter) Remove(id string) error {
	delete(*r.resumes, id)

	return nil
}
