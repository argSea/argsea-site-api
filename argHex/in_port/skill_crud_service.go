package in_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

//Skill service for CRUD
type SkillCRUDService interface {
	Create(skill domain.Skill) (string, error)
	Read(id string) domain.Skill
	ReadAll(limit int64, offset int64, sort interface{}) domain.Skills
	Update(skill domain.Skill) error
	Delete(skill domain.Skill) error
}
