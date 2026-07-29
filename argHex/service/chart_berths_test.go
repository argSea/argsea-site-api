package service_test

import (
	"encoding/json"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
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

// The earth-range check guards the update path too, not just create: an
// off-earth bearing must never reach the store by the back door of an edit, and
// a rejected edit must leave the stored bearing exactly as it was.
func TestCoordValidatesEarthRangeOnUpdate(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "A light", Coord: &domain.Coord{Lat: 58.1, Lon: -7.2}})

	// a berth above the hobby chart's old ceiling is a legitimate edit
	if _, err := projects.Update(domain.Project{Id: light.Id, Title: "A light", Coord: &domain.Coord{Lat: 58.61, Lon: -7.2}}); nil != err {
		t.Fatalf("project update with an on-earth coord failed: %v", err)
	}

	if stored := projects.Read(light.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat {
		t.Fatalf("project update must store an on-earth coord unmoved, got %+v", stored.Coord)
	}

	if _, err := projects.Update(domain.Project{Id: light.Id, Title: "A light", Coord: &domain.Coord{Lat: 99.0, Lon: 99.0}}); nil == err {
		t.Fatalf("expected project update to reject an off-earth coord")
	}

	if stored := projects.Read(light.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat {
		t.Fatalf("a rejected project update must leave the stored coord intact, got %+v", stored.Coord)
	}

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "A thought", Coord: &domain.Coord{Lat: 58.1, Lon: -7.2}})

	if _, err := notes.Update(domain.Note{Id: thought.Id, Title: "A thought", Coord: &domain.Coord{Lat: 58.61, Lon: -8.2}}); nil != err {
		t.Fatalf("note update with an on-earth coord failed: %v", err)
	}

	if stored := notes.Read(thought.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat || -8.2 != stored.Coord.Lon {
		t.Fatalf("note update must store an on-earth coord unmoved, got %+v", stored.Coord)
	}

	if _, err := notes.Update(domain.Note{Id: thought.Id, Title: "A thought", Coord: &domain.Coord{Lat: -99.0, Lon: -199.0}}); nil == err {
		t.Fatalf("expected note update to reject an off-earth coord")
	}

	if stored := notes.Read(thought.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat {
		t.Fatalf("a rejected note update must leave the stored coord intact, got %+v", stored.Coord)
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

// earthEdges walks the only band any chartable's bearing is held to, a
// project's, a note's and a hobby's alike. The two in-band cases sit
// deliberately outside the hobby chart's old window:
// clamping berths to that window is the bug this replaced, so a berth above its
// ceiling has to survive untouched.
var earthEdges = []struct {
	name     string
	lat, lon float64
	valid    bool
}{
	{"a berth above the hobby chart's ceiling is on earth", 58.61, -7.2, true},
	{"a berth west of the hobby chart's edge is on earth", 58.3, -8.2, true},
	{"the poles and the antimeridian are on earth", 90.0, 180.0, true},
	{"a latitude past the pole is not", 99.0, 99.0, false},
	{"a latitude past the south pole is not", -99.0, -7.2, false},
	{"a longitude past the antimeridian is not", 58.3, -199.0, false},
}

// A coord is held to the globe and nothing narrower, and a null one stays
// uncharted. Which chart window a berth renders in is the site's business.
func TestCoordValidatesEarthRangeOnProjectAndNote(t *testing.T) {
	for _, edge := range earthEdges {
		projects := newProjects()
		light, err := projects.Create(domain.Project{Title: "A light", Coord: &domain.Coord{Lat: edge.lat, Lon: edge.lon}})

		if edge.valid != (nil == err) {
			t.Fatalf("%s: project create returned %v", edge.name, err)
		}

		if edge.valid {
			stored := projects.Read(light.Id)

			if nil == stored.Coord || edge.lat != stored.Coord.Lat || edge.lon != stored.Coord.Lon {
				t.Fatalf("%s: an on-earth project coord must store unmoved, got %+v", edge.name, stored.Coord)
			}
		}

		if !edge.valid {
			all, _ := projects.List(false, 0)

			if 0 != len(all) {
				t.Fatalf("%s: a rejected project create must persist nothing, found %d", edge.name, len(all))
			}
		}

		notes := newNotes()
		thought, err := notes.Create(domain.Note{Title: "A thought", Coord: &domain.Coord{Lat: edge.lat, Lon: edge.lon}})

		if edge.valid != (nil == err) {
			t.Fatalf("%s: note create returned %v", edge.name, err)
		}

		if edge.valid {
			stored := notes.Read(thought.Id)

			if nil == stored.Coord || edge.lat != stored.Coord.Lat || edge.lon != stored.Coord.Lon {
				t.Fatalf("%s: an on-earth note coord must store unmoved, got %+v", edge.name, stored.Coord)
			}
		}

		if !edge.valid {
			all, _ := notes.List(false, 0)

			if 0 != len(all) {
				t.Fatalf("%s: a rejected note create must persist nothing, found %d", edge.name, len(all))
			}
		}
	}

	projects := newProjects()
	uncharted, _ := projects.Create(domain.Project{Title: "An unplaced light"})

	if nil != projects.Read(uncharted.Id).Coord {
		t.Fatalf("an unplaced light must keep coord nil, got %+v", projects.Read(uncharted.Id).Coord)
	}

	notes := newNotes()
	unplaced, _ := notes.Create(domain.Note{Title: "An unplaced thought"})

	if nil != notes.Read(unplaced.Id).Coord {
		t.Fatalf("an unplaced thought must keep coord nil, got %+v", notes.Read(unplaced.Id).Coord)
	}
}

// dressingKeysPresent reports whether a cleared plate and caption still reach
// the stored document as keys. The fakes hand a Go struct straight back, so the
// only place the omitempty decision is observable is the marshalled document,
// the same way the gauge absence test checks the wire.
func dressingKeysPresent(t *testing.T, entity interface{}) (bool, bool) {
	t.Helper()

	doc, err := bson.Marshal(entity)

	if nil != err {
		t.Fatalf("entity did not marshal to bson: %v", err)
	}

	var raw bson.M

	if err := bson.Unmarshal(doc, &raw); nil != err {
		t.Fatalf("bson did not unmarshal: %v", err)
	}

	_, plate := raw["plate"]
	_, capKey := raw["cap"]

	return plate, capKey
}

// The berth fields carry no omitempty because clearing one is a real edit: a
// replace write drops the whole document, so a plate of 0 and an empty caption
// have to travel as keys rather than vanishing and leaving the old pair to be
// read back off the previous document.
func TestClearingTheDressingSurvivesAReplaceWrite(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "A light", Plate: 2, Cap: "a caption"})
	cleared, _ := projects.Update(domain.Project{Id: light.Id, Title: "A light"})

	if plate, capKey := dressingKeysPresent(t, cleared); !plate || !capKey {
		t.Fatalf("a cleared project dressing must still carry both keys, got plate %v / cap %v", plate, capKey)
	}

	if stored := projects.Read(light.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a project's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "A thought", Plate: 3, Cap: "a caption"})
	clearedNote, _ := notes.Update(domain.Note{Id: thought.Id, Title: "A thought"})

	if plate, capKey := dressingKeysPresent(t, clearedNote); !plate || !capKey {
		t.Fatalf("a cleared note dressing must still carry both keys, got plate %v / cap %v", plate, capKey)
	}

	if stored := notes.Read(thought.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a note's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}

	hobbies := newHobbies()
	ship, _ := hobbies.Create(domain.Hobby{Name: "A ship", State: domain.StateMoored, Plate: 1, Cap: "a caption"})
	clearedShip, _ := hobbies.Update(domain.Hobby{Id: ship.Id, Name: "A ship", State: domain.StateMoored})

	if plate, capKey := dressingKeysPresent(t, clearedShip); !plate || !capKey {
		t.Fatalf("a cleared ship dressing must still carry both keys, got plate %v / cap %v", plate, capKey)
	}

	if stored := hobbies.Read(ship.Id); 0 != stored.Plate || "" != stored.Cap {
		t.Fatalf("clearing a ship's dressing must survive the write, got %d / %q", stored.Plate, stored.Cap)
	}
}

// assertBerthKeysOnTheWire checks the exact JSON a consumer receives for an
// entity carrying no berth data: all three keys present, with null / 0 / "".
// Comparing the raw bytes rather than a decoded value is the point, since an
// absent key and a zero one both decode to the same Go value and only the raw
// document tells them apart.
func assertBerthKeysOnTheWire(t *testing.T, what string, entity interface{}) {
	t.Helper()

	body, err := json.Marshal(entity)

	if nil != err {
		t.Fatalf("%s did not marshal to json: %v", what, err)
	}

	var raw map[string]json.RawMessage

	if err := json.Unmarshal(body, &raw); nil != err {
		t.Fatalf("%s json did not unmarshal: %v", what, err)
	}

	for key, want := range map[string]string{"coord": "null", "plate": "0", "cap": `""`} {
		got, present := raw[key]

		if !present {
			t.Fatalf("%s must always serialize %q, got %s", what, key, body)
		}

		if want != string(got) {
			t.Fatalf("%s %q must serialize as %s, got %s", what, key, want, got)
		}
	}
}

// The consumers build against a promise the contract's amendment makes: plate,
// cap and coord always reach the wire, so an admin or a site never has to tell
// an absent key from a cleared one. omitempty on any of the three would break
// that silently, which is exactly what this guards.
func TestBerthKeysAlwaysReachTheWire(t *testing.T) {
	projects := newProjects()
	light, _ := projects.Create(domain.Project{Title: "An undressed light"})
	assertBerthKeysOnTheWire(t, "project", projects.Read(light.Id))

	notes := newNotes()
	thought, _ := notes.Create(domain.Note{Title: "An undressed thought"})
	assertBerthKeysOnTheWire(t, "note", notes.Read(thought.Id))

	hobbies := newHobbies()
	ship, _ := hobbies.Create(domain.Hobby{Name: "An undressed ship", State: domain.StateMoored})
	assertBerthKeysOnTheWire(t, "hobby", hobbies.Read(ship.Id))
}
