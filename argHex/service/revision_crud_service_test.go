package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newRevisions wires the revision service onto its in-memory fake.
func newRevisions() in_port.RevisionService {
	return service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
}

func countCurrent(revs domain.Revisions) int {
	current := 0

	for _, rev := range revs {
		if rev.IsCurrent {
			current++
		}
	}

	return current
}

func TestSnapshotMovesCurrentPointer(t *testing.T) {
	revisions := newRevisions()

	first, _ := revisions.Snapshot(domain.EntityProject, "p1", `{"title":"one"}`, "one")
	revisions.Snapshot(domain.EntityProject, "p1", `{"title":"two"}`, "two")
	latest, _ := revisions.Snapshot(domain.EntityProject, "p1", `{"title":"three"}`, "three")

	all, _ := revisions.List(domain.EntityProject, "p1", 100)

	if 3 != len(all) {
		t.Fatalf("expected 3 revisions, got %d", len(all))
	}

	// exactly one revision may be current — the pointer moved to the newest
	if 1 != countCurrent(all) {
		t.Fatalf("expected exactly 1 current revision, got %d", countCurrent(all))
	}

	current := revisions.Current(domain.EntityProject, "p1")

	if current.Id != latest {
		t.Fatalf("expected current to be newest revision %q, got %q", latest, current.Id)
	}

	if !current.IsCurrent {
		t.Fatalf("current revision should carry the current flag")
	}

	// the first revision must no longer be current
	if revisions.Get(first).IsCurrent {
		t.Fatalf("first revision should have lost the current flag")
	}
}

func TestListIsNewestFirstAndHonoursLimit(t *testing.T) {
	revisions := newRevisions()

	revisions.Snapshot(domain.EntityNote, "n1", `{"v":1}`, "one")
	revisions.Snapshot(domain.EntityNote, "n1", `{"v":2}`, "two")
	revisions.Snapshot(domain.EntityNote, "n1", `{"v":3}`, "three")

	limited, _ := revisions.List(domain.EntityNote, "n1", 2)

	if 2 != len(limited) {
		t.Fatalf("expected limit of 2 to be honoured, got %d", len(limited))
	}

	if "three" != limited[0].Summary {
		t.Fatalf("expected newest-first ordering, got %q first", limited[0].Summary)
	}
}

func TestListIsScopedToEntity(t *testing.T) {
	revisions := newRevisions()

	revisions.Snapshot(domain.EntityProject, "p1", `{}`, "p1")
	revisions.Snapshot(domain.EntityProject, "p2", `{}`, "p2")

	forP1, _ := revisions.List(domain.EntityProject, "p1", 100)

	if 1 != len(forP1) {
		t.Fatalf("expected revisions scoped to entity, got %d for p1", len(forP1))
	}
}
