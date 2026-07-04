package service

import (
	"log"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type resumeCRUDService struct {
	repo out_port.ResumeRepo
}

func NewResumeCRUDService(repo out_port.ResumeRepo) in_port.ResumeCRUDService {
	return resumeCRUDService{
		repo: repo,
	}
}

func (res resumeCRUDService) Create(resume domain.Resume) (string, error) {
	resume_id, err := res.repo.Add(resume)

	if nil == err {
		log.Printf("Resume created with ID: %v\n", resume_id)
	} else {
		log.Printf("Resume not created. err: %v", err)
	}

	return resume_id, err
}

func (res resumeCRUDService) Read(id string) domain.Resume {
	resu := res.repo.Get(id)

	return resu
}

func (res resumeCRUDService) Update(resume domain.Resume) error {
	err := res.repo.Set(resume)

	if nil == err {
		log.Printf("Resume updated, resume: %v\n", resume)
	} else {
		log.Printf("Resume not updated, error: %v\n", err)
	}

	return err
}

func (res resumeCRUDService) Delete(resume domain.Resume) error {
	err := res.repo.Remove(resume)

	if nil == err {
		log.Printf("Resume with ID %v deleted successfully\n", resume.Id)
	} else {
		log.Printf("Resume with ID %v could not be deleted, possible resume doesn't exist? err: %v\n", resume.Id, err)
	}

	return err
}
