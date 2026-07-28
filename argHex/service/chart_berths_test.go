package service_test

import (
	"encoding/json"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// An older document carries none of the berth keys. Every chartable must read
// it as uncharted and undressed rather than failing to parse, which is the
// whole reason the fields are optional on the wire.
func TestBerthFieldsAbsentParseAsUnchartedAndUndressed(t *testing.T) {
	var project domain.Project

	if err := json.Unmarshal([]byte(`{"title":"An old light"}`), &project); nil != err {
		t.Fatalf("an older project document must still parse, got %v", err)
	}

	if nil != project.Coord || 0 != project.Plate || "" != project.Cap {
		t.Fatalf("an older project must read as uncharted and undressed, got %+v / %d / %q", project.Coord, project.Plate, project.Cap)
	}

	var note domain.Note

	if err := json.Unmarshal([]byte(`{"title":"An old thought"}`), &note); nil != err {
		t.Fatalf("an older note document must still parse, got %v", err)
	}

	if nil != note.Coord || 0 != note.Plate || "" != note.Cap {
		t.Fatalf("an older note must read as uncharted and undressed, got %+v / %d / %q", note.Coord, note.Plate, note.Cap)
	}

	var hobby domain.Hobby

	if err := json.Unmarshal([]byte(`{"name":"An old ship","state":"moored"}`), &hobby); nil != err {
		t.Fatalf("an older hobby document must still parse, got %v", err)
	}

	if nil != hobby.Coord || 0 != hobby.Plate || "" != hobby.Cap {
		t.Fatalf("an older hobby must read as undressed, got %+v / %d / %q", hobby.Coord, hobby.Plate, hobby.Cap)
	}
}

// plateEdges walks the plate's one bound: a negative index snaps to zero, and
// anything at or above zero rides through, because the site owns how many
// plates exist and resolves an unknown index to its fallback.
var plateEdges = []struct {
	name  string
	plate int
	want  int
}{
	{"negative snaps to zero", -3, 0},
	{"zero rides through", 0, 0},
	{"a known plate rides through", 4, 4},
	{"an index past the site's plates still rides through", 9001, 9001},
}

func TestPlateClampsAtZeroBelowOnEveryChartable(t *testing.T) {
	for _, edge := range plateEdges {
		projects := newProjects()
		saved, err := projects.Create(domain.Project{Title: "A light", Plate: edge.plate})

		if nil != err {
			t.Fatalf("%s: project create failed: %v", edge.name, err)
		}

		if edge.want != projects.Read(saved.Id).Plate {
			t.Fatalf("%s: project plate got %d, want %d", edge.name, projects.Read(saved.Id).Plate, edge.want)
		}

		notes := newNotes()
		note, err := notes.Create(domain.Note{Title: "A thought", Plate: edge.plate})

		if nil != err {
			t.Fatalf("%s: note create failed: %v", edge.name, err)
		}

		if edge.want != notes.Read(note.Id).Plate {
			t.Fatalf("%s: note plate got %d, want %d", edge.name, notes.Read(note.Id).Plate, edge.want)
		}

		hobbies := newHobbies()
		hobby, err := hobbies.Create(domain.Hobby{Name: "A ship", State: domain.StateMoored, Plate: edge.plate})

		if nil != err {
			t.Fatalf("%s: hobby create failed: %v", edge.name, err)
		}

		if edge.want != hobbies.Read(hobby.Id).Plate {
			t.Fatalf("%s: hobby plate got %d, want %d", edge.name, hobbies.Read(hobby.Id).Plate, edge.want)
		}
	}
}

// The update path clamps too: a negative plate must never reach the store by
// the back door.
func TestPlateClampsOnUpdate(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "A light", Plate: 4})

	if _, err := projects.Update(domain.Project{Id: light.Id, Title: "A light", Plate: -2}); nil != err {
		t.Fatalf("project update failed: %v", err)
	}

	if 0 != projects.Read(light.Id).Plate {
		t.Fatalf("project update must clamp a negative plate to zero, got %d", projects.Read(light.Id).Plate)
	}

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "A thought", Plate: 4})

	if _, err := notes.Update(domain.Note{Id: thought.Id, Title: "A thought", Plate: -2}); nil != err {
		t.Fatalf("note update failed: %v", err)
	}

	if 0 != notes.Read(thought.Id).Plate {
		t.Fatalf("note update must clamp a negative plate to zero, got %d", notes.Read(thought.Id).Plate)
	}

	hobbies := newHobbies()
	ship, _ := hobbies.Create(domain.Hobby{Name: "A ship", State: domain.StateMoored, Plate: 4})

	if _, err := hobbies.Update(domain.Hobby{Id: ship.Id, Name: "A ship", State: domain.StateMoored, Plate: -2}); nil != err {
		t.Fatalf("hobby update failed: %v", err)
	}

	if 0 != hobbies.Read(ship.Id).Plate {
		t.Fatalf("hobby update must clamp a negative plate to zero, got %d", hobbies.Read(ship.Id).Plate)
	}
}

func TestBerthFieldsRoundTripThroughTheStore(t *testing.T) {
	projects := newProjects()
	light, err := projects.Create(domain.Project{
		Title: "A light",
		Coord: &domain.Coord{Lat: 58.22, Lon: -7.5},
		Plate: 2,
		Cap:   "the light at the head of the sound",
	})

	if nil != err {
		t.Fatalf("project create failed: %v", err)
	}

	storedLight := projects.Read(light.Id)

	if nil == storedLight.Coord || 58.22 != storedLight.Coord.Lat || -7.5 != storedLight.Coord.Lon {
		t.Fatalf("project coord did not round-trip, got %+v", storedLight.Coord)
	}

	if 2 != storedLight.Plate || "the light at the head of the sound" != storedLight.Cap {
		t.Fatalf("project dressing did not round-trip, got %d / %q", storedLight.Plate, storedLight.Cap)
	}

	notes := newNotes()
	thought, err := notes.Create(domain.Note{
		Title: "A thought",
		Coord: &domain.Coord{Lat: 58.05, Lon: -7.1},
		Plate: 3,
		Cap:   "written the morning the fog lifted",
	})

	if nil != err {
		t.Fatalf("note create failed: %v", err)
	}

	storedNote := notes.Read(thought.Id)

	if nil == storedNote.Coord || 58.05 != storedNote.Coord.Lat || -7.1 != storedNote.Coord.Lon {
		t.Fatalf("note coord did not round-trip, got %+v", storedNote.Coord)
	}

	if 3 != storedNote.Plate || "written the morning the fog lifted" != storedNote.Cap {
		t.Fatalf("note dressing did not round-trip, got %d / %q", storedNote.Plate, storedNote.Cap)
	}

	hobbies := newHobbies()
	ship, err := hobbies.Create(domain.Hobby{
		Name:  "A ship",
		State: domain.StateMoored,
		Coord: &domain.Coord{Lat: 58.1, Lon: -7.2},
		Plate: 1,
		Cap:   "still at her berth",
	})

	if nil != err {
		t.Fatalf("hobby create failed: %v", err)
	}

	storedShip := hobbies.Read(ship.Id)

	if 1 != storedShip.Plate || "still at her berth" != storedShip.Cap {
		t.Fatalf("hobby dressing did not round-trip, got %d / %q", storedShip.Plate, storedShip.Cap)
	}
}

// An off-window coord snaps into the chart window on the new chartables the
// same way it always has on a hobby, and a null one stays uncharted.
func TestCoordClampsAndStaysNullableOnProjectAndNote(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "A light", Coord: &domain.Coord{Lat: 99.0, Lon: 99.0}})
	storedLight := projects.Read(light.Id)

	if 58.56 != storedLight.Coord.Lat || -6.59 != storedLight.Coord.Lon {
		t.Fatalf("an off-window project coord must store clamped, got %+v", storedLight.Coord)
	}

	uncharted, _ := projects.Create(domain.Project{Title: "An unplaced light"})

	if nil != projects.Read(uncharted.Id).Coord {
		t.Fatalf("an unplaced light must keep coord nil, got %+v", projects.Read(uncharted.Id).Coord)
	}

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "A thought", Coord: &domain.Coord{Lat: -99.0, Lon: -99.0}})
	storedNote := notes.Read(thought.Id)

	if 57.82 != storedNote.Coord.Lat || -7.94 != storedNote.Coord.Lon {
		t.Fatalf("an off-window note coord must store clamped, got %+v", storedNote.Coord)
	}

	unplaced, _ := notes.Create(domain.Note{Title: "An unplaced thought"})

	if nil != notes.Read(unplaced.Id).Coord {
		t.Fatalf("an unplaced thought must keep coord nil, got %+v", notes.Read(unplaced.Id).Coord)
	}
}

// The berth fields carry no omitempty because clearing one is a real edit: an
// update that drops the caption and the plate must not silently keep the old
// pair alive through the replace write.
func TestClearingTheDressingSurvivesAReplaceWrite(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "A light", Plate: 2, Cap: "a caption"})
	projects.Update(domain.Project{Id: light.Id, Title: "A light"})

	if stored := projects.Read(light.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a project's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "A thought", Plate: 3, Cap: "a caption"})
	notes.Update(domain.Note{Id: thought.Id, Title: "A thought"})

	if stored := notes.Read(thought.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a note's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}

	hobbies := newHobbies()
	ship, _ := hobbies.Create(domain.Hobby{Name: "A ship", State: domain.StateMoored, Plate: 1, Cap: "a caption"})
	hobbies.Update(domain.Hobby{Id: ship.Id, Name: "A ship", State: domain.StateMoored})

	if stored := hobbies.Read(ship.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a ship's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}
}
