package out_adapter

import (
	"fmt"
	"os"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/stores"
	"go.mongodb.org/mongo-driver/bson"
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
// build consumes). A limit of 0 means "no limit". The rack sort happens in the
// query so a limited read still takes the front of the rack, not an arbitrary
// subset; the service re-sorts, which is a stable no-op here.
func (p projectMongoAdapter) List(publishedOnly bool, limit int64) (domain.Projects, error) {
	var projects domain.Projects
	var err error

	sort := bson.D{{Key: "order", Value: 1}, {Key: "createdAt", Value: 1}}

	if publishedOnly {
		_, err = p.store.GetMany("status", domain.StatusPublished, limit, 0, sort, &projects)
	} else {
		_, err = p.store.GetAll(limit, 0, sort, &projects)
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
	project.Id = "" // replacement doc must not carry _id; mongo keeps the existing one
	return p.store.Replace(key, project)
}

func (p projectMongoAdapter) Remove(id string) error {
	return p.store.Delete(id)
}
