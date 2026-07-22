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

// chartEdges walks each side of the clamp band: a bearing past a bound snaps to
// the bound, an in-band component rides through untouched.
var chartEdges = []struct {
	name             string
	lat, lon         float64
	wantLat, wantLon float64
}{
	{"lat below the south bound", 50.0, -7.0, 57.82, -7.0},
	{"lat above the north bound", 60.0, -7.0, 58.56, -7.0},
	{"lon west of the west bound", 58.0, -20.0, 58.0, -7.94},
	{"lon east of the east bound", 58.0, 0.0, 58.0, -6.59},
}

func TestBearingsClampToChartOnCreate(t *testing.T) {
	hobbies := newHobbies()

	for _, edge := range chartEdges {
		saved, err := hobbies.Create(domain.Hobby{
			Name:  "Piano",
			State: domain.StateAdrift,
			Coord: &domain.Coord{Lat: edge.lat, Lon: edge.lon},
		})

		if nil != err {
			t.Fatalf("%s: create failed: %v", edge.name, err)
		}

		stored := hobbies.Read(saved.Id)

		if edge.wantLat != stored.Coord.Lat || edge.wantLon != stored.Coord.Lon {
			t.Fatalf("%s: got %+v, want lat %v lon %v", edge.name, stored.Coord, edge.wantLat, edge.wantLon)
		}
	}
}

func TestBearingsClampToChartOnUpdate(t *testing.T) {
	hobbies := newHobbies()

	for _, edge := range chartEdges {
		// seed a charted ship in-band, then steer it off-window on update
		saved, _ := hobbies.Create(domain.Hobby{
			Name:  "Piano",
			State: domain.StateAdrift,
			Coord: &domain.Coord{Lat: 58.0, Lon: -7.0},
		})

		if _, err := hobbies.Update(domain.Hobby{
			Id:    saved.Id,
			Name:  "Piano",
			State: domain.StateAdrift,
			Coord: &domain.Coord{Lat: edge.lat, Lon: edge.lon},
		}); nil != err {
			t.Fatalf("%s: update failed: %v", edge.name, err)
		}

		stored := hobbies.Read(saved.Id)

		if edge.wantLat != stored.Coord.Lat || edge.wantLon != stored.Coord.Lon {
			t.Fatalf("%s: got %+v, want lat %v lon %v", edge.name, stored.Coord, edge.wantLat, edge.wantLon)
		}
	}
}

func TestCoordAndFromClampIndependently(t *testing.T) {
	hobbies := newHobbies()

	// coord sits in-band and must ride through, while from runs off-window west
	// and snaps to the bound: each bearing is clamped on its own
	saved, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 58.1, Lon: -7.2},
		From:  &domain.Coord{Lat: 99.0, Lon: -99.0},
	})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 58.1 != stored.Coord.Lat || -7.2 != stored.Coord.Lon {
		t.Fatalf("an in-band coord must ride through unclamped, got %+v", stored.Coord)
	}

	if 58.56 != stored.From.Lat || -7.94 != stored.From.Lon {
		t.Fatalf("an off-window from must clamp to the bounds, got %+v", stored.From)
	}
}

func TestNullBearingsUntouchedByClamp(t *testing.T) {
	hobbies := newHobbies()

	// an uncharted ship carries no bearing; the clamp must leave both null
	saved, err := hobbies.Create(domain.Hobby{Name: "Kite", State: domain.StateMarooned})

	if nil != err {
		t.Fatalf("uncharted create failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if nil != stored.Coord || nil != stored.From {
		t.Fatalf("clamp must leave a null bearing null, got %+v / %+v", stored.Coord, stored.From)
	}

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Kite", State: domain.StateMarooned}); nil != err {
		t.Fatalf("uncharted update failed: %v", err)
	}

	stored = hobbies.Read(saved.Id)

	if nil != stored.Coord || nil != stored.From {
		t.Fatalf("clamp must leave a null bearing null on update, got %+v / %+v", stored.Coord, stored.From)
	}
}

func TestClampComposesWithStateValidation(t *testing.T) {
	hobbies := newHobbies()

	// the state gate runs first: a bad state rejects the create before the clamp
	// ever touches the coord, and nothing is stored
	if _, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: "sunk",
		Coord: &domain.Coord{Lat: 99.0, Lon: 99.0},
	}); nil == err {
		t.Fatalf("expected an out-of-set state to reject the create")
	}

	all, _ := hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("a rejected create must store nothing, found %d hobbies", len(all))
	}

	// a valid state carries an off-window coord through, clamped to the bounds
	saved, err := hobbies.Create(domain.Hobby{
		Name:  "Piano",
		State: domain.StateAdrift,
		Coord: &domain.Coord{Lat: 99.0, Lon: 99.0},
	})

	if nil != err {
		t.Fatalf("valid create with an off-window coord failed: %v", err)
	}

	stored := hobbies.Read(saved.Id)

	if 58.56 != stored.Coord.Lat || -6.59 != stored.Coord.Lon {
		t.Fatalf("an off-window coord must store clamped, got %+v", stored.Coord)
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
