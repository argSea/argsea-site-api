package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
	"github.com/argSea/portfolio_blog_api/argHex/stores"
)

type projectMongoAdapter struct {
	store *stores.Mordor
}

func NewProjectMongoAdapter(store *stores.Mordor) out_port.ProjectRepo {
	return projectMongoAdapter{
		store: store,
	}
}

func (p projectMongoAdapter) GetProjectsByUserID(id string) (domain.Projects, int64, error) {
	var projects domain.Projects
	count, err := p.store.GetMany("userIDs", id, 50, 0, nil, &projects)

	return projects, count, err
}

func (p projectMongoAdapter) Get(id string) domain.Project {
	var project domain.Project
	err := p.store.Get("_id", id, &project)

	if nil != err {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return domain.Project{}
	}

	return project
}

func (p projectMongoAdapter) Set(project domain.Project) error {
	key := project.Id
	project.Id = "" //unset so mongo doesn't try to set it

	err := p.store.Update(key, project)

	return err
}

func (p projectMongoAdapter) Add(project domain.Project) (string, error) {
	project.Id = "" //make sure it wasn't set

	new_id, err := p.store.Write(project)

	return new_id, err
}

func (p projectMongoAdapter) Remove(project domain.Project) error {
	res_id := project.Id

	err := p.store.Delete(res_id)

	return err
}
