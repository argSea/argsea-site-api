package service_test

import (
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newRack wires a project service like newProjects but also hands back the
// activity seam, so the rack tests can assert what lands in the keeper's log.
func newRack() (in_port.ProjectCRUDService, in_port.ActivityService) {
	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	return service.NewProjectCRUDService(out_adapter.NewProjectFakeOutAdapter(), out_adapter.NewNoteFakeOutAdapter(), revisions, activity), activity
}

func TestCreateAssignsNextOrder(t *testing.T) {
	projects, _ := newRack()

	// a body-supplied order is ignored; placement is server-assigned
	first, _ := projects.Create(domain.Project{Title: "First", Order: 99})
	second, _ := projects.Create(domain.Project{Title: "Second"})

	if second.Order != first.Order+1 {
		t.Fatalf("expected each new postcard at the end of the rack, got %d then %d", first.Order, second.Order)
	}
}

func TestCreateNeverStartsFeatured(t *testing.T) {
	projects, _ := newRack()

	saved, _ := projects.Create(domain.Project{Title: "Sneaky", Featured: true})

	if saved.Featured {
		t.Fatalf("nothing reaches the mantel except through the feature endpoint")
	}
}

func TestListSortsByOrderThenCreatedAt(t *testing.T) {
	projects, _ := newRack()

	a, _ := projects.Create(domain.Project{Title: "A"})
	b, _ := projects.Create(domain.Project{Title: "B"})
	c, _ := projects.Create(domain.Project{Title: "C"})

	// shuffle: C to the front, A and B tied behind it; the tie breaks on
	// createdAt asc, so A (created first) comes before B
	projects.Reorder(c.Id, 0)
	projects.Reorder(a.Id, 5)
	projects.Reorder(b.Id, 5)

	all, err := projects.List(false, 0)

	if nil != err {
		t.Fatalf("list failed: %v", err)
	}

	got := []string{all[0].Title, all[1].Title, all[2].Title}

	if "C" != got[0] || "A" != got[1] || "B" != got[2] {
		t.Fatalf("expected order C, A, B; got %v", got)
	}
}

func TestReorderAndFeatureSkipSnapshotsButLog(t *testing.T) {
	projects, activity := newRack()

	saved, _ := projects.Create(domain.Project{Title: "Rack me"})

	if _, err := projects.Reorder(saved.Id, 7); nil != err {
		t.Fatalf("reorder failed: %v", err)
	}

	if _, err := projects.Feature(saved.Id); nil != err {
		t.Fatalf("feature failed: %v", err)
	}

	if _, err := projects.Unfeature(saved.Id); nil != err {
		t.Fatalf("unfeature failed: %v", err)
	}

	// lifecycle-style: the create snapshot stays the only revision
	revs, _ := projects.Revisions(saved.Id, 100)

	if 1 != len(revs) {
		t.Fatalf("reorder/feature must not snapshot, expected 1 revision, got %d", len(revs))
	}

	// but every move lands in the keeper's log: create + reorder + feature + unfeature
	entries, _ := activity.Recent(10)

	if 4 != len(entries) {
		t.Fatalf("expected 4 activity entries, got %d", len(entries))
	}

	for _, verb := range []string{"reordered", "featured", "unfeatured"} {
		found := false

		for _, entry := range entries {
			if strings.Contains(entry.Message, verb) {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("expected a %q entry in the keeper's log, got %+v", verb, entries)
		}
	}
}

func TestFeatureAndUnfeatureToggleTheFlag(t *testing.T) {
	projects, _ := newRack()

	saved, _ := projects.Create(domain.Project{Title: "Mantel"})

	featured, err := projects.Feature(saved.Id)

	if nil != err || !featured.Featured {
		t.Fatalf("expected the postcard on the mantel, got %+v (%v)", featured, err)
	}

	unfeatured, err := projects.Unfeature(saved.Id)

	if nil != err || unfeatured.Featured {
		t.Fatalf("expected the postcard off the mantel, got %+v (%v)", unfeatured, err)
	}
}

func TestUpdatePreservesOrderAndFeatured(t *testing.T) {
	projects, _ := newRack()

	saved, _ := projects.Create(domain.Project{Title: "Hold position"})
	projects.Reorder(saved.Id, 5)
	projects.Feature(saved.Id)

	// a full-replace PUT carrying stale placement must not move the postcard;
	// order and featured only change through their endpoints
	edited, err := projects.Update(domain.Project{Id: saved.Id, Title: "Hold position", Order: 99, Featured: false})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if 5 != edited.Order || !edited.Featured {
		t.Fatalf("expected order 5 and featured preserved, got %d / %v", edited.Order, edited.Featured)
	}
}

func TestReorderAndFeatureRejectUnknownProject(t *testing.T) {
	projects, _ := newRack()

	if _, err := projects.Reorder("nope", 1); nil == err {
		t.Fatalf("expected reorder to reject an unknown project")
	}

	if _, err := projects.Feature("nope"); nil == err {
		t.Fatalf("expected feature to reject an unknown project")
	}
}
