package service_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newProjects wires a project service onto in-memory fakes for repo, revisions,
// and activity, so the real business logic runs end-to-end.
func newProjects() in_port.ProjectCRUDService {
	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	return service.NewProjectCRUDService(out_adapter.NewProjectFakeOutAdapter(), revisions, activity)
}

func TestCreateSnapshotsAndDefaultsToDraft(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{Title: "One-off"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if domain.StatusDraft != saved.Status {
		t.Fatalf("expected new project to default to draft, got %q", saved.Status)
	}

	if "" != saved.PublishedAt {
		t.Fatalf("a draft must not have a published_at, got %q", saved.PublishedAt)
	}

	revs, _ := projects.Revisions(saved.Id, 100)

	if 1 != len(revs) {
		t.Fatalf("expected create to record 1 revision, got %d", len(revs))
	}
}

func TestUpdateSnapshotsEachEdit(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "First"})
	projects.Update(domain.Project{Id: saved.Id, Title: "Second"})
	projects.Update(domain.Project{Id: saved.Id, Title: "Third"})

	revs, _ := projects.Revisions(saved.Id, 100)

	// one create + two edits
	if 3 != len(revs) {
		t.Fatalf("expected 3 revisions after create + 2 edits, got %d", len(revs))
	}

	if !revs[0].IsCurrent {
		t.Fatalf("the newest revision should be current")
	}
}

func TestUpdateLeavesPublicationLifecycleAlone(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Ship it"})
	published, _ := projects.Publish(saved.Id)

	// an edit that carries a stale/blank status must not un-publish the project
	edited, err := projects.Update(domain.Project{Id: saved.Id, Title: "Ship it (typo fix)", Status: domain.StatusDraft})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if domain.StatusPublished != edited.Status {
		t.Fatalf("edit should preserve published status, got %q", edited.Status)
	}

	if edited.PublishedAt != published.PublishedAt {
		t.Fatalf("edit should preserve published_at %q, got %q", published.PublishedAt, edited.PublishedAt)
	}
}

func TestPublishAndUnpublishToggleTheStamp(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Draft first"})

	published, err := projects.Publish(saved.Id)

	if nil != err {
		t.Fatalf("publish failed: %v", err)
	}

	if domain.StatusPublished != published.Status || "" == published.PublishedAt {
		t.Fatalf("publish must set status and a published_at, got %q / %q", published.Status, published.PublishedAt)
	}

	unpublished, err := projects.Unpublish(saved.Id)

	if nil != err {
		t.Fatalf("unpublish failed: %v", err)
	}

	if domain.StatusDraft != unpublished.Status {
		t.Fatalf("unpublish must return to draft, got %q", unpublished.Status)
	}

	if "" != unpublished.PublishedAt {
		t.Fatalf("unpublish must clear published_at, got %q", unpublished.PublishedAt)
	}
}

func TestListPublishedOnlyFilters(t *testing.T) {
	projects := newProjects()

	shown, _ := projects.Create(domain.Project{Title: "Shown"})
	projects.Create(domain.Project{Title: "Hidden"})
	projects.Publish(shown.Id)

	published, _ := projects.List(true, 0)

	if 1 != len(published) {
		t.Fatalf("expected only published projects, got %d", len(published))
	}

	if "Shown" != published[0].Title {
		t.Fatalf("expected the published project, got %q", published[0].Title)
	}

	all, _ := projects.List(false, 0)

	if 2 != len(all) {
		t.Fatalf("expected all projects without the filter, got %d", len(all))
	}
}

func TestRestoreRollsBackAndStaysAuditable(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Original"})
	projects.Update(domain.Project{Id: saved.Id, Title: "Revised"})

	// newest-first: [0] = revised edit, [1] = original create
	revs, _ := projects.Revisions(saved.Id, 100)
	originalRev := revs[len(revs)-1]

	restored, err := projects.Restore(saved.Id, originalRev.Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	if "Original" != restored.Title {
		t.Fatalf("expected live document to roll back to Original, got %q", restored.Title)
	}

	// the live document reflects the rollback
	if "Original" != projects.Read(saved.Id).Title {
		t.Fatalf("stored document was not rolled back")
	}

	// rollback copied the old snapshot forward as a new current revision
	after, _ := projects.Revisions(saved.Id, 100)

	if 3 != len(after) {
		t.Fatalf("expected rollback to append a new revision (3 total), got %d", len(after))
	}

	if 1 != countCurrent(after) || !after[0].IsCurrent {
		t.Fatalf("expected the rollback revision to be the sole current one")
	}
}

func TestRestoreRejectsForeignRevision(t *testing.T) {
	projects := newProjects()

	a, _ := projects.Create(domain.Project{Title: "A"})
	b, _ := projects.Create(domain.Project{Title: "B"})

	bRevs, _ := projects.Revisions(b.Id, 100)

	// restoring project A from project B's revision must be refused
	if _, err := projects.Restore(a.Id, bRevs[0].Id); nil == err {
		t.Fatalf("expected restore to reject a revision from another entity")
	}
}

func TestBodyIsSanitizedOnWrite(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{
		Title: "XSS attempt",
		Body:  `<p>hello</p><script>alert('x')</script>`,
	})

	if strings.Contains(saved.Body, "<script") {
		t.Fatalf("sanitizer should strip <script>, body was %q", saved.Body)
	}

	if !strings.Contains(saved.Body, "hello") {
		t.Fatalf("sanitizer should keep safe content, body was %q", saved.Body)
	}
}

func TestUpdateClearsEmptiedFields(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Keep", Moral: "a moral", Tags: []string{"one", "two"}})

	// an update that empties moral and tags must actually clear them —
	// replace semantics, not a $set merge
	cleared, err := projects.Update(domain.Project{Id: saved.Id, Title: "Keep"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if "" != cleared.Moral || 0 != len(cleared.Tags) {
		t.Fatalf("expected moral/tags cleared, got %q / %v", cleared.Moral, cleared.Tags)
	}
}

func TestRestoreClearsFieldsEmptyInSnapshot(t *testing.T) {
	projects := newProjects()

	// rev 1: moral empty — rev 2: moral filled
	saved, _ := projects.Create(domain.Project{Title: "Original"})
	projects.Update(domain.Project{Id: saved.Id, Title: "Original", Moral: "added later"})

	revs, _ := projects.Revisions(saved.Id, 100)
	restored, err := projects.Restore(saved.Id, revs[len(revs)-1].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	// the field filled in the live doc but empty in the snapshot must be empty
	// after restore, and the new "rolled back" revision must record it empty
	if "" != restored.Moral {
		t.Fatalf("expected restore to clear moral, got %q", restored.Moral)
	}

	after, _ := projects.Revisions(saved.Id, 1)
	var recorded domain.Project

	if err := json.Unmarshal([]byte(after[0].Snapshot), &recorded); nil != err {
		t.Fatalf("could not parse rolled-back snapshot: %v", err)
	}

	if "" != recorded.Moral {
		t.Fatalf("rolled-back revision must record moral empty, got %q", recorded.Moral)
	}
}

// validStamp returns a stamp that passes every gate, for tests to mutate.
func validStamp() *domain.Stamp {
	return &domain.Stamp{Shape: "rect", Motif: "lighthouse", Ink: "#f0d9a8", Cents: "3¢"}
}

func TestValidStampAcceptedOnCreateAndUpdate(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{Title: "Stamped", Stamp: validStamp()})

	if nil != err {
		t.Fatalf("create with a valid stamp failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Stamp || "lighthouse" != stored.Stamp.Motif || "3¢" != stored.Stamp.Cents {
		t.Fatalf("expected the stamp persisted on create, got %+v", stored.Stamp)
	}

	// swap to the other shape/ink and a text motif on update
	edited, err := projects.Update(domain.Project{
		Id:    saved.Id,
		Title: "Stamped",
		Stamp: &domain.Stamp{Shape: "circle", Motif: "text", Ink: "#93a0e8", Text: "DAILY SINCE 1786"},
	})

	if nil != err {
		t.Fatalf("update with a valid stamp failed: %v", err)
	}

	if nil == edited.Stamp || "circle" != edited.Stamp.Shape || "DAILY SINCE 1786" != edited.Stamp.Text {
		t.Fatalf("expected the stamp replaced on update, got %+v", edited.Stamp)
	}
}

func TestStampRejectsOutOfSetEnumValues(t *testing.T) {
	projects := newProjects()

	// each enum field in turn, mutated to a value outside its set — uppercase
	// ink included, because the contract is exact lowercase strings
	cases := map[string]func(s *domain.Stamp){
		"shape": func(s *domain.Stamp) { s.Shape = "triangle" },
		"motif": func(s *domain.Stamp) { s.Motif = "kraken" },
		"ink":   func(s *domain.Stamp) { s.Ink = "#F0D9A8" },
	}

	for field, corrupt := range cases {
		stamp := validStamp()
		corrupt(stamp)

		if _, err := projects.Create(domain.Project{Title: "Bad " + field, Stamp: stamp}); nil == err {
			t.Fatalf("expected create to reject an out-of-set %s", field)
		}
	}

	// none of the rejected creates may have written anything
	all, _ := projects.List(false, 0)

	if 0 != len(all) {
		t.Fatalf("rejected creates must persist nothing, found %d projects", len(all))
	}

	// the update path rejects too, and the stored stamp survives untouched
	saved, _ := projects.Create(domain.Project{Title: "Good", Stamp: validStamp()})

	bad := validStamp()
	bad.Ink = "javascript:alert(1)"

	if _, err := projects.Update(domain.Project{Id: saved.Id, Title: "Good", Stamp: bad}); nil == err {
		t.Fatalf("expected update to reject an out-of-set ink")
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Stamp || "#f0d9a8" != stored.Stamp.Ink {
		t.Fatalf("rejected update must leave the stored stamp intact, got %+v", stored.Stamp)
	}
}

func TestStampCentsPattern(t *testing.T) {
	projects := newProjects()

	// one or two digits plus the cent sign; empty means "no denomination"
	for _, ok := range []string{"3¢", "12¢", ""} {
		stamp := validStamp()
		stamp.Cents = ok

		if _, err := projects.Create(domain.Project{Title: "Cents", Stamp: stamp}); nil != err {
			t.Fatalf("expected cents %q accepted, got %v", ok, err)
		}
	}

	for _, bad := range []string{"123¢", "3c", "¢", "3¢ ", "¢3"} {
		stamp := validStamp()
		stamp.Cents = bad

		if _, err := projects.Create(domain.Project{Title: "Cents", Stamp: stamp}); nil == err {
			t.Fatalf("expected cents %q rejected", bad)
		}
	}
}

func TestStampTextLengthCap(t *testing.T) {
	projects := newProjects()

	// exactly 40 characters is fine, and surrounding whitespace doesn't count
	stamp := validStamp()
	stamp.Text = "  " + strings.Repeat("a", 40) + "  "

	if _, err := projects.Create(domain.Project{Title: "Cap", Stamp: stamp}); nil != err {
		t.Fatalf("expected a 40-char (after trim) text accepted, got %v", err)
	}

	over := validStamp()
	over.Text = strings.Repeat("a", 41)

	if _, err := projects.Create(domain.Project{Title: "Cap", Stamp: over}); nil == err {
		t.Fatalf("expected a 41-char text rejected")
	}
}

func TestAbsentStampIsValid(t *testing.T) {
	projects := newProjects()

	// no stamp at all is the valid default state — the site falls back to its
	// default decoration
	saved, err := projects.Create(domain.Project{Title: "Unstamped"})

	if nil != err {
		t.Fatalf("create without a stamp failed: %v", err)
	}

	if nil != saved.Stamp {
		t.Fatalf("expected no stamp on the saved project, got %+v", saved.Stamp)
	}
}

func TestStampRoundTripsThroughSnapshotRestore(t *testing.T) {
	projects := newProjects()

	// rev 1: lighthouse rect — rev 2: text circle
	saved, _ := projects.Create(domain.Project{Title: "Round trip", Stamp: validStamp()})
	projects.Update(domain.Project{
		Id:    saved.Id,
		Title: "Round trip",
		Stamp: &domain.Stamp{Shape: "circle", Motif: "text", Ink: "#93a0e8", Text: "AIR MAIL"},
	})

	revs, _ := projects.Revisions(saved.Id, 100)
	restored, err := projects.Restore(saved.Id, revs[len(revs)-1].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	// the original stamp comes back exactly as snapshotted
	if nil == restored.Stamp || "rect" != restored.Stamp.Shape || "lighthouse" != restored.Stamp.Motif || "3¢" != restored.Stamp.Cents {
		t.Fatalf("expected restore to bring the original stamp back, got %+v", restored.Stamp)
	}
}

func TestUpdateWithoutStampClearsIt(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Clear me", Stamp: validStamp()})

	// PUT is full-replace: a stamp omitted from the update is a stamp removed
	cleared, err := projects.Update(domain.Project{Id: saved.Id, Title: "Clear me"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if nil != cleared.Stamp {
		t.Fatalf("expected the stamp cleared by an update without one, got %+v", cleared.Stamp)
	}

	if nil != projects.Read(saved.Id).Stamp {
		t.Fatalf("stored document should have no stamp after the clearing update")
	}
}
