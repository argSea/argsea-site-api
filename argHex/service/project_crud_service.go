package service

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/utility"
)

// The stamp vocabulary is closed on purpose: ink is rendered into style
// attributes on the public site, and bluemonday only covers rich-text bodies,
// so this enum gate is the XSS boundary for stamp data.
var stampShapes = map[string]bool{"rect": true, "circle": true}
var stampMotifs = map[string]bool{"lighthouse": true, "boat": true, "sun": true, "wave": true, "moon": true, "anchor": true, "text": true}
var stampInks = map[string]bool{"#f0d9a8": true, "#93a0e8": true}

// stampCents matches the denomination printed on rect stamps: one or two
// digits followed by the cent sign, nothing else.
var stampCents = regexp.MustCompile(`^[0-9]{1,2}¢$`)

// validateStamp normalizes a stamp in place and checks it against the closed
// vocabulary. A nil stamp is valid — the site falls back to its default
// decoration. Cents is coupled to the rect shape and text to the text motif
// (operator ruling 2026-07-04): a stray field on the wrong variant is rejected
// rather than silently stored unrendered.
func validateStamp(stamp *domain.Stamp) error {
	if nil == stamp {
		return nil
	}

	// trim before measuring, so the store holds exactly what was validated
	stamp.Text = strings.TrimSpace(stamp.Text)

	if !stampShapes[stamp.Shape] {
		return errors.New("stamp shape must be rect or circle")
	}

	if !stampMotifs[stamp.Motif] {
		return errors.New("stamp motif must be one of lighthouse, boat, sun, wave, moon, anchor, text")
	}

	if !stampInks[stamp.Ink] {
		return errors.New("stamp ink must be #f0d9a8 or #93a0e8")
	}

	// the denomination belongs to rect stamps only
	if "" != stamp.Cents && "rect" != stamp.Shape {
		return errors.New("stamp cents is only valid on a rect stamp")
	}

	if "" != stamp.Cents && !stampCents.MatchString(stamp.Cents) {
		return errors.New("stamp cents must be one or two digits followed by ¢")
	}

	// the text motif requires words; every other motif forbids them
	if "text" == stamp.Motif && "" == stamp.Text {
		return errors.New("stamp text is required for the text motif")
	}

	if "text" != stamp.Motif && "" != stamp.Text {
		return errors.New("stamp text is only valid on the text motif")
	}

	if 40 < utf8.RuneCountInString(stamp.Text) {
		return errors.New("stamp text must be 40 characters or fewer")
	}

	return nil
}

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
	// an invalid stamp never reaches the store — reject before anything is written
	if err := validateStamp(project.Stamp); nil != err {
		return domain.Project{}, err
	}

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
	if err := p.snapshot(saved, "created"); nil != err {
		return domain.Project{}, err
	}

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

	// an invalid stamp never reaches the store — reject before anything is written
	if err := validateStamp(project.Stamp); nil != err {
		return domain.Project{}, err
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
	if err := p.snapshot(saved, "edited"); nil != err {
		return domain.Project{}, err
	}

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
	if err := p.snapshot(saved, "rolled back"); nil != err {
		return domain.Project{}, err
	}

	p.record("postcard \""+saved.Title+"\" rolled back", id)

	return saved, nil
}

// snapshot marshals the full document and records it as the new current
// revision. A failed snapshot fails the write — history must not silently
// diverge from the live document.
func (p projectCRUDService) snapshot(project domain.Project, verb string) error {
	data, err := json.Marshal(project)

	if nil != err {
		return err
	}

	_, err = p.revisions.Snapshot(domain.EntityProject, project.Id, string(data), verb+": "+project.Title)

	return err
}

func (p projectCRUDService) record(message string, id string) {
	if err := p.activity.Record(message, domain.EntityProject, id); nil != err {
		log.Printf("activity record failed for project %v: %v\n", id, err)
	}
}
