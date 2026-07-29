package service_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
	"go.mongodb.org/mongo-driver/bson"
)

// newHobbies wires a hobby service onto an in-memory fake repo plus the
// shared activity log, so the real business logic runs end-to-end.
func newHobbies() in_port.HobbyCRUDService {
	hobbies, _ := newHobbiesWithNotes()

	return hobbies
}

// newHobbiesWithNotes wires a hobby service like newHobbies but hands back
// the note repo directly too, so a test can seed notes for the tie check
// without going through a second, unrelated NoteCRUDService.
func newHobbiesWithNotes() (in_port.HobbyCRUDService, out_port.NoteRepo) {
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	notes := out_adapter.NewNoteFakeOutAdapter()

	return service.NewHobbyCRUDService(out_adapter.NewHobbyFakeOutAdapter(), notes, activity), notes
}

func TestStateAcceptsEachVocabularyValue(t *testing.T) {
	hobbies := newHobbies()

	for _, state := range []string{domain.StateMoored, domain.StatePort, domain.StateAdrift, domain.StateMarooned, domain.StateInkspill} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Piano", State: state}); nil != err {
			t.Fatalf("expected state %q accepted, got %v", state, err)
		}
	}
}

func TestStateRejectsEmptyAndOutOfSetValue(t *testing.T) {
	hobbies := newHobbies()

	// empty is not a state: every ship stands somewhere
	for _, state := range []string{"", "sunk", "MOORED"} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Piano", State: state}); nil == err {
			t.Fatalf("expected state %q rejected", state)
		}
	}

	// nothing rejected may have been written
	all, _ := hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("rejected create must persist nothing, found %d hobbies", len(all))
	}

	// the update path rejects too, and the stored state survives untouched
	saved, _ := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored})

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Piano", State: "sunk"}); nil == err {
		t.Fatalf("expected update to reject an out-of-set state")
	}

	stored := hobbies.Read(saved.Id)

	if domain.StateMoored != stored.State {
		t.Fatalf("rejected update must leave the stored state intact, got %q", stored.State)
	}
}

func TestCoordAndFromRoundTrip(t *testing.T) {
	hobbies := newHobbies()

	// a charted ship carries a coord and a wake origin
	charted, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.22, Lon: -7.5},
		From:  &domain.Coord{Lat: 58.05, Lon: -7.1},
	})

	if nil != err {
		t.Fatalf("charted create failed: %v", err)
	}

	back := hobbies.Read(charted.Id)

	if nil == back.Coord || 58.22 != back.Coord.Lat || -7.5 != back.Coord.Lon {
		t.Fatalf("coord did not round-trip, got %+v", back.Coord)
	}

	if nil == back.From || 58.05 != back.From.Lat || -7.1 != back.From.Lon {
		t.Fatalf("from did not round-trip, got %+v", back.From)
	}

	// an uncharted ship leaves both nil, which must serialize as JSON null
	uncharted, err := hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateMarooned})

	if nil != err {
		t.Fatalf("uncharted create failed: %v", err)
	}

	stored := hobbies.Read(uncharted.Id)

	if nil != stored.Coord || nil != stored.From {
		t.Fatalf("an uncharted ship must keep coord and from nil, got %+v / %+v", stored.Coord, stored.From)
	}

	body, err := json.Marshal(stored)

	if nil != err {
		t.Fatalf("hobby did not marshal: %v", err)
	}

	if !strings.Contains(string(body), `"coord":null`) || !strings.Contains(string(body), `"from":null`) {
		t.Fatalf("an uncharted ship must serialize coord and from as null, got %s", body)
	}

	charge, _ := json.Marshal(back)

	if !strings.Contains(string(charge), `"coord":{"lat":58.22,"lon":-7.5}`) {
		t.Fatalf("a charted ship must serialize coord as an object, got %s", charge)
	}
}

// A hobby's bearings are held to the globe and nothing narrower, the same gate
// a project's and a note's get. The chart window the log renders in is the
// site's business, so a berth above the retired band's ceiling stores unmoved.
func TestBearingsValidateEarthRangeOnCreate(t *testing.T) {
	for _, edge := range earthEdges {
		hobbies := newHobbies()
		saved, err := hobbies.Create(domain.Hobby{
			Name:  "Piano",
			State: domain.StateAdrift,
			Coord: &domain.Coord{Lat: edge.lat, Lon: edge.lon},
		})

		if edge.valid != (nil == err) {
			t.Fatalf("%s: hobby create returned %v", edge.name, err)
		}

		if !edge.valid {
			all, _ := hobbies.List(false)

			if 0 != len(all) {
				t.Fatalf("%s: a rejected hobby create must persist nothing, found %d", edge.name, len(all))
			}

			continue
		}

		stored := hobbies.Read(saved.Id)

		if nil == stored.Coord || edge.lat != stored.Coord.Lat || edge.lon != stored.Coord.Lon {
			t.Fatalf("%s: an on-earth bearing must store unmoved, got %+v", edge.name, stored.Coord)
		}
	}
}

// The update path gets the same gate as create: an off-earth bearing must never
// reach the store by the back door of an edit, and a rejected edit leaves the
// stored bearing exactly as it was.
func TestBearingsValidateEarthRangeOnUpdate(t *testing.T) {
	hobbies := newHobbies()
	saved, _ := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.0, Lon: -7.0},
	})

	// a berth above the retired band's ceiling is a legitimate edit
	if _, err := hobbies.Update(domain.Hobby{
		Id:    saved.Id,
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.61, Lon: -8.2},
	}); nil != err {
		t.Fatalf("hobby update with an on-earth bearing failed: %v", err)
	}

	if stored := hobbies.Read(saved.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat || -8.2 != stored.Coord.Lon {
		t.Fatalf("an on-earth bearing must store unmoved on update, got %+v", stored.Coord)
	}

	if _, err := hobbies.Update(domain.Hobby{
		Id:    saved.Id,
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 99.0, Lon: 99.0},
	}); nil == err {
		t.Fatalf("expected hobby update to reject an off-earth bearing")
	}

	if stored := hobbies.Read(saved.Id); nil == stored.Coord || 58.61 != stored.Coord.Lat {
		t.Fatalf("a rejected update must leave the stored bearing intact, got %+v", stored.Coord)
	}
}

func TestCoordAndFromValidateIndependently(t *testing.T) {
	hobbies := newHobbies()

	// coord is a real point on earth and from is not: each bearing is gated on
	// its own, and one bad half rejects the whole write
	if _, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.1, Lon: -7.2},
		From:  &domain.Coord{Lat: 99.0, Lon: -99.0},
	}); nil == err {
		t.Fatalf("expected an off-earth wake origin to reject the create")
	}

	all, _ := hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("a rejected create must store nothing, found %d hobbies", len(all))
	}

	saved, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.1, Lon: -7.2},
		From:  &domain.Coord{Lat: 58.61, Lon: -8.2},
	})

	if nil != err {
		t.Fatalf("create with two on-earth bearings failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 58.1 != stored.Coord.Lat || -7.2 != stored.Coord.Lon {
		t.Fatalf("an on-earth coord must store unmoved, got %+v", stored.Coord)
	}

	if 58.61 != stored.From.Lat || -8.2 != stored.From.Lon {
		t.Fatalf("an on-earth from must store unmoved, got %+v", stored.From)
	}
}

func TestNullBearingsRideThroughValidation(t *testing.T) {
	hobbies := newHobbies()

	// an uncharted ship carries no bearing; the gate must leave both null
	saved, err := hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateMarooned})

	if nil != err {
		t.Fatalf("uncharted create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if nil != stored.Coord || nil != stored.From {
		t.Fatalf("validation must leave a null bearing null, got %+v / %+v", stored.Coord, stored.From)
	}

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Kite", State: domain.StateMarooned}); nil != err {
		t.Fatalf("uncharted update failed: %v", err)
	}

	stored = hobbies.Read(saved.Id)

	if nil != stored.Coord || nil != stored.From {
		t.Fatalf("validation must leave a null bearing null on update, got %+v / %+v", stored.Coord, stored.From)
	}
}

func TestBearingValidationComposesWithStateValidation(t *testing.T) {
	hobbies := newHobbies()

	// the state gate runs first: a bad state rejects the create before the
	// bearing is ever looked at, and nothing is stored
	if _, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: "sunk",
		Coord: &domain.Coord{Lat: 58.1, Lon: -7.2},
	}); nil == err {
		t.Fatalf("expected an out-of-set state to reject the create")
	}

	all, _ := hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("a rejected create must store nothing, found %d hobbies", len(all))
	}

	// a valid state still rejects on an off-earth bearing
	if _, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 99.0, Lon: 99.0},
	}); nil == err {
		t.Fatalf("expected an off-earth bearing to reject the create")
	}

	all, _ = hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("a rejected create must store nothing, found %d hobbies", len(all))
	}
}

func TestTagsSurviveCreateAndUpdate(t *testing.T) {
	hobbies := newHobbies()

	// the home currently-learning card renders tags, so they must ride every
	// write untouched
	saved, err := hobbies.Create(domain.Hobby{Name: "Plex", State: domain.StateMoored, Tags: []string{"plex", "htpc"}})

	if nil != err {
		t.Fatalf("tagged create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 2 != len(stored.Tags) || "plex" != stored.Tags[0] || "htpc" != stored.Tags[1] {
		t.Fatalf("tags did not round-trip the create, got %+v", stored.Tags)
	}

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Plex", State: domain.StateMoored, Tags: []string{"plex"}}); nil != err {
		t.Fatalf("tagged update failed: %v", err)
	}

	stored = hobbies.Read(saved.Id)

	if 1 != len(stored.Tags) || "plex" != stored.Tags[0] {
		t.Fatalf("tags did not survive the replace write, got %+v", stored.Tags)
	}
}

func TestHobbyNoteIdsRejectsUnknownId(t *testing.T) {
	hobbies, notes := newHobbiesWithNotes()

	noteID, err := notes.Add(domain.Note{Title: "Journal entry"})

	if nil != err {
		t.Fatalf("seed note failed: %v", err)
	}

	tied, err := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored, NoteIds: []string{noteID}})

	if nil != err {
		t.Fatalf("expected a known note id accepted, got %v", err)
	}

	stored := hobbies.Read(tied.Id)

	if 1 != len(stored.NoteIds) || noteID != stored.NoteIds[0] {
		t.Fatalf("noteIds did not round-trip the create, got %+v", stored.NoteIds)
	}

	if _, err := hobbies.Create(domain.Hobby{Name: "Untied", State: domain.StateMoored, NoteIds: []string{"nope"}}); nil == err {
		t.Fatalf("expected an unknown note id rejected")
	}

	// the update path rejects too, and the stored noteIds survive untouched
	if _, err := hobbies.Update(domain.Hobby{Id: tied.Id, Name: "Piano", State: domain.StateMoored, NoteIds: []string{"nope"}}); nil == err {
		t.Fatalf("expected update to reject an unknown note id")
	}

	stored = hobbies.Read(tied.Id)

	if 1 != len(stored.NoteIds) || noteID != stored.NoteIds[0] {
		t.Fatalf("rejected update must leave the stored noteIds intact, got %+v", stored.NoteIds)
	}
}

func TestHobbyNoteIdsAbsentStaysAbsentOnTheWire(t *testing.T) {
	hobbies := newHobbies()

	saved, err := hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateMarooned})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 0 != len(stored.NoteIds) {
		t.Fatalf("expected no noteIds, got %+v", stored.NoteIds)
	}

	// the json tag round-trips noteIds like every other Hobby field (no
	// omitempty there, same as Project.NoteIds); the store-side omission is on
	// the bson tag instead, matching an absent tags
	doc, err := bson.Marshal(stored)

	if nil != err {
		t.Fatalf("hobby did not marshal to bson: %v", err)
	}

	var raw bson.M

	if err := bson.Unmarshal(doc, &raw); nil != err {
		t.Fatalf("bson did not unmarshal: %v", err)
	}

	if _, present := raw["noteIds"]; present {
		t.Fatalf("an absent noteIds must omit the field in the stored document, got %+v", raw)
	}
}

// gaugeEdges walks each side of the gauge's clamp band: an out-of-range value
// snaps to the bound, an in-band value rides through untouched.
var gaugeEdges = []struct {
	name string
	in   int
	want int
}{
	{"below the low bound", -20, 0},
	{"above the high bound", 250, 100},
	{"in band", 42, 42},
}

func TestGaugeClampsHighAndLow(t *testing.T) {
	for _, edge := range gaugeEdges {
		hobbies := newHobbies()
		gauge := edge.in

		saved, err := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored, Gauge: &gauge})

		if nil != err {
			t.Fatalf("%s: create failed: %v", edge.name, err)
		}

		stored := hobbies.Read(saved.Id)

		if nil == stored.Gauge || edge.want != *stored.Gauge {
			t.Fatalf("%s: got %+v, want %v", edge.name, stored.Gauge, edge.want)
		}

		update := edge.in

		if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Piano", State: domain.StateMoored, Gauge: &update}); nil != err {
			t.Fatalf("%s: update failed: %v", edge.name, err)
		}

		stored = hobbies.Read(saved.Id)

		if nil == stored.Gauge || edge.want != *stored.Gauge {
			t.Fatalf("%s (update): got %+v, want %v", edge.name, stored.Gauge, edge.want)
		}
	}
}

func TestGaugeAbsentStaysAbsentAndDistinctFromZero(t *testing.T) {
	hobbies := newHobbies()

	unrated, err := hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateMarooned})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	stored := hobbies.Read(unrated.Id)

	if nil != stored.Gauge {
		t.Fatalf("expected no gauge on an unrated hobby, got %+v", stored.Gauge)
	}

	// the json tag omits gauge on absence too (unlike coord/from's json null),
	// so check the wire the same way the noteIds absence test does
	doc, err := bson.Marshal(stored)

	if nil != err {
		t.Fatalf("hobby did not marshal to bson: %v", err)
	}

	var raw bson.M

	if err := bson.Unmarshal(doc, &raw); nil != err {
		t.Fatalf("bson did not unmarshal: %v", err)
	}

	if _, present := raw["gauge"]; present {
		t.Fatalf("an absent gauge must omit the field in the stored document, got %+v", raw)
	}

	// a gauge explicitly set to zero is a real rating, not a missing one
	zero := 0
	rated, err := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored, Gauge: &zero})

	if nil != err {
		t.Fatalf("create with a zero gauge failed: %v", err)
	}

	storedZero := hobbies.Read(rated.Id)

	if nil == storedZero.Gauge || 0 != *storedZero.Gauge {
		t.Fatalf("expected an explicit zero gauge to round-trip as zero, not absent, got %+v", storedZero.Gauge)
	}
}

func TestListActiveOnlyIsMooredOnly(t *testing.T) {
	hobbies := newHobbies()

	moored, _ := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored})
	hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateAdrift})

	only, err := hobbies.List(true)

	if nil != err {
		t.Fatalf("active list failed: %v", err)
	}

	if 1 != len(only) || moored.Id != only[0].Id {
		t.Fatalf("active list must return only the moored ship, got %+v", only)
	}

	all, _ := hobbies.List(false)

	if 2 != len(all) {
		t.Fatalf("the full list must return every ship, got %d", len(all))
	}
}

// A ship's gallery is a project's gallery: the same trim, the same rejection of
// an empty name, the same six-print cap. The plate indexes into it, so a
// divergence here would render a different print on the hobby sheet than on the
// light's.
func TestGalleryTrimmedAndEmptyRejected(t *testing.T) {
	hobbies := newHobbies()

	saved, err := hobbies.Create(domain.Hobby{
		Name:   "Piano",
		State:  domain.StateMoored,
		Images: []string{"  one.jpg  ", "two.jpg"},
	})

	if nil != err {
		t.Fatalf("gallery create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 2 != len(stored.Images) || "one.jpg" != stored.Images[0] || "two.jpg" != stored.Images[1] {
		t.Fatalf("expected gallery names trimmed on store, got %v", stored.Images)
	}

	for _, bad := range [][]string{{""}, {"one.jpg", "   "}} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Piano", State: domain.StateMoored, Images: bad}); nil == err {
			t.Fatalf("expected an empty gallery name to reject the create, got %v", bad)
		}
	}

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Piano", State: domain.StateMoored, Images: []string{" "}}); nil == err {
		t.Fatalf("expected an empty gallery name to reject the update")
	}

	if kept := hobbies.Read(saved.Id); 2 != len(kept.Images) {
		t.Fatalf("a rejected update must leave the stored gallery intact, got %v", kept.Images)
	}
}

func TestGalleryCappedAtSix(t *testing.T) {
	hobbies := newHobbies()
	six := []string{"1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg", "6.jpg"}
	seven := append(append([]string{}, six...), "7.jpg")

	if _, err := hobbies.Create(domain.Hobby{Name: "Full album", State: domain.StateMoored, Images: six}); nil != err {
		t.Fatalf("six prints must be accepted, got %v", err)
	}

	if _, err := hobbies.Create(domain.Hobby{Name: "Overfull", State: domain.StateMoored, Images: seven}); nil == err {
		t.Fatalf("expected a seventh print to reject the create")
	}
}

// The consumers tell a ship with no gallery from one whose gallery was emptied
// by the raw json, the same way they do for a light: nil serializes as null and
// an empty gallery as []. Comparing the raw bytes is the point, since both
// decode to the same zero-length Go slice.
func TestGalleryNullAndEmptyAreDistinctOnTheWire(t *testing.T) {
	for _, subject := range []struct {
		name   string
		images []string
		want   string
	}{
		{"a ship with no gallery", nil, "null"},
		{"a ship whose gallery was emptied", []string{}, "[]"},
	} {
		hobby := domain.Hobby{Name: "Piano", State: domain.StateMoored, Images: subject.images}
		body, err := json.Marshal(hobby)

		if nil != err {
			t.Fatalf("%s did not marshal: %v", subject.name, err)
		}

		var raw map[string]json.RawMessage

		if err := json.Unmarshal(body, &raw); nil != err {
			t.Fatalf("%s json did not unmarshal: %v", subject.name, err)
		}

		got, present := raw["images"]

		if !present {
			t.Fatalf("%s must always serialize images, got %s", subject.name, body)
		}

		if subject.want != string(got) {
			t.Fatalf("%s must serialize images as %s, got %s", subject.name, subject.want, got)
		}

		light, err := json.Marshal(domain.Project{Title: "A light", Images: subject.images})

		if nil != err {
			t.Fatalf("the matching light did not marshal: %v", err)
		}

		var lightRaw map[string]json.RawMessage

		if err := json.Unmarshal(light, &lightRaw); nil != err {
			t.Fatalf("the matching light's json did not unmarshal: %v", err)
		}

		if string(lightRaw["images"]) != string(got) {
			t.Fatalf("%s must serialize images exactly as a light does, got %s against %s", subject.name, got, lightRaw["images"])
		}
	}
}

// An older ship document carries no gallery key at all and must read as no
// gallery rather than failing to parse, the same promise the berth fields make.
func TestGalleryAbsentParsesAsNoGallery(t *testing.T) {
	var hobby domain.Hobby

	if err := json.Unmarshal([]byte(`{"name":"An old ship","state":"moored"}`), &hobby); nil != err {
		t.Fatalf("an older hobby document must still parse, got %v", err)
	}

	if nil != hobby.Images {
		t.Fatalf("an older hobby must read as having no gallery, got %v", hobby.Images)
	}
}
