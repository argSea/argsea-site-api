package out_port

import "github.com/argSea/portfolio_blog_api/argHex/domain"

//Skill repo to connect to a store
type SkillRepo interface {
	GetAll(limit int64, offset int64, sort interface{}) domain.Skills
	Get(id string) domain.Skill
	Set(skill domain.Skill) error
	Add(skill domain.Skill) (string, error)
	Remove(skill domain.Skill) error
}
