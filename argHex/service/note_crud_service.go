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

type noteCRUDService struct {
	repo      out_port.NoteRepo
	revisions in_port.RevisionService
	activity  in_port.ActivityService
}

func NewNoteCRUDService(repo out_port.NoteRepo, revisions in_port.RevisionService, activity in_port.ActivityService) in_port.NoteCRUDService {
	return noteCRUDService{
		repo:      repo,
		revisions: revisions,
		activity:  activity,
	}
}

func (n noteCRUDService) List(publishedOnly bool, limit int64) (domain.Notes, error) {
	return n.repo.List(publishedOnly, limit)
}

func (n noteCRUDService) Read(id string) domain.Note {
	return n.repo.Get(id)
}

func (n noteCRUDService) Create(note domain.Note) (domain.Note, error) {
	now := nowStamp()

	note.Id = ""
	note.Body = utility.SanitizeHTML(note.Body)

	if "" == note.Status {
		note.Status = domain.StatusDraft
	}

	if domain.StatusPublished == note.Status && "" == note.PublishedAt {
		note.PublishedAt = now
	}

	note.CreatedAt = now
	note.UpdatedAt = now

	id, err := n.repo.Add(note)

	if nil != err {
		return domain.Note{}, err
	}

	saved := n.repo.Get(id)
	n.snapshot(saved, "created")
	n.record("note \""+saved.Title+"\" created", saved.Id)

	return saved, nil
}

// Update writes new content but leaves the publication lifecycle alone — status
// and published_at only move through Publish/Unpublish. Every edit snapshots the
// full document.
func (n noteCRUDService) Update(note domain.Note) (domain.Note, error) {
	existing := n.repo.Get(note.Id)

	if "" == existing.Id {
		return domain.Note{}, errors.New("note not found")
	}

	note.Body = utility.SanitizeHTML(note.Body)
	note.Status = existing.Status
	note.PublishedAt = existing.PublishedAt
	note.CreatedAt = existing.CreatedAt
	note.UpdatedAt = nowStamp()

	if err := n.repo.Set(note); nil != err {
		return domain.Note{}, err
	}

	saved := n.repo.Get(note.Id)
	n.snapshot(saved, "edited")
	n.record("note \""+saved.Title+"\" edited", saved.Id)

	return saved, nil
}

func (n noteCRUDService) Delete(id string) error {
	existing := n.repo.Get(id)

	if err := n.repo.Remove(id); nil != err {
		return err
	}

	n.record("note \""+existing.Title+"\" deleted", id)

	return nil
}

func (n noteCRUDService) Publish(id string) (domain.Note, error) {
	note := n.repo.Get(id)

	if "" == note.Id {
		return domain.Note{}, errors.New("note not found")
	}

	now := nowStamp()
	note.Status = domain.StatusPublished
	note.PublishedAt = now
	note.UpdatedAt = now

	if err := n.repo.Set(note); nil != err {
		return domain.Note{}, err
	}

	n.record("note \""+note.Title+"\" published", id)

	return n.repo.Get(id), nil
}

func (n noteCRUDService) Unpublish(id string) (domain.Note, error) {
	note := n.repo.Get(id)

	if "" == note.Id {
		return domain.Note{}, errors.New("note not found")
	}

	note.Status = domain.StatusDraft
	note.PublishedAt = ""
	note.UpdatedAt = nowStamp()

	if err := n.repo.Set(note); nil != err {
		return domain.Note{}, err
	}

	n.record("note \""+note.Title+"\" unpublished", id)

	return n.repo.Get(id), nil
}

func (n noteCRUDService) Revisions(id string, limit int64) (domain.Revisions, error) {
	return n.revisions.List(domain.EntityNote, id, limit)
}

// Restore rolls the live note back to an earlier revision's snapshot and copies
// that state forward as a new current revision, keeping the rollback auditable.
func (n noteCRUDService) Restore(id string, revisionID string) (domain.Note, error) {
	rev := n.revisions.Get(revisionID)

	if "" == rev.Id || rev.EntityId != id {
		return domain.Note{}, errors.New("revision not found for note")
	}

	var restored domain.Note

	if err := json.Unmarshal([]byte(rev.Snapshot), &restored); nil != err {
		return domain.Note{}, err
	}

	restored.Id = id
	restored.UpdatedAt = nowStamp()

	if err := n.repo.Set(restored); nil != err {
		return domain.Note{}, err
	}

	saved := n.repo.Get(id)
	n.snapshot(saved, "rolled back")
	n.record("note \""+saved.Title+"\" rolled back", id)

	return saved, nil
}

func (n noteCRUDService) snapshot(note domain.Note, verb string) {
	data, err := json.Marshal(note)

	if nil != err {
		log.Printf("note snapshot marshal failed for %v: %v\n", note.Id, err)
		return
	}

	if _, err := n.revisions.Snapshot(domain.EntityNote, note.Id, string(data), verb+": "+note.Title); nil != err {
		log.Printf("note snapshot failed for %v: %v\n", note.Id, err)
	}
}

func (n noteCRUDService) record(message string, id string) {
	if err := n.activity.Record(message, domain.EntityNote, id); nil != err {
		log.Printf("activity record failed for note %v: %v\n", id, err)
	}
}
