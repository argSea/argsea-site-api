package service

import (
	"encoding/json"
	"errors"
	"log"
	"sort"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type caseLogCRUDService struct {
	repo      out_port.CaseLogRepo
	projects  out_port.ProjectRepo
	revisions in_port.RevisionService
	activity  in_port.ActivityService
}

// NewCaseLogCRUDService wires the case study desk onto its repo, a read-only
// handle to projects for the projectId existence and publish-slug checks, and
// the shared revision/activity plumbing, the same shape as the project rack.
func NewCaseLogCRUDService(repo out_port.CaseLogRepo, projects out_port.ProjectRepo, revisions in_port.RevisionService, activity in_port.ActivityService) in_port.CaseLogCRUDService {
	return caseLogCRUDService{
		repo:      repo,
		projects:  projects,
		revisions: revisions,
		activity:  activity,
	}
}

// validateProject enforces the required projectId reference. A log pinned to a
// project that doesn't exist would render nowhere, so it is rejected before
// anything is written.
func (c caseLogCRUDService) validateProject(projectId string) error {
	if "" == projectId {
		return errors.New("projectId is required")
	}

	if "" == c.projects.Get(projectId).Id {
		return errors.New("projectId must reference an existing project")
	}

	return nil
}

// List returns the logs oldest-first by creation time. Sorting here (not in the
// repo) keeps both adapters and both filters on one rule, the same as projects.
func (c caseLogCRUDService) List(publishedOnly bool, limit int64) (domain.CaseLogs, error) {
	logs, err := c.repo.List(publishedOnly, limit)

	if nil != err {
		return nil, err
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].CreatedAt < logs[j].CreatedAt
	})

	return logs, nil
}

func (c caseLogCRUDService) Read(id string) domain.CaseLog {
	return c.repo.Get(id)
}

// Create stores a new log as a draft. Publication never arrives through the
// door: status only moves to published through Publish, which enforces the
// one-published-per-project invariant. Every create snapshots the document.
func (c caseLogCRUDService) Create(log domain.CaseLog) (domain.CaseLog, error) {
	log.Id = ""

	if err := c.validateProject(log.ProjectId); nil != err {
		return domain.CaseLog{}, err
	}

	now := nowStamp()
	log.Status = domain.StatusDraft
	log.PublishedAt = ""
	log.CreatedAt = now
	log.UpdatedAt = now

	id, err := c.repo.Add(log)

	if nil != err {
		return domain.CaseLog{}, err
	}

	saved := c.repo.Get(id)

	if err := c.snapshot(saved, "created"); nil != err {
		return domain.CaseLog{}, err
	}

	c.record("case study \""+saved.Title+"\" created", saved.Id)

	return saved, nil
}

// Update writes new content but leaves the publication lifecycle alone; status
// and publishedAt only move through Publish/Unpublish, so an edit never silently
// publishes a draft. A published log cannot move to another project either:
// re-pointing it would plant a second published log on the target light behind
// Publish's back. Every edit snapshots the full document.
func (c caseLogCRUDService) Update(log domain.CaseLog) (domain.CaseLog, error) {
	existing := c.repo.Get(log.Id)

	if "" == existing.Id {
		return domain.CaseLog{}, errors.New("case study not found")
	}

	if err := c.validateProject(log.ProjectId); nil != err {
		return domain.CaseLog{}, err
	}

	if domain.StatusPublished == existing.Status && log.ProjectId != existing.ProjectId {
		return domain.CaseLog{}, errors.New("unpublish before moving a log to another light")
	}

	log.Status = existing.Status
	log.PublishedAt = existing.PublishedAt
	log.CreatedAt = existing.CreatedAt
	log.UpdatedAt = nowStamp()

	if err := c.repo.Set(log); nil != err {
		return domain.CaseLog{}, err
	}

	saved := c.repo.Get(log.Id)

	if err := c.snapshot(saved, "edited"); nil != err {
		return domain.CaseLog{}, err
	}

	c.record("case study \""+saved.Title+"\" edited", saved.Id)

	return saved, nil
}

func (c caseLogCRUDService) Delete(id string) error {
	existing := c.repo.Get(id)

	if err := c.repo.Remove(id); nil != err {
		return err
	}

	c.record("case study \""+existing.Title+"\" deleted", id)

	return nil
}

// Publish hoists this log as the project's published case study and lowers any
// other published log for the same project. Hoist first, lower after: a crash
// between the writes leaves two published (the site build can pick one) rather
// than none. The referenced project must exist and carry a slug, since the
// public route is /projects/<slug>.
func (c caseLogCRUDService) Publish(id string) (domain.CaseLog, error) {
	log := c.repo.Get(id)

	if "" == log.Id {
		return domain.CaseLog{}, errors.New("case study not found")
	}

	project := c.projects.Get(log.ProjectId)

	if "" == project.Id {
		return domain.CaseLog{}, errors.New("publish requires an existing project")
	}

	if "" == project.Slug {
		return domain.CaseLog{}, errors.New("publish requires the project to have a slug")
	}

	now := nowStamp()
	log.Status = domain.StatusPublished
	log.PublishedAt = now
	log.UpdatedAt = now

	if err := c.repo.Set(log); nil != err {
		return domain.CaseLog{}, err
	}

	others, err := c.repo.List(false, 0)

	if nil != err {
		return domain.CaseLog{}, err
	}

	for _, other := range others {
		if other.Id == log.Id || other.ProjectId != log.ProjectId || domain.StatusPublished != other.Status {
			continue
		}

		other.Status = domain.StatusDraft
		other.PublishedAt = ""
		other.UpdatedAt = now

		if err := c.repo.Set(other); nil != err {
			return domain.CaseLog{}, err
		}
	}

	c.record("case study \""+log.Title+"\" published", id)

	return c.repo.Get(id), nil
}

func (c caseLogCRUDService) Unpublish(id string) (domain.CaseLog, error) {
	log := c.repo.Get(id)

	if "" == log.Id {
		return domain.CaseLog{}, errors.New("case study not found")
	}

	log.Status = domain.StatusDraft
	log.PublishedAt = ""
	log.UpdatedAt = nowStamp()

	if err := c.repo.Set(log); nil != err {
		return domain.CaseLog{}, err
	}

	c.record("case study \""+log.Title+"\" unpublished", id)

	return c.repo.Get(id), nil
}

func (c caseLogCRUDService) Revisions(id string, limit int64) (domain.Revisions, error) {
	return c.revisions.List(domain.EntityCaseLog, id, limit)
}

// Restore rolls the live log back to an earlier revision's snapshot and copies
// that state forward as a new current revision, so the rollback is itself an
// auditable printing in the history, the same as projects. A rollback is an
// edit too: it restores content only and keeps the live document's publication
// lifecycle, so a published-era snapshot restored onto a now-draft log never
// resurrects a second published log behind Publish's back.
func (c caseLogCRUDService) Restore(id string, revisionID string) (domain.CaseLog, error) {
	existing := c.repo.Get(id)

	if "" == existing.Id {
		return domain.CaseLog{}, errors.New("case study not found")
	}

	rev := c.revisions.Get(revisionID)

	if "" == rev.Id || rev.EntityId != id {
		return domain.CaseLog{}, errors.New("revision not found for case study")
	}

	var restored domain.CaseLog

	if err := json.Unmarshal([]byte(rev.Snapshot), &restored); nil != err {
		return domain.CaseLog{}, err
	}

	// the same rule as Update: a published log stays pinned to its light, even
	// when the snapshot being restored pointed somewhere else
	if domain.StatusPublished == existing.Status && restored.ProjectId != existing.ProjectId {
		return domain.CaseLog{}, errors.New("unpublish before moving a log to another light")
	}

	restored.Id = id
	restored.Status = existing.Status
	restored.PublishedAt = existing.PublishedAt
	restored.UpdatedAt = nowStamp()

	if err := c.repo.Set(restored); nil != err {
		return domain.CaseLog{}, err
	}

	saved := c.repo.Get(id)

	if err := c.snapshot(saved, "rolled back"); nil != err {
		return domain.CaseLog{}, err
	}

	c.record("case study \""+saved.Title+"\" rolled back", id)

	return saved, nil
}

// snapshot marshals the full document and records it as the new current
// revision. A failed snapshot fails the write; history must not silently diverge
// from the live document.
func (c caseLogCRUDService) snapshot(log domain.CaseLog, verb string) error {
	data, err := json.Marshal(log)

	if nil != err {
		return err
	}

	_, err = c.revisions.Snapshot(domain.EntityCaseLog, log.Id, string(data), verb+": "+log.Title)

	return err
}

func (c caseLogCRUDService) record(message string, id string) {
	if err := c.activity.Record(message, domain.EntityCaseLog, id); nil != err {
		log.Printf("activity record failed for case study %v: %v\n", id, err)
	}
}
