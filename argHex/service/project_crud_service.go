package service

import (
	"log"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type projectCRUDService struct {
	repo out_port.ProjectRepo
}

func NewProjectCRUDService(repo out_port.ProjectRepo) in_port.ProjectCRUDService {
	return projectCRUDService{
		repo: repo,
	}
}

func (p projectCRUDService) Create(project domain.Project) (string, error) {
	proj_id, err := p.repo.Add(project)

	if nil == err {
		log.Printf("Project created with ID: %v\n", proj_id)
	} else {
		log.Printf("Project not created. err: %v", err)
	}

	return proj_id, err
}

func (p projectCRUDService) GetByUserID(userID string) ([]domain.Project, int64, error) {
	projects, count, err := p.repo.GetProjectsByUserID(userID)

	return projects, count, err
}

func (p projectCRUDService) Read(id string) domain.Project {
	project := p.repo.Get(id)

	return project
}

func (p projectCRUDService) Update(project domain.Project) error {
	err := p.repo.Set(project)

	if nil == err {
		log.Printf("Project updated, resume: %v\n", project)
	} else {
		log.Printf("Project not updated, error: %v\n", err)
	}

	return err
}

func (p projectCRUDService) Delete(project domain.Project) error {
	err := p.repo.Remove(project)

	if nil == err {
		log.Printf("Project with ID %v deleted successfully\n", project.Id)
	} else {
		log.Printf("Project with ID %v could not be deleted, possible project doesn't exist? err: %v\n", project.Id, err)
	}

	return err
}
