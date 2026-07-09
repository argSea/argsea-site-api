package service

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"sort"
	"strconv"
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
// vocabulary. A nil stamp is valid; the site falls back to its default
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

// The light vocabulary is closed for the same reason as the stamp's: kind and
// color select animation names and glow colors rendered into style attributes
// on the public site, so this enum gate is the injection boundary for light
// data.
var lightKinds = map[string]bool{
	"fixed": true, "flash": true, "occult": true, "iso": true,
	"quick": true, "veryquick": true, "morse": true,
}
var lightColors = map[string]bool{"white": true, "red": true, "green": true}

// The kinds whose cycle the keeper sets. Fixed holds steady, and quick and
// veryquick blink at rates set by convention (roughly 60 and 120 flashes a
// minute), so a stored period on any of those would sit unrendered.
var lightPeriodKinds = map[string]bool{"flash": true, "occult": true, "iso": true, "morse": true}

// validateLight normalizes a light in place and checks it against the closed
// vocabulary. A nil light is valid; the site burns it as the default fixed
// white. The period is coupled to the keeper-timed kinds the way stamp cents
// is coupled to the rect shape: a period on a convention-timed kind is
// rejected rather than silently stored unrendered. The letter is to morse
// what text is to the text stamp: required there, meaningless anywhere else.
func validateLight(light *domain.Light) error {
	if nil == light {
		return nil
	}

	// trim before measuring, so the store holds exactly what was validated
	light.Extinguished = strings.TrimSpace(light.Extinguished)
	light.Letter = strings.ToUpper(strings.TrimSpace(light.Letter))

	if !lightKinds[light.Kind] {
		return errors.New("light kind must be one of fixed, flash, occult, iso, quick, veryquick, morse")
	}

	if !lightColors[light.Color] {
		return errors.New("light color must be white, red, or green")
	}

	// fixed holds steady and quick/veryquick keep convention time; a period
	// on any of them would be stored unrendered
	if !lightPeriodKinds[light.Kind] && 0 != light.Period {
		return errors.New("light period is only valid on flash, occult, iso, or morse")
	}

	// every keeper-timed kind needs a cycle the site can actually animate
	if lightPeriodKinds[light.Kind] && (2 > light.Period || 30 < light.Period) {
		return errors.New("light period must be 2 to 30 seconds")
	}

	// a morse letter needs room for its pattern; two seconds cannot fit one
	if "morse" == light.Kind && 4 > light.Period {
		return errors.New("light period must be 4 to 30 seconds for morse")
	}

	if "morse" == light.Kind && (1 != len(light.Letter) || 'A' > light.Letter[0] || 'Z' < light.Letter[0]) {
		return errors.New("light letter must be a single letter A to Z for morse")
	}

	if "morse" != light.Kind && "" != light.Letter {
		return errors.New("light letter is only valid on the morse kind")
	}

	if 40 < utf8.RuneCountInString(light.Extinguished) {
		return errors.New("light extinguished must be 40 characters or fewer")
	}

	return nil
}

// validateImages trims gallery entries in place. An empty name would 404 on
// the public site, so it is rejected rather than silently dropped; the cap is
// a backstop against a runaway gallery, not a design constraint.
func validateImages(images []string) error {
	if 12 < len(images) {
		return errors.New("images is capped at 12 prints")
	}

	for i, name := range images {
		images[i] = strings.TrimSpace(name)

		if "" == images[i] {
			return errors.New("images must not contain an empty name")
		}
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

// List returns the rack in display order: order asc, ties broken by createdAt
// asc. Sorting here (not in the repo) keeps both adapters and both filters on
// one rule.
func (p projectCRUDService) List(publishedOnly bool, limit int64) (domain.Projects, error) {
	projects, err := p.repo.List(publishedOnly, limit)

	if nil != err {
		return nil, err
	}

	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Order != projects[j].Order {
			return projects[i].Order < projects[j].Order
		}

		return projects[i].CreatedAt < projects[j].CreatedAt
	})

	return projects, nil
}

func (p projectCRUDService) Read(id string) domain.Project {
	return p.repo.Get(id)
}

func (p projectCRUDService) Create(project domain.Project) (domain.Project, error) {
	// an invalid stamp or light never reaches the store; reject before
	// anything is written
	if err := validateStamp(project.Stamp); nil != err {
		return domain.Project{}, err
	}

	if err := validateLight(project.Light); nil != err {
		return domain.Project{}, err
	}

	if err := validateImages(project.Images); nil != err {
		return domain.Project{}, err
	}

	now := nowStamp()

	// body is rich text; it lives in the store already sanitized
	project.Id = ""
	project.Body = utility.SanitizeHTML(project.Body)
	project.FirstLit = strings.TrimSpace(project.FirstLit)

	if "" == project.Status {
		project.Status = domain.StatusDraft
	}

	// a project created straight into "published" gets its stamp now
	if domain.StatusPublished == project.Status && "" == project.PublishedAt {
		project.PublishedAt = now
	}

	// rack placement is server-assigned: a new light lands at the end, and
	// nothing reaches the front window except through the feature endpoint
	order, orderErr := p.nextOrder()

	if nil != orderErr {
		return domain.Project{}, orderErr
	}

	project.Order = order
	project.Featured = false

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

	p.record("light \""+saved.Title+"\" created", saved.Id)

	return saved, nil
}

// Update writes new content but leaves the publication lifecycle alone; status
// and published_at only move through Publish/Unpublish, so an edit never
// silently publishes a draft. Every edit snapshots the full document.
func (p projectCRUDService) Update(project domain.Project) (domain.Project, error) {
	existing := p.repo.Get(project.Id)

	if "" == existing.Id {
		return domain.Project{}, errors.New("project not found")
	}

	// an invalid stamp or light never reaches the store; reject before
	// anything is written
	if err := validateStamp(project.Stamp); nil != err {
		return domain.Project{}, err
	}

	if err := validateLight(project.Light); nil != err {
		return domain.Project{}, err
	}

	if err := validateImages(project.Images); nil != err {
		return domain.Project{}, err
	}

	project.Body = utility.SanitizeHTML(project.Body)
	project.FirstLit = strings.TrimSpace(project.FirstLit)
	project.Status = existing.Status
	project.PublishedAt = existing.PublishedAt
	project.Order = existing.Order
	project.Featured = existing.Featured
	project.WallPos = existing.WallPos
	project.CreatedAt = existing.CreatedAt
	project.UpdatedAt = nowStamp()

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	saved := p.repo.Get(project.Id)
	if err := p.snapshot(saved, "edited"); nil != err {
		return domain.Project{}, err
	}

	p.record("light \""+saved.Title+"\" edited", saved.Id)

	return saved, nil
}

func (p projectCRUDService) Delete(id string) error {
	existing := p.repo.Get(id)

	if err := p.repo.Remove(id); nil != err {
		return err
	}

	p.record("light \""+existing.Title+"\" deleted", id)

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

	p.record("light \""+project.Title+"\" published", id)

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

	p.record("light \""+project.Title+"\" unpublished", id)

	return p.repo.Get(id), nil
}

// Reorder moves the light to a new rack position. Lifecycle-style like
// Publish: activity-logged but never snapshotted; reordering the rack must
// not spam the revision history.
func (p projectCRUDService) Reorder(id string, order int) (domain.Project, error) {
	project := p.repo.Get(id)

	if "" == project.Id {
		return domain.Project{}, errors.New("project not found")
	}

	project.Order = order
	project.UpdatedAt = nowStamp()

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	p.record("light \""+project.Title+"\" reordered to "+strconv.Itoa(order), id)

	return p.repo.Get(id), nil
}

// Arrangement pins each listed light to its coast position. Non-destructive:
// a project left out of the batch keeps whatever position it already has.
// Lifecycle-style like Reorder: activity-logged but never snapshotted;
// dragging lights along the coast must not spam the revision history.
func (p projectCRUDService) Arrangement(placements []domain.WallPlacement) ([]domain.Project, error) {
	saved := make([]domain.Project, 0, len(placements))

	for _, placement := range placements {
		project := p.repo.Get(placement.Id)

		if "" == project.Id {
			return nil, errors.New("project not found: " + placement.Id)
		}

		project.WallPos = &domain.WallPos{X: placement.X, Y: placement.Y, Rotation: placement.Rotation}
		project.UpdatedAt = nowStamp()

		if err := p.repo.Set(project); nil != err {
			return nil, err
		}

		saved = append(saved, p.repo.Get(placement.Id))
	}

	p.record("coast arrangement saved: "+strconv.Itoa(len(placements))+" light(s) pinned", "")

	return saved, nil
}

// Feature puts the light in the front window. No cap here; the admin enforces
// the window-fits-three rule; no snapshot either, same as Reorder.
func (p projectCRUDService) Feature(id string) (domain.Project, error) {
	return p.setFeatured(id, true, "featured")
}

// Unfeature takes the light out of the front window.
func (p projectCRUDService) Unfeature(id string) (domain.Project, error) {
	return p.setFeatured(id, false, "unfeatured")
}

func (p projectCRUDService) setFeatured(id string, featured bool, verb string) (domain.Project, error) {
	project := p.repo.Get(id)

	if "" == project.Id {
		return domain.Project{}, errors.New("project not found")
	}

	project.Featured = featured
	project.UpdatedAt = nowStamp()

	if err := p.repo.Set(project); nil != err {
		return domain.Project{}, err
	}

	p.record("light \""+project.Title+"\" "+verb, id)

	return p.repo.Get(id), nil
}

// nextOrder places a new light after everything already on the rack:
// max(order)+1 across all projects, published or not. A failed list fails the
// create; silently defaulting would collide at the front of the rack.
func (p projectCRUDService) nextOrder() (int, error) {
	projects, err := p.repo.List(false, 0)

	if nil != err {
		return 0, err
	}

	max := 0

	for _, project := range projects {
		if project.Order > max {
			max = project.Order
		}
	}

	return max + 1, nil
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

	p.record("light \""+saved.Title+"\" rolled back", id)

	return saved, nil
}

// snapshot marshals the full document and records it as the new current
// revision. A failed snapshot fails the write; history must not silently
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
