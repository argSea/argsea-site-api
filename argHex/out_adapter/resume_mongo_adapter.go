package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type resumeMongoAdapter struct {
	store *stores.Mordor
}

func NewResumeMongoAdapter(store *stores.Mordor) out_port.ResumeRepo {
	return resumeMongoAdapter{
		store: store,
	}
}

func (r resumeMongoAdapter) List() (domain.Resumes, error) {
	var resumes domain.Resumes
	_, err := r.store.GetAll(0, 0, nil, &resumes)

	return resumes, err
}

func (r resumeMongoAdapter) Get(id string) domain.Resume {
	var resume domain.Resume
	err := r.store.Get("_id", id, &resume)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Resume{}
	}

	return resume
}

func (r resumeMongoAdapter) Add(resume domain.Resume) (string, error) {
	resume.Id = ""
	return r.store.Write(resume)
}

func (r resumeMongoAdapter) Set(resume domain.Resume) error {
	key := resume.Id
	resume.Id = ""
	return r.store.Replace(key, resume)
}

func (r resumeMongoAdapter) Remove(id string) error {
	return r.store.Delete(id)
}
