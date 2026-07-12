package service_test

import (
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

func TestMarkerAcceptsEmptyAndEachVocabularyValue(t *testing.T) {
	hobbies := newHobbies()

	// empty is the valid default; the site falls back to its default
	// headstone, same as an absent stamp or light
	for _, marker := range []string{"", "stone", "sticks", "driftwood", "cairn", "buoy", "lamp"} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Grave", Marker: marker}); nil != err {
			t.Fatalf("expected marker %q accepted, got %v", marker, err)
		}
	}
}

func TestMarkerRejectsOutOfSetValue(t *testing.T) {
	hobbies := newHobbies()

	if _, err := hobbies.Create(domain.Hobby{Name: "Grave", Marker: "obelisk"}); nil == err {
		t.Fatalf("expected an out-of-set marker rejected")
	}

	// nothing rejected may have been written
	all, _ := hobbies.List(false)

	if 0 != len(all) {
		t.Fatalf("rejected create must persist nothing, found %d hobbies", len(all))
	}

	// the update path rejects too, and the stored marker survives untouched
	saved, _ := hobbies.Create(domain.Hobby{Name: "Grave", Marker: "cairn"})

	if _, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Grave", Marker: "obelisk"}); nil == err {
		t.Fatalf("expected update to reject an out-of-set marker")
	}

	stored := hobbies.Read(saved.Id)

	if "cairn" != stored.Marker {
		t.Fatalf("rejected update must leave the stored marker intact, got %q", stored.Marker)
	}
}

func TestWearAcceptsBoundsAndRejectsOutOfRange(t *testing.T) {
	hobbies := newHobbies()

	// 0 is a real fraction (fresh stone), not an absent value; it must be
	// accepted and must survive a replace write the same way featured/flagship do
	for _, wear := range []float64{0, 0.5, 1} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Weathered", Wear: wear}); nil != err {
			t.Fatalf("expected wear %v accepted, got %v", wear, err)
		}
	}

	for _, wear := range []float64{-0.1, 1.1} {
		if _, err := hobbies.Create(domain.Hobby{Name: "Weathered", Wear: wear}); nil == err {
			t.Fatalf("expected wear %v rejected", wear)
		}
	}

	saved, _ := hobbies.Create(domain.Hobby{Name: "Weathered", Wear: 0.75})

	cleared, err := hobbies.Update(domain.Hobby{Id: saved.Id, Name: "Weathered", Wear: 0})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if 0 != cleared.Wear {
		t.Fatalf("expected wear 0 to survive the replace write, got %v", cleared.Wear)
	}

	if 0 != hobbies.Read(saved.Id).Wear {
		t.Fatalf("expected the stored document to also read wear 0")
	}
}
