package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
)

type resumeMongoAdapter struct {
	store *stores.Mordor
}

func NewResumeMongoAdapter(store *stores.Mordor) out_port.ResumeRepo {
	r := resumeMongoAdapter{
		store: store,
	}

	return r
}

func (res resumeMongoAdapter) GetByUserID(id string) (domain.Resumes, int64, error) {
	var resumes domain.Resumes
	count, err := res.store.GetMany("userID", id, 50, 0, nil, &resumes)

	return resumes, count, err
}

func (res resumeMongoAdapter) Get(id string) domain.Resume {
	var resume domain.Resume
	err := res.store.Get("_id", id, &resume)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Resume{}
	}

	return resume
}

func (res resumeMongoAdapter) Set(resume domain.Resume) error {
	key := resume.Id
	resume.Id = "" //unset so mongo doesn't try to set it

	err := res.store.Update(key, resume)

	return err
}

func (res resumeMongoAdapter) Add(resume domain.Resume) (string, error) {
	resume.Id = "" //make sure it wasn't set

	new_id, err := res.store.Write(resume)

	return new_id, err
}

func (res resumeMongoAdapter) Remove(resume domain.Resume) error {
	res_id := resume.Id

	err := res.store.Delete(res_id)

	return err
}
