package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newNotes() in_port.NoteCRUDService {
	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	return service.NewNoteCRUDService(out_adapter.NewNoteFakeOutAdapter(), revisions, activity)
}

// Notes share the generic revision + status machinery with projects; this
// confirms the note path is wired the same way.
func TestNoteRestoreRollsBackAndStaysAuditable(t *testing.T) {
	notes := newNotes()

	saved, _ := notes.Create(domain.Note{Title: "Draft thought"})
	notes.Update(domain.Note{Id: saved.Id, Title: "Better thought"})

	revs, _ := notes.Revisions(saved.Id, 100)
	originalRev := revs[len(revs)-1]

	restored, err := notes.Restore(saved.Id, originalRev.Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	if "Draft thought" != restored.Title {
		t.Fatalf("expected rollback to original title, got %q", restored.Title)
	}

	after, _ := notes.Revisions(saved.Id, 100)

	if 3 != len(after) {
		t.Fatalf("expected rollback to append a revision (3 total), got %d", len(after))
	}

	if 1 != countCurrent(after) {
		t.Fatalf("expected exactly one current revision after rollback")
	}
}

func TestNotePublishUnpublishTogglesStamp(t *testing.T) {
	notes := newNotes()

	saved, _ := notes.Create(domain.Note{Title: "A note"})

	published, _ := notes.Publish(saved.Id)

	if domain.StatusPublished != published.Status || "" == published.PublishedAt {
		t.Fatalf("publish must set status + published_at, got %q / %q", published.Status, published.PublishedAt)
	}

	unpublished, _ := notes.Unpublish(saved.Id)

	if domain.StatusDraft != unpublished.Status || "" != unpublished.PublishedAt {
		t.Fatalf("unpublish must reset to draft and clear published_at, got %q / %q", unpublished.Status, unpublished.PublishedAt)
	}
}
