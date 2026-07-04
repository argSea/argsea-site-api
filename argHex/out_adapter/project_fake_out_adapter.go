package out_adapter

import (
	"fmt"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// projectFakeOutAdapter is an in-memory ProjectRepo for tests.
type projectFakeOutAdapter struct {
	projects *map[string]domain.Project
	seq      *int
}

func NewProjectFakeOutAdapter() out_port.ProjectRepo {
	return projectFakeOutAdapter{
		projects: &map[string]domain.Project{},
		seq:      new(int),
	}
}

func (p projectFakeOutAdapter) List(publishedOnly bool, limit int64) (domain.Projects, error) {
	var out domain.Projects

	for _, project := range *p.projects {
		if publishedOnly && domain.StatusPublished != project.Status {
			continue
		}

		out = append(out, project)
	}

	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (p projectFakeOutAdapter) Get(id string) domain.Project {
	return (*p.projects)[id]
}

func (p projectFakeOutAdapter) Add(project domain.Project) (string, error) {
	*p.seq++
	id := fmt.Sprintf("proj-%d", *p.seq)
	project.Id = id
	(*p.projects)[id] = project

	return id, nil
}

func (p projectFakeOutAdapter) Set(project domain.Project) error {
	(*p.projects)[project.Id] = project

	return nil
}

func (p projectFakeOutAdapter) Remove(id string) error {
	delete(*p.projects, id)

	return nil
}
