package service_test

import (
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// validLight returns a light that passes every gate, for tests to mutate.
func validLight() *domain.Light {
	return &domain.Light{Kind: "flash", Color: "white", Period: 8}
}

func TestValidLightAcceptedOnCreateAndUpdate(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{Title: "Lit", Light: validLight()})

	if nil != err {
		t.Fatalf("create with a valid light failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "flash" != stored.Light.Kind || 8 != stored.Light.Period {
		t.Fatalf("expected the light persisted on create, got %+v", stored.Light)
	}

	// swap kind and color, and put the light out, on update
	edited, err := projects.Update(domain.Project{
		Id:    saved.Id,
		Title: "Lit",
		Light: &domain.Light{Kind: "occult", Color: "green", Period: 6, Extinguished: "2020"},
	})

	if nil != err {
		t.Fatalf("update with a valid light failed: %v", err)
	}

	if nil == edited.Light || "occult" != edited.Light.Kind || "2020" != edited.Light.Extinguished {
		t.Fatalf("expected the light replaced on update, got %+v", edited.Light)
	}
}

func TestLightRejectsOutOfSetEnumValues(t *testing.T) {
	projects := newProjects()

	cases := map[string]func(l *domain.Light){
		"kind":  func(l *domain.Light) { l.Kind = "strobe" },
		"color": func(l *domain.Light) { l.Color = "ultraviolet" },
	}

	for field, corrupt := range cases {
		light := validLight()
		corrupt(light)

		if _, err := projects.Create(domain.Project{Title: "Bad " + field, Light: light}); nil == err {
			t.Fatalf("expected create to reject an out-of-set %s", field)
		}
	}

	// none of the rejected creates may have written anything
	all, _ := projects.List(false, 0)

	if 0 != len(all) {
		t.Fatalf("rejected creates must persist nothing, found %d projects", len(all))
	}

	// the update path rejects too, and the stored light survives untouched
	saved, _ := projects.Create(domain.Project{Title: "Good", Light: validLight()})

	bad := validLight()
	bad.Color = "javascript:alert(1)"

	if _, err := projects.Update(domain.Project{Id: saved.Id, Title: "Good", Light: bad}); nil == err {
		t.Fatalf("expected update to reject an out-of-set color")
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "white" != stored.Light.Color {
		t.Fatalf("rejected update must leave the stored light intact, got %+v", stored.Light)
	}
}

func TestLightPeriodCoupledToBlinkingKinds(t *testing.T) {
	projects := newProjects()

	// a fixed light carries no period; any other kind needs one in bounds
	fixed := &domain.Light{Kind: "fixed", Color: "red", Period: 0}

	if _, err := projects.Create(domain.Project{Title: "Fixed", Light: fixed}); nil != err {
		t.Fatalf("expected a fixed light with period 0 accepted, got %v", err)
	}

	stray := &domain.Light{Kind: "fixed", Color: "red", Period: 5}

	if _, err := projects.Create(domain.Project{Title: "Fixed stray", Light: stray}); nil == err {
		t.Fatalf("expected a period on a fixed light rejected")
	}

	for _, ok := range []int{2, 30} {
		light := validLight()
		light.Period = ok

		if _, err := projects.Create(domain.Project{Title: "Bounds", Light: light}); nil != err {
			t.Fatalf("expected period %d accepted, got %v", ok, err)
		}
	}

	for _, bad := range []int{0, 1, 31} {
		light := validLight()
		light.Period = bad

		if _, err := projects.Create(domain.Project{Title: "Bounds", Light: light}); nil == err {
			t.Fatalf("expected period %d rejected", bad)
		}
	}
}

func TestConventionTimedKindsCarryNoPeriod(t *testing.T) {
	projects := newProjects()

	// quick and veryquick keep convention time like fixed does; they carry
	// no period, and a stray one is rejected rather than stored unrendered
	for _, kind := range []string{"quick", "veryquick"} {
		light := &domain.Light{Kind: kind, Color: "white", Period: 0}

		if _, err := projects.Create(domain.Project{Title: "Convention", Light: light}); nil != err {
			t.Fatalf("expected a %s light with period 0 accepted, got %v", kind, err)
		}

		stray := &domain.Light{Kind: kind, Color: "white", Period: 5}

		if _, err := projects.Create(domain.Project{Title: "Convention stray", Light: stray}); nil == err {
			t.Fatalf("expected a period on a %s light rejected", kind)
		}
	}
}

func TestMorseLetterCoupledToKind(t *testing.T) {
	projects := newProjects()

	// the letter normalizes like extinguished trims: the store holds exactly
	// what was validated
	morse := &domain.Light{Kind: "morse", Color: "white", Period: 8, Letter: "  a  "}

	saved, err := projects.Create(domain.Project{Title: "Mo", Light: morse})

	if nil != err {
		t.Fatalf("create with a valid morse light failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "A" != stored.Light.Letter {
		t.Fatalf("expected the letter normalized to %q, got %+v", "A", stored.Light)
	}

	cases := map[string]*domain.Light{
		"morse without a letter":   {Kind: "morse", Color: "white", Period: 8},
		"morse with two letters":   {Kind: "morse", Color: "white", Period: 8, Letter: "AB"},
		"morse with a digit":       {Kind: "morse", Color: "white", Period: 8, Letter: "7"},
		"a letter off morse":       {Kind: "flash", Color: "white", Period: 8, Letter: "A"},
		"a letter on a fixed kind": {Kind: "fixed", Color: "white", Letter: "A"},
	}

	for name, light := range cases {
		if _, err := projects.Create(domain.Project{Title: "Bad " + name, Light: light}); nil == err {
			t.Fatalf("expected %s rejected", name)
		}
	}
}

func TestMorsePeriodNeedsRoom(t *testing.T) {
	projects := newProjects()

	// morse needs room for its pattern: the longest letters run 5.2s of
	// dots and dashes, so cycles under six seconds are rejected
	for _, ok := range []int{6, 30} {
		light := &domain.Light{Kind: "morse", Color: "white", Period: ok, Letter: "K"}

		if _, err := projects.Create(domain.Project{Title: "Mo bounds", Light: light}); nil != err {
			t.Fatalf("expected morse period %d accepted, got %v", ok, err)
		}
	}

	for _, bad := range []int{0, 5, 31} {
		light := &domain.Light{Kind: "morse", Color: "white", Period: bad, Letter: "K"}

		if _, err := projects.Create(domain.Project{Title: "Mo bounds", Light: light}); nil == err {
			t.Fatalf("expected morse period %d rejected", bad)
		}
	}
}

func TestMorseLightSurvivesLifecycleOps(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{
		Title: "Signal",
		Light: &domain.Light{Kind: "morse", Color: "red", Period: 10, Letter: "K"},
	})

	// lifecycle ops reconstruct from the stored document; the letter must
	// ride along untouched like the rest of the light
	projects.Publish(saved.Id)
	projects.Reorder(saved.Id, 3)
	projects.Feature(saved.Id)
	projects.Unpublish(saved.Id)

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "morse" != stored.Light.Kind || "K" != stored.Light.Letter {
		t.Fatalf("expected the morse light to survive lifecycle ops, got %+v", stored.Light)
	}
}

func TestLightExtinguishedTrimmedAndCapped(t *testing.T) {
	projects := newProjects()

	// the store holds exactly what was validated; no surrounding padding
	light := validLight()
	light.Extinguished = "  2020  "

	saved, err := projects.Create(domain.Project{Title: "Dark", Light: light})

	if nil != err {
		t.Fatalf("create with a padded extinguished failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "2020" != stored.Light.Extinguished {
		t.Fatalf("expected extinguished trimmed to %q, got %+v", "2020", stored.Light)
	}

	over := validLight()
	over.Extinguished = strings.Repeat("a", 41)

	if _, err := projects.Create(domain.Project{Title: "Dark", Light: over}); nil == err {
		t.Fatalf("expected a 41-char extinguished rejected")
	}
}

func TestAbsentLightIsValid(t *testing.T) {
	projects := newProjects()

	// no light at all is the valid default state; the site burns it as the
	// default fixed white
	saved, err := projects.Create(domain.Project{Title: "Unlit"})

	if nil != err {
		t.Fatalf("create without a light failed: %v", err)
	}

	if nil != saved.Light {
		t.Fatalf("expected no light on the saved project, got %+v", saved.Light)
	}
}

func TestImagesTrimmedAndEmptyRejected(t *testing.T) {
	projects := newProjects()

	saved, err := projects.Create(domain.Project{Title: "Gallery", Images: []string{"  one.jpg  ", "two.jpg"}})

	if nil != err {
		t.Fatalf("create with a valid gallery failed: %v", err)
	}

	stored := projects.Read(saved.Id)

	if 2 != len(stored.Images) || "one.jpg" != stored.Images[0] {
		t.Fatalf("expected gallery names trimmed on store, got %v", stored.Images)
	}

	for _, bad := range [][]string{{""}, {"ok.jpg", "   "}} {
		if _, err := projects.Create(domain.Project{Title: "Gallery", Images: bad}); nil == err {
			t.Fatalf("expected a gallery with an empty name (%v) rejected", bad)
		}
	}
}

func TestImagesCappedAtTwelve(t *testing.T) {
	projects := newProjects()

	twelve := make([]string, 12)
	for i := range twelve {
		twelve[i] = "print.jpg"
	}

	if _, err := projects.Create(domain.Project{Title: "Full album", Images: twelve}); nil != err {
		t.Fatalf("expected 12 prints accepted, got %v", err)
	}

	thirteen := append(twelve, "one-too-many.jpg")

	if _, err := projects.Create(domain.Project{Title: "Overfull", Images: thirteen}); nil == err {
		t.Fatalf("expected 13 prints rejected")
	}
}

func TestFirstLitTrimmedOnStore(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{Title: "Est", FirstLit: "  2024  "})

	stored := projects.Read(saved.Id)

	if "2024" != stored.FirstLit {
		t.Fatalf("expected firstLit trimmed on store, got %q", stored.FirstLit)
	}

	edited, err := projects.Update(domain.Project{Id: saved.Id, Title: "Est", FirstLit: " 2025 "})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if "2025" != edited.FirstLit {
		t.Fatalf("expected firstLit trimmed on update, got %q", edited.FirstLit)
	}
}

func TestUpdateWithoutLightFieldsClearsThem(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{
		Title:    "Clear me",
		Light:    validLight(),
		Images:   []string{"one.jpg"},
		FirstLit: "2024",
	})

	// PUT is full-replace: fields omitted from the update are fields removed
	cleared, err := projects.Update(domain.Project{Id: saved.Id, Title: "Clear me"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if nil != cleared.Light || 0 != len(cleared.Images) || "" != cleared.FirstLit {
		t.Fatalf("expected light/images/firstLit cleared, got %+v / %v / %q", cleared.Light, cleared.Images, cleared.FirstLit)
	}
}

func TestLifecycleOpsPreserveLightFields(t *testing.T) {
	projects := newProjects()

	saved, _ := projects.Create(domain.Project{
		Title:    "Steady",
		Light:    validLight(),
		Images:   []string{"one.jpg"},
		FirstLit: "2024",
	})

	// lifecycle ops reconstruct from the stored document and touch only
	// lifecycle fields; the light fields must ride along untouched
	projects.Publish(saved.Id)
	projects.Reorder(saved.Id, 5)
	projects.Feature(saved.Id)
	projects.Unpublish(saved.Id)

	stored := projects.Read(saved.Id)

	if nil == stored.Light || "flash" != stored.Light.Kind {
		t.Fatalf("expected the light to survive lifecycle ops, got %+v", stored.Light)
	}

	if 1 != len(stored.Images) || "2024" != stored.FirstLit {
		t.Fatalf("expected images/firstLit to survive lifecycle ops, got %v / %q", stored.Images, stored.FirstLit)
	}
}

func TestLightRoundTripsThroughSnapshotRestore(t *testing.T) {
	projects := newProjects()

	// rev 1: flashing white; rev 2: extinguished green
	saved, _ := projects.Create(domain.Project{Title: "Round trip", Light: validLight(), FirstLit: "2019"})
	projects.Update(domain.Project{
		Id:    saved.Id,
		Title: "Round trip",
		Light: &domain.Light{Kind: "iso", Color: "green", Period: 4, Extinguished: "2022"},
	})

	revs, _ := projects.Revisions(saved.Id, 100)
	restored, err := projects.Restore(saved.Id, revs[len(revs)-1].Id)

	if nil != err {
		t.Fatalf("restore failed: %v", err)
	}

	// the original light comes back exactly as snapshotted, still burning
	if nil == restored.Light || "flash" != restored.Light.Kind || "" != restored.Light.Extinguished {
		t.Fatalf("expected restore to bring the original light back, got %+v", restored.Light)
	}

	if "2019" != restored.FirstLit {
		t.Fatalf("expected restore to bring firstLit back, got %q", restored.FirstLit)
	}
}
