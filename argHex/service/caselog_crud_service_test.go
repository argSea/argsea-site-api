package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newCaseLogs wires a caselog service onto in-memory fakes and hands back the
// project repo too, so a test can seed the projects a log must reference.
func newCaseLogs() (in_port.CaseLogCRUDService, out_port.ProjectRepo) {
	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	projects := out_adapter.NewProjectFakeOutAdapter()

	return service.NewCaseLogCRUDService(out_adapter.NewCaseLogFakeOutAdapter(), projects, revisions, activity), projects
}

// seedProject drops a project straight into the repo and returns its id, for a
// caselog to reference without going through the whole project service.
func seedProject(projects out_port.ProjectRepo, slug string) string {
	id, _ := projects.Add(domain.Project{Title: "A light", Slug: slug})
	return id
}

func TestCaseLogCreateDefaultsToDraftAndSnapshots(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	saved, err := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Story"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if domain.StatusDraft != saved.Status {
		t.Fatalf("expected a new log to default to draft, got %q", saved.Status)
	}

	if "" != saved.PublishedAt {
		t.Fatalf("a draft must not carry a published_at, got %q", saved.PublishedAt)
	}

	revs, _ := logs.Revisions(saved.Id, 100)

	if 1 != len(revs) {
		t.Fatalf("expected create to record 1 revision, got %d", len(revs))
	}
}

func TestCaseLogCreateRejectsMissingOrUnknownProject(t *testing.T) {
	logs, _ := newCaseLogs()

	if _, err := logs.Create(domain.CaseLog{Title: "No project"}); nil == err {
		t.Fatalf("expected a log without a projectId rejected")
	}

	if _, err := logs.Create(domain.CaseLog{ProjectId: "nope", Title: "Ghost"}); nil == err {
		t.Fatalf("expected a log pinned to an unknown project rejected")
	}
}

func TestCaseLogUpdateLeavesPublicationLifecycleAlone(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	saved, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Ship it"})
	published, _ := logs.Publish(saved.Id)

	// an edit carrying a stale draft status must not un-publish the log
	edited, err := logs.Update(domain.CaseLog{Id: saved.Id, ProjectId: projectID, Title: "Ship it (typo)", Status: domain.StatusDraft})

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

func TestCaseLogPublishRejectsAProjectWithoutASlug(t *testing.T) {
	logs, projects := newCaseLogs()
	noSlug := seedProject(projects, "")

	log, _ := logs.Create(domain.CaseLog{ProjectId: noSlug, Title: "Unreachable"})

	// the public route is /projects/<slug>; a published log needs one
	if _, err := logs.Publish(log.Id); nil == err {
		t.Fatalf("expected publish rejected when the project has no slug")
	}

	if domain.StatusPublished == logs.Read(log.Id).Status {
		t.Fatalf("the rejected publish must not have flipped the status")
	}
}

func TestCaseLogPublishSwapKeepsOnePublishedPerProject(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	first, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "First cut"})
	second, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Second cut"})

	if _, err := logs.Publish(first.Id); nil != err {
		t.Fatalf("publish first failed: %v", err)
	}

	if _, err := logs.Publish(second.Id); nil != err {
		t.Fatalf("publish second failed: %v", err)
	}

	// publishing the second must have pushed the first back to draft
	if domain.StatusDraft != logs.Read(first.Id).Status {
		t.Fatalf("expected the first log demoted to draft, got %q", logs.Read(first.Id).Status)
	}

	if domain.StatusPublished != logs.Read(second.Id).Status {
		t.Fatalf("expected the second log published, got %q", logs.Read(second.Id).Status)
	}

	// the invariant: at most one published log for the project
	published, _ := logs.List(true, 0)

	if 1 != len(published) {
		t.Fatalf("expected exactly one published log for the project, got %d", len(published))
	}
}

func TestCaseLogPublishSwapIsScopedToOneProject(t *testing.T) {
	logs, projects := newCaseLogs()
	projectA := seedProject(projects, "light-a")
	projectB := seedProject(projects, "light-b")

	a, _ := logs.Create(domain.CaseLog{ProjectId: projectA, Title: "A"})
	b, _ := logs.Create(domain.CaseLog{ProjectId: projectB, Title: "B"})

	logs.Publish(a.Id)
	logs.Publish(b.Id)

	// publishing B's log must not disturb A's: the swap is per project
	if domain.StatusPublished != logs.Read(a.Id).Status {
		t.Fatalf("expected A's log to stay published, got %q", logs.Read(a.Id).Status)
	}

	if domain.StatusPublished != logs.Read(b.Id).Status {
		t.Fatalf("expected B's log published, got %q", logs.Read(b.Id).Status)
	}
}

func TestCaseLogUnpublishReturnsToDraft(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	saved, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Live one"})
	logs.Publish(saved.Id)

	unpublished, err := logs.Unpublish(saved.Id)

	if nil != err {
		t.Fatalf("unpublish failed: %v", err)
	}

	if domain.StatusDraft != unpublished.Status || "" != unpublished.PublishedAt {
		t.Fatalf("expected unpublish to draft with no published_at, got %q / %q", unpublished.Status, unpublished.PublishedAt)
	}
}

func TestCaseLogRestoreRollsBackAndStaysAuditable(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	saved, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Original"})
	logs.Update(domain.CaseLog{Id: saved.Id, ProjectId: projectID, Title: "Revised"})

	// newest-first: the create is the last revision
	revs, _ := logs.Revisions(saved.Id, 100)
	restored, err := logs.Restore(saved.Id, revs[len(revs)-1].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	if "Original" != restored.Title {
		t.Fatalf("expected the log rolled back to Original, got %q", restored.Title)
	}

	// the rollback copied the old snapshot forward as a new revision
	after, _ := logs.Revisions(saved.Id, 100)

	if 3 != len(after) {
		t.Fatalf("expected create + edit + rollback = 3 revisions, got %d", len(after))
	}
}

func TestCaseLogRestorePreservesLifecycleAndTheOnePublishedInvariant(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	// snapshot X while it is published: publish doesn't snapshot, but the edit
	// right after records Status published into the revision
	x, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "X"})
	logs.Publish(x.Id)
	logs.Update(domain.CaseLog{Id: x.Id, ProjectId: projectID, Title: "X, published era"})

	// publishing Y lowers X to draft
	y, _ := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "Y"})
	logs.Publish(y.Id)

	// newest-first: [0] is the published-era edit of X
	revs, _ := logs.Revisions(x.Id, 100)
	restored, err := logs.Restore(x.Id, revs[0].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	// content came back, the lifecycle did not: X stays the draft it now is
	if "X, published era" != restored.Title {
		t.Fatalf("expected the snapshot content restored, got %q", restored.Title)
	}

	if domain.StatusDraft != restored.Status || "" != restored.PublishedAt {
		t.Fatalf("restore must not resurrect the snapshot's lifecycle, got %q / %q", restored.Status, restored.PublishedAt)
	}

	if domain.StatusPublished != logs.Read(y.Id).Status {
		t.Fatalf("expected Y to stay published, got %q", logs.Read(y.Id).Status)
	}

	// the invariant holds: exactly one published log for the project
	published, _ := logs.List(true, 0)

	if 1 != len(published) || "Y" != published[0].Title {
		t.Fatalf("expected Y as the sole published log, got %+v", published)
	}
}

func TestCaseLogUpdateRejectsMovingAPublishedLog(t *testing.T) {
	logs, projects := newCaseLogs()
	projectA := seedProject(projects, "light-a")
	projectB := seedProject(projects, "light-b")

	published, _ := logs.Create(domain.CaseLog{ProjectId: projectA, Title: "Pinned"})
	logs.Publish(published.Id)

	// a published log stays pinned to its light
	if _, err := logs.Update(domain.CaseLog{Id: published.Id, ProjectId: projectB, Title: "Pinned"}); nil == err {
		t.Fatalf("expected update to reject moving a published log to another project")
	}

	stored := logs.Read(published.Id)

	if projectA != stored.ProjectId || domain.StatusPublished != stored.Status {
		t.Fatalf("the rejected move must leave the log untouched, got %+v", stored)
	}

	// a draft moves freely
	draft, _ := logs.Create(domain.CaseLog{ProjectId: projectA, Title: "Loose"})

	moved, err := logs.Update(domain.CaseLog{Id: draft.Id, ProjectId: projectB, Title: "Loose"})

	if nil != err {
		t.Fatalf("expected a draft to move projects, got %v", err)
	}

	if projectB != moved.ProjectId || domain.StatusDraft != moved.Status {
		t.Fatalf("expected the draft re-pointed at B, got %+v", moved)
	}

	// and an unpublished log may move too: unpublish is the sanctioned door
	logs.Unpublish(published.Id)

	if _, err := logs.Update(domain.CaseLog{Id: published.Id, ProjectId: projectB, Title: "Pinned"}); nil != err {
		t.Fatalf("expected the unpublished log to move, got %v", err)
	}
}

func TestCaseLogRevisionCountsPrintingsNotLifecycle(t *testing.T) {
	logs, projects := newCaseLogs()
	projectID := seedProject(projects, "the-light")

	// a new log is printing one, whatever the caller sent
	saved, err := logs.Create(domain.CaseLog{ProjectId: projectID, Title: "First printing", Revision: 42})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if 1 != saved.Revision {
		t.Fatalf("expected create to start the counter at 1, got %d", saved.Revision)
	}

	// every edit is a new printing, server-counted; a stale client number is ignored
	edited, err := logs.Update(domain.CaseLog{Id: saved.Id, ProjectId: projectID, Title: "Second printing", Revision: 42})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if 2 != edited.Revision {
		t.Fatalf("expected update to increment to 2, got %d", edited.Revision)
	}

	// the lifecycle is not an edit: publish and unpublish leave the counter alone
	published, _ := logs.Publish(saved.Id)

	if 2 != published.Revision {
		t.Fatalf("expected publish to leave the counter at 2, got %d", published.Revision)
	}

	unpublished, _ := logs.Unpublish(saved.Id)

	if 2 != unpublished.Revision {
		t.Fatalf("expected unpublish to leave the counter at 2, got %d", unpublished.Revision)
	}

	// a rollback is a new printing: the counter moves on from the live document,
	// never back to the snapshot's number
	revs, _ := logs.Revisions(saved.Id, 100)
	restored, err := logs.Restore(saved.Id, revs[len(revs)-1].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	if 3 != restored.Revision {
		t.Fatalf("expected restore to increment to 3, got %d", restored.Revision)
	}
}

func TestCaseLogListPublishedOnlyFilters(t *testing.T) {
	logs, projects := newCaseLogs()
	shownProject := seedProject(projects, "shown")
	hiddenProject := seedProject(projects, "hidden")

	shown, _ := logs.Create(domain.CaseLog{ProjectId: shownProject, Title: "Shown"})
	logs.Create(domain.CaseLog{ProjectId: hiddenProject, Title: "Hidden"})
	logs.Publish(shown.Id)

	published, _ := logs.List(true, 0)

	if 1 != len(published) || "Shown" != published[0].Title {
		t.Fatalf("expected only the published log, got %+v", published)
	}

	all, _ := logs.List(false, 0)

	if 2 != len(all) {
		t.Fatalf("expected both logs without the filter, got %d", len(all))
	}
}
