package out_adapter

import (
	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type projectFakeOutAdapter struct {
}

func NewProjectFakeOutAdapter() out_port.ProjectRepo {
	return projectFakeOutAdapter{}
}

func (p projectFakeOutAdapter) GetProjectsByUserID(id string) (domain.Projects, int64, error) {
	ps := domain.Projects{}
	pr := domain.Project{}

	pr.Id = "mooo"
	pr.UserIDs = append(pr.UserIDs, "mememe")
	pr.Type = "something"
	pr.Name = "Super duper project"
	*pr.ShortName = "SDP"
	pr.Icon = domain.SimpleImage{Source: "google.com"}
	pr.Slug = "super-duper-project"
	*pr.RepoURL = "github.com/argsea/super-duper-project"
	*pr.Skills = append(*pr.Skills, string("Skills"))
	*pr.Roles = append(*pr.Roles, "Only me")
	pr.Priority = 0
	pr.IsActive = false
	pr.IsReleased = false
	pr.RelatedCourse = &domain.Course{}
	pr.RelatedExperience = &domain.Experience{}
	pr.Links = append(pr.Links, domain.Link{URL: "github.com"})
	pr.Snippets = append(pr.Snippets, domain.Snippet{Name: "Test", Code: "<?php echo\"cheese\""})
	pr.Features = append(pr.Features, domain.Feature{})
	*pr.BookID = "12345"
	*pr.Description = "Something to describe"

	ps = append(ps, pr)

	return ps, 1, nil
}

func (p projectFakeOutAdapter) Get(id string) domain.Project {
	return domain.Project{}
}

func (p projectFakeOutAdapter) Set(proj domain.Project) error {
	return nil
}

func (p projectFakeOutAdapter) Add(proj domain.Project) (string, error) {
	return "", nil
}

func (p projectFakeOutAdapter) Remove(proj domain.Project) error {
	return nil
}
