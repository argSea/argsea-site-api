package service

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/utility"
)

type projectCRUDService struct {
	repo      out_port.ProjectRepo
	revisions in_port.RevisionService
	activity  in_port.ActivityService
}

func NewProjectCRUDService(repo out_port.ProjectRepo, revisions in_port.RevisionService, activity in_port.ActivityService) in_port.ProjectCRUDService {
	return projectCRUDService{
		repo:      repo,
		revisions: revisions,
		activity:  activity,
	}
}

func (p projectCRUDService) List(publishedOnly bool, limit int64) (domain.Projects, error) {
	return p.repo.List(publishedOnly, limit)
}

func (p projectCRUDService) Read(id string) domain.Project {
	return p.repo.Get(id)
}

func (p projectCRUDService) Create(project domain.Project) (domain.Project, error) {
	now := nowStamp()

	// body is rich text — it lives in the store already sanitized
	project.Id = ""
	project.Body = utility.SanitizeHTML(project.Body)

	if "" == project.Status {
		project.Status = domain.StatusDraft
	}

	// a project created straight into "published" gets its stamp now
	if domain.StatusPublished == project.Status && "" == project.PublishedAt {
		project.PublishedAt = now
	}

	project.CreatedAt = now
	project.UpdatedAt = now

	id, err := p.repo.Add(project)

	if nil != err {
		return domain.Project{}, err
	}

	saved := p.repo.Get(id)
	p.snapshot(saved, "created")
	p.record("postcard \""+saved.Title+"\" created", saved.Id)

	return saved, nil
}

// Update writes new content but leaves the publication lifecycle alone — status
// and published_at only move through Publish/Unpublish, so an edit never
// silently publishes a draft. Every edit snapshots the full document.
func (p projectCRUDService) Update(project domain.Project) (domain.Project, error) {
	existing := p.repo.Get(project.Id)

	if "" == existing.Id {
		return domain.Project{}, errors.New("project not found")
	}

	project.Body = utility.SanitizeHTML(project.Body)
	project.Status = existing.Status
	project.PublishedAt = existing.PublishedAt
	project.CreatedAt = existing.CreatedAt
	project.UpdatedAt = nowStamp()

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	saved := p.repo.Get(project.Id)
	p.snapshot(saved, "edited")
	p.record("postcard \""+saved.Title+"\" edited", saved.Id)

	return saved, nil
}

func (p projectCRUDService) Delete(id string) error {
	existing := p.repo.Get(id)

	if err := p.repo.Remove(id); nil != err {
		return err
	}

	p.record("postcard \""+existing.Title+"\" deleted", id)

	return nil
}

func (p projectCRUDService) Publish(id string) (domain.Project, error) {
	project := p.repo.Get(id)

	if "" == project.Id {
		return domain.Project{}, errors.New("project not found")
	}

	now := nowStamp()
	project.Status = domain.StatusPublished
	project.PublishedAt = now
	project.UpdatedAt = now

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	p.record("postcard \""+project.Title+"\" published", id)

	return p.repo.Get(id), nil
}

func (p projectCRUDService) Unpublish(id string) (domain.Project, error) {
	project := p.repo.Get(id)

	if "" == project.Id {
		return domain.Project{}, errors.New("project not found")
	}

	project.Status = domain.StatusDraft
	project.PublishedAt = ""
	project.UpdatedAt = nowStamp()

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	p.record("postcard \""+project.Title+"\" unpublished", id)

	return p.repo.Get(id), nil
}

func (p projectCRUDService) Revisions(id string, limit int64) (domain.Revisions, error) {
	return p.revisions.List(domain.EntityProject, id, limit)
}

// Restore rolls the live document back to an earlier revision's snapshot and
// then copies that state forward as a new current revision, so the rollback is
// itself an auditable printing in the history.
func (p projectCRUDService) Restore(id string, revisionID string) (domain.Project, error) {
	rev := p.revisions.Get(revisionID)

	if "" == rev.Id || rev.EntityId != id {
		return domain.Project{}, errors.New("revision not found for project")
	}

	var restored domain.Project

	if err := json.Unmarshal([]byte(rev.Snapshot), &restored); nil != err {
		return domain.Project{}, err
	}

	restored.Id = id
	restored.UpdatedAt = nowStamp()

	if err := p.repo.Set(restored); nil != err {
		return domain.Project{}, err
	}

	saved := p.repo.Get(id)
	p.snapshot(saved, "rolled back")
	p.record("postcard \""+saved.Title+"\" rolled back", id)

	return saved, nil
}

// snapshot marshals the full document and records it as the new current
// revision. Snapshot failures are logged, not fatal — the write already landed.
func (p projectCRUDService) snapshot(project domain.Project, verb string) {
	data, err := json.Marshal(project)

	if nil != err {
		log.Printf("project snapshot marshal failed for %v: %v\n", project.Id, err)
		return
	}

	if _, err := p.revisions.Snapshot(domain.EntityProject, project.Id, string(data), verb+": "+project.Title); nil != err {
		log.Printf("project snapshot failed for %v: %v\n", project.Id, err)
	}
}

func (p projectCRUDService) record(message string, id string) {
	if err := p.activity.Record(message, domain.EntityProject, id); nil != err {
		log.Printf("activity record failed for project %v: %v\n", id, err)
	}
}
