package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newProjectsWithNotes wires a project service like newProjects but hands
// back the note repo directly too, so a test can seed notes for the tie
// check without going through a second, unrelated NoteCRUDService.
func newProjectsWithNotes() (in_port.ProjectCRUDService, out_port.NoteRepo) {
	revisions := service.NewRevisionService(out_adapter.NewRevisionFakeOutAdapter())
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	notes := out_adapter.NewNoteFakeOutAdapter()

	return service.NewProjectCRUDService(out_adapter.NewProjectFakeOutAdapter(), notes, revisions, activity), notes
}

// sixFacts returns a full stat strip that passes the cap, for tests to mutate.
func sixFacts() []domain.ProjectFact {
	facts := make([]domain.ProjectFact, 6)

	for i := range facts {
		facts[i] = domain.ProjectFact{Heading: "Heading", Fact: "Fact"}
	}

	return facts
}

func TestFactsCappedAtSix(t *testing.T) {
	projects := newProjects()

	if _, err := projects.Create(domain.Project{Title: "Full strip", Facts: sixFacts()}); nil != err {
		t.Fatalf("expected 6 facts accepted, got %v", err)
	}

	seven := append(sixFacts(), domain.ProjectFact{Heading: "One too many", Fact: "x"})

	if _, err := projects.Create(domain.Project{Title: "Overfull", Facts: seven}); nil == err {
		t.Fatalf("expected 7 facts rejected")
	}
}

func TestFactsRejectEmptyHeadingOrFact(t *testing.T) {
	projects := newProjects()

	cases := map[string][]domain.ProjectFact{
		"empty heading": {{Heading: "", Fact: "Fact"}},
		"empty fact":    {{Heading: "Heading", Fact: ""}},
		"both blank":    {{Heading: "   ", Fact: "   "}},
	}

	for name, facts := range cases {
		if _, err := projects.Create(domain.Project{Title: "Bad " + name, Facts: facts}); nil == err {
			t.Fatalf("expected %s rejected", name)
		}
	}

	// none of the rejected creates may have written anything
	all, _ := projects.List(false, 0)

	if 0 != len(all) {
		t.Fatalf("rejected creates must persist nothing, found %d projects", len(all))
	}
}

func TestFactsTrimmedOnStore(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{
		Title: "Trim",
		Facts: []domain.ProjectFact{{Heading: "  Founded  ", Fact: "  2019  "}},
	})

	if nil != err {
		t.Fatalf("create with padded facts failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if 1 != len(stored.Facts) || "Founded" != stored.Facts[0].Heading || "2019" != stored.Facts[0].Fact {
		t.Fatalf("expected the fact trimmed on store, got %+v", stored.Facts)
	}
}

func TestNoteIdsRejectsUnknownId(t *testing.T) {
	projects, notes := newProjectsWithNotes()

	noteID, err := notes.Add(domain.Note{Title: "Journal entry"})

	if nil != err {
		t.Fatalf("seed note failed: %v", err)
	}

	if _, err := projects.Create(domain.Project{Title: "Tied", NoteIds: []string{noteID}}); nil != err {
		t.Fatalf("expected a known note id accepted, got %v", err)
	}

	if _, err := projects.Create(domain.Project{Title: "Untied", NoteIds: []string{"nope"}}); nil == err {
		t.Fatalf("expected an unknown note id rejected")
	}

	// the update path rejects too
	saved, _ := projects.Create(domain.Project{Title: "Good", NoteIds: []string{noteID}})

	if _, err := projects.Update(domain.Project{Id: saved.Id, Title: "Good", NoteIds: []string{"nope"}}); nil == err {
		t.Fatalf("expected update to reject an unknown note id")
	}
}

func TestSlugUniqueAcrossProjects(t *testing.T) {
	projects := newProjects()

	// an empty slug is fine: the slug field is optional now that the case study
	// (and its /projects/<slug> route) lives in its own caselog
	if _, err := projects.Create(domain.Project{Title: "Plain"}); nil != err {
		t.Fatalf("expected a project without a slug accepted, got %v", err)
	}

	first, err := projects.Create(domain.Project{Title: "First", Slug: "the-light"})

	if nil != err {
		t.Fatalf("expected a project with a slug accepted, got %v", err)
	}

	// case-insensitive collision
	if _, err := projects.Create(domain.Project{Title: "Second", Slug: "THE-LIGHT"}); nil == err {
		t.Fatalf("expected a case-insensitive slug collision rejected")
	}

	// the update path rejects a collision too
	second, _ := projects.Create(domain.Project{Title: "Second", Slug: "another-light"})

	if _, err := projects.Update(domain.Project{Id: second.Id, Title: "Second", Slug: "The-Light"}); nil == err {
		t.Fatalf("expected update to reject a case-insensitive slug collision")
	}

	// a project keeping its own slug on update must not collide with itself
	if _, err := projects.Update(domain.Project{Id: first.Id, Title: "First, retitled", Slug: "the-light"}); nil != err {
		t.Fatalf("expected a project to keep its own slug on update, got %v", err)
	}
}

func TestFlagshipFalseSurvivesReplaceWrite(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{Title: "Flagship", Flagship: true})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if !saved.Flagship {
		t.Fatalf("expected flagship true to persist on create")
	}

	// unlike featured, flagship has no dedicated endpoint; it rides the
	// ordinary replace write, so a PUT clearing it must stick
	cleared, err := projects.Update(domain.Project{Id: saved.Id, Title: "Flagship", Flagship: false})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if cleared.Flagship {
		t.Fatalf("expected flagship false to survive the replace write, got true")
	}

	if projects.Read(saved.Id).Flagship {
		t.Fatalf("expected the stored document to also read flagship false")
	}
}
