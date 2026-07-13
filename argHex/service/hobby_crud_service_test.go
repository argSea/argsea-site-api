package service_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newHobbies wires a hobby service onto an in-memory fake repo plus the
// shared activity log, so the real business logic runs end-to-end.
func newHobbies() in_port.HobbyCRUDService {
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	return service.NewHobbyCRUDService(out_adapter.NewHobbyFakeOutAdapter(), activity)
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
