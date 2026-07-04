package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
)

type projectMongoAdapter struct {
	store *stores.Mordor
}

func NewProjectMongoAdapter(store *stores.Mordor) out_port.ProjectRepo {
	return projectMongoAdapter{
		store: store,
	}
}

// List returns projects, optionally narrowed to published ones (what the Astro
// build consumes). A limit of 0 means "no limit".
func (p projectMongoAdapter) List(publishedOnly bool, limit int64) (domain.Projects, error) {
	var projects domain.Projects
	var err error

	if publishedOnly {
		_, err = p.store.GetMany("status", domain.StatusPublished, limit, 0, nil, &projects)
	} else {
		_, err = p.store.GetAll(limit, 0, nil, &projects)
	}

	return projects, err
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

func (p projectMongoAdapter) Add(project domain.Project) (string, error) {
	project.Id = "" // make sure it wasn't set
	return p.store.Write(project)
}

func (p projectMongoAdapter) Set(project domain.Project) error {
	key := project.Id
	project.Id = "" // unset so mongo doesn't try to set it
	return p.store.Update(key, project)
}

func (p projectMongoAdapter) Remove(id string) error {
	return p.store.Delete(id)
}
