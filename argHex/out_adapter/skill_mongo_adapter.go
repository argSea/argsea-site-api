package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
)

type skillMongoAdapter struct {
	store *stores.Mordor
}

func NewSkillMongoAdapter(store *stores.Mordor) out_port.SkillRepo {
	s := skillMongoAdapter{
		store: store,
	}

	return s
}

func (s skillMongoAdapter) GetAll(limit int64, offset int64, sort interface{}) domain.Skills {
	var skills domain.Skills
	count, err := s.store.GetAll(limit, offset, nil, &skills)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return []domain.Skill{}
	}

	fmt.Printf("count: %v\n", count)

	return skills
}

func (s skillMongoAdapter) Get(id string) domain.Skill {
	var skill domain.Skill
	err := s.store.Get("_id", id, &skill)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Skill{}
	}

	return skill
}

func (s skillMongoAdapter) Set(skill domain.Skill) error {
	key := skill.Id
	skill.Id = ""

	err := s.store.Update(key, skill)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	return err
}

func (s skillMongoAdapter) Add(skill domain.Skill) (string, error) {
	skill.Id = ""

	id, err := s.store.Write(skill)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	return id, err
}

func (s skillMongoAdapter) Remove(skill domain.Skill) error {
	skill_id := skill.Id

	err := s.store.Delete(skill_id)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	return err
}
