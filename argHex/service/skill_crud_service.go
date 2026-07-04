package service

import (
	"log"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/in_port"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type skillCRUDService struct {
	repo out_port.SkillRepo
}

func NewSkillCRUDService(repo out_port.SkillRepo) in_port.SkillCRUDService {
	return skillCRUDService{
		repo: repo,
	}
}

func (u skillCRUDService) Create(skill domain.Skill) (string, error) {
	skill_id, err := u.repo.Add(skill)

	if nil == err {
		log.Printf("Skill created with ID: %v\n", skill_id)
	} else {
		log.Printf("Skill not created. err: %v", err)
	}

	return skill_id, err
}

func (u skillCRUDService) Read(id string) domain.Skill {
	skillI := u.repo.Get(id)

	return skillI
}

func (u skillCRUDService) ReadAll(limit int64, offset int64, sort interface{}) domain.Skills {
	skills := u.repo.GetAll(limit, offset, sort)

	return skills
}

func (u skillCRUDService) Update(skill domain.Skill) error {
	err := u.repo.Set(skill)

	if nil == err {
		log.Printf("Skill updated, skill: %v\n", skill)
	} else {
		log.Printf("Skill not updated, error: %v\n", err)
	}

	return err
}

func (u skillCRUDService) Delete(skill domain.Skill) error {
	err := u.repo.Remove(skill)

	if nil == err {
		log.Printf("Skill deleted, skill: %v\n", skill)
	} else {
		log.Printf("Skill not deleted, error: %v\n", err)
	}

	return err
}
