package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

func TestArrangementPinsListedProjects(t *testing.T) {
	projects, _ := newRack()

	first, _ := projects.Create(domain.Project{Title: "First"})
	second, _ := projects.Create(domain.Project{Title: "Second"})

	saved, err := projects.Arrangement([]domain.WallPlacement{
		{Id: first.Id, X: 10, Y: 20, Rotation: -5},
		{Id: second.Id, X: 30, Y: 40, Rotation: 5},
	})

	if nil != err {
		t.Fatalf("arrangement failed: %v", err)
	}

	if 2 != len(saved) {
		t.Fatalf("expected 2 pinned postcards, got %d", len(saved))
	}

	pinned := projects.Read(first.Id)

	if nil == pinned.WallPos {
		t.Fatalf("expected %q to be pinned", first.Title)
	}

	if 10 != pinned.WallPos.X || 20 != pinned.WallPos.Y || -5 != pinned.WallPos.Rotation {
		t.Fatalf("expected wall pos {10 20 -5}, got %+v", pinned.WallPos)
	}
}

func TestArrangementLeavesUnlistedProjectsUntouched(t *testing.T) {
	projects, _ := newRack()

	pinned, _ := projects.Create(domain.Project{Title: "Pinned"})
	untouched, _ := projects.Create(domain.Project{Title: "Untouched"})

	projects.Arrangement([]domain.WallPlacement{{Id: pinned.Id, X: 1, Y: 2, Rotation: 3}})

	still := projects.Read(untouched.Id)

	if nil != still.WallPos {
		t.Fatalf("expected %q to stay unpinned, got %+v", untouched.Title, still.WallPos)
	}
}

func TestUpdatePreservesWallPos(t *testing.T) {
	projects, _ := newRack()

	saved, _ := projects.Create(domain.Project{Title: "On the wall"})
	projects.Arrangement([]domain.WallPlacement{{Id: saved.Id, X: 15, Y: 25, Rotation: 10}})

	// a content-edit PUT must not move a card's wall position
	edited, err := projects.Update(domain.Project{Id: saved.Id, Title: "On the wall, retitled"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if nil == edited.WallPos {
		t.Fatalf("expected wall position preserved through update")
	}

	if 15 != edited.WallPos.X || 25 != edited.WallPos.Y || 10 != edited.WallPos.Rotation {
		t.Fatalf("expected wall pos {15 25 10} preserved, got %+v", edited.WallPos)
	}
}

func TestArrangementRejectsUnknownProject(t *testing.T) {
	projects, _ := newRack()

	if _, err := projects.Arrangement([]domain.WallPlacement{{Id: "nope", X: 1, Y: 1, Rotation: 1}}); nil == err {
		t.Fatalf("expected arrangement to reject an unknown project")
	}
}
