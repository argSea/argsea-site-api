package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newCarvings(t *testing.T) (in_port.CarvingService, in_port.ActivityService) {
	t.Helper()

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	carvings := service.NewCarvingService(out_adapter.NewCarvingFakeOutAdapter(), activity)

	if err := carvings.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	return carvings, activity
}

func carvingBySpot(t *testing.T, carvings in_port.CarvingService) map[string]domain.Carving {
	t.Helper()

	all, err := carvings.List()

	if nil != err {
		t.Fatalf("list failed: %v", err)
	}

	out := map[string]domain.Carving{}

	for _, carving := range all {
		for _, spot := range carving.BoltedTo {
			if _, dup := out[spot]; dup {
				t.Fatalf("two carvings hold the %s spot", spot)
			}

			out[spot] = carving
		}
	}

	return out
}

func TestSeedPlantsSevenBuiltinCarvingsBoltedToTheirOwnSpot(t *testing.T) {
	carvings, _ := newCarvings(t)

	all, err := carvings.List()

	if nil != err {
		t.Fatalf("list failed: %v", err)
	}

	if 7 != len(all) {
		t.Fatalf("expected seven seeded carvings, got %d", len(all))
	}

	bySpot := carvingBySpot(t, carvings)

	for spot := range domain.CarvingSpots {
		carving, ok := bySpot[spot]

		if !ok || !carving.Builtin || "" == carving.Svg {
			t.Fatalf("spot %s should hold a builtin seed with svg, got %+v", spot, carving)
		}
	}
}

func TestCarvingSeedIsIdempotent(t *testing.T) {
	carvings, _ := newCarvings(t)

	// second boot: the bench is populated, the seed must not touch it
	if err := carvings.Seed(); nil != err {
		t.Fatalf("re-seed failed: %v", err)
	}

	all, _ := carvings.List()

	if 7 != len(all) {
		t.Fatalf("re-seeding grew the bench to %d carvings", len(all))
	}
}

func TestCarvingSeedLeavesANonEmptyCollectionAlone(t *testing.T) {
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	carvings := service.NewCarvingService(out_adapter.NewCarvingFakeOutAdapter(), activity)

	// a lone draft in the collection means a keeper has been here; anything
	// present at all suppresses the seed, not just a full bench
	if _, err := carvings.Create(domain.Carving{Name: "hand-carved", Svg: "<svg></svg>"}); nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if err := carvings.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	all, _ := carvings.List()

	if 1 != len(all) {
		t.Fatalf("the seed must not run against a populated collection, got %d carvings", len(all))
	}
}

func TestCreateRejectsEmptyName(t *testing.T) {
	carvings, _ := newCarvings(t)

	if _, err := carvings.Create(domain.Carving{Svg: "<svg></svg>"}); nil == err {
		t.Fatalf("expected an empty name to be rejected")
	}
}

func TestCreateAcceptsSvgAtTheSizeCap(t *testing.T) {
	carvings, _ := newCarvings(t)

	svg := "<svg>" + strings.Repeat("a", 100*1024-11) + "</svg>" // exactly 100KB

	saved, err := carvings.Create(domain.Carving{Name: "big", Svg: svg})

	if nil != err {
		t.Fatalf("expected a carving at the size cap to be accepted, got %v", err)
	}

	if len(svg) != len(saved.Svg) {
		t.Fatalf("svg was not stored intact")
	}
}

func TestCreateRejectsSvgOverTheSizeCap(t *testing.T) {
	carvings, _ := newCarvings(t)

	svg := "<svg>" + strings.Repeat("a", 100*1024-10) + "</svg>" // one byte over 100KB

	if _, err := carvings.Create(domain.Carving{Name: "too big", Svg: svg}); nil == err {
		t.Fatalf("expected an oversized svg to be rejected")
	}
}

func TestUpdateAcceptsUnchangedNameAndSvgOnABuiltin(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	saved, err := carvings.Update(domain.Carving{Id: seed.Id, Name: seed.Name, Svg: seed.Svg})

	if nil != err {
		t.Fatalf("expected an unchanged builtin update to be accepted, got %v", err)
	}

	if seed.Name != saved.Name || seed.Svg != saved.Svg {
		t.Fatalf("update must not have moved the builtin's name/svg: %+v", saved)
	}
}

func TestUpdateRejectsBuiltinNameChange(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	_, err := carvings.Update(domain.Carving{Id: seed.Id, Name: "defaced", Svg: seed.Svg})

	if !errors.Is(err, in_port.ErrCarvingBuiltin) {
		t.Fatalf("expected the builtin guard on a name change, got %v", err)
	}
}

func TestUpdateRejectsBuiltinSvgChange(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	_, err := carvings.Update(domain.Carving{Id: seed.Id, Name: seed.Name, Svg: "<svg>defaced</svg>"})

	if !errors.Is(err, in_port.ErrCarvingBuiltin) {
		t.Fatalf("expected the builtin guard on an svg change, got %v", err)
	}
}

func TestUpdateRejectsBlankSvgOnABoltedCarving(t *testing.T) {
	carvings, _ := newCarvings(t)

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(fresh.Id, domain.SpotBoat); nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	if _, err := carvings.Update(domain.Carving{Id: fresh.Id, Name: fresh.Name, Svg: ""}); !errors.Is(err, in_port.ErrCarvingBolted) {
		t.Fatalf("expected the bolted guard on a blanked svg, got %v", err)
	}
}

func TestUpdateAcceptsBlankSvgOnANeverBoltedCarving(t *testing.T) {
	carvings, _ := newCarvings(t)

	fresh, err := carvings.Create(domain.Carving{Name: "spare", Svg: "<svg>spare</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	saved, err := carvings.Update(domain.Carving{Id: fresh.Id, Name: fresh.Name, Svg: ""})

	if nil != err {
		t.Fatalf("expected a blanked svg on a never-bolted carving to be accepted, got %v", err)
	}

	if "" != saved.Svg {
		t.Fatalf("expected the svg to be blanked, got %q", saved.Svg)
	}
}

func TestUpdateAcceptsBlankSvgAfterUnbolting(t *testing.T) {
	carvings, _ := newCarvings(t)
	boatSeed := carvingBySpot(t, carvings)[domain.SpotBoat]

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(fresh.Id, domain.SpotBoat); nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	// bolting the v1 seed back onto the spot strips fresh's hold on it: the
	// unbolt path a keeper reaches for before blanking or deleting
	if _, err := carvings.Bolt(boatSeed.Id, domain.SpotBoat); nil != err {
		t.Fatalf("re-bolt failed: %v", err)
	}

	saved, err := carvings.Update(domain.Carving{Id: fresh.Id, Name: fresh.Name, Svg: ""})

	if nil != err {
		t.Fatalf("expected a blanked svg to be accepted after unbolting, got %v", err)
	}

	if "" != saved.Svg {
		t.Fatalf("expected the svg to be blanked, got %q", saved.Svg)
	}
}

func TestUpdateAcceptsNonEmptySvgOnABoltedCarving(t *testing.T) {
	carvings, _ := newCarvings(t)

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(fresh.Id, domain.SpotBoat); nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	// a partial-feeling body that still carries a non-empty svg must go
	// through untouched by the bolted guard
	saved, err := carvings.Update(domain.Carving{Id: fresh.Id, Name: "renamed boat", Svg: fresh.Svg})

	if nil != err {
		t.Fatalf("expected an update that keeps a non-empty svg to be accepted while bolted, got %v", err)
	}

	if "renamed boat" != saved.Name || fresh.Svg != saved.Svg {
		t.Fatalf("update did not apply as expected: %+v", saved)
	}
}

func TestBoltAllowsRebindingASpotOntoABuiltin(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	// the spot is already bolted to this seed; re-bolting it must not trip
	// the builtin guard, since boltedTo is deliberately exempt from it
	saved, err := carvings.Bolt(seed.Id, domain.SpotBoat)

	if nil != err {
		t.Fatalf("expected boltedTo to stay mutable on a builtin, got %v", err)
	}

	found := false
	for _, spot := range saved.BoltedTo {
		if domain.SpotBoat == spot {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected the boat spot to still be bolted, got %+v", saved.BoltedTo)
	}
}

func TestDeleteRejectsBuiltin(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	if err := carvings.Delete(seed.Id); !errors.Is(err, in_port.ErrCarvingBuiltin) {
		t.Fatalf("expected the builtin guard on delete, got %v", err)
	}
}

func TestDeleteAllowsNonBuiltin(t *testing.T) {
	carvings, _ := newCarvings(t)

	scrap, err := carvings.Create(domain.Carving{Name: "scrap", Svg: "<svg></svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if err := carvings.Delete(scrap.Id); nil != err {
		t.Fatalf("deleting a plain carving failed: %v", err)
	}
}

func TestDeleteRejectsBoltedCarving(t *testing.T) {
	carvings, _ := newCarvings(t)

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(fresh.Id, domain.SpotBoat); nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	if err := carvings.Delete(fresh.Id); !errors.Is(err, in_port.ErrCarvingBolted) {
		t.Fatalf("expected the bolted guard on delete, got %v", err)
	}
}

func TestDeleteAllowsBoltedCarvingAfterUnbolting(t *testing.T) {
	carvings, _ := newCarvings(t)
	boatSeed := carvingBySpot(t, carvings)[domain.SpotBoat]

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(fresh.Id, domain.SpotBoat); nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	// bolting the v1 seed back onto the spot strips fresh's hold on it: the
	// unbolt path a keeper reaches for before wanting it gone
	if _, err := carvings.Bolt(boatSeed.Id, domain.SpotBoat); nil != err {
		t.Fatalf("re-bolt failed: %v", err)
	}

	if err := carvings.Delete(fresh.Id); nil != err {
		t.Fatalf("expected delete to be accepted after unbolting, got %v", err)
	}
}

func TestBoltSwapsSpotExclusively(t *testing.T) {
	carvings, _ := newCarvings(t)
	before := carvingBySpot(t, carvings)[domain.SpotBoat]

	fresh, err := carvings.Create(domain.Carving{Name: "new boat", Svg: "<svg>new</svg>"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	saved, err := carvings.Bolt(fresh.Id, domain.SpotBoat)

	if nil != err {
		t.Fatalf("bolt failed: %v", err)
	}

	after := carvingBySpot(t, carvings)

	if after[domain.SpotBoat].Id != saved.Id {
		t.Fatalf("the boat spot should hold the new carving, got %+v", after[domain.SpotBoat])
	}

	// the old holder lost the spot, it was not deleted
	all, _ := carvings.List()
	for _, carving := range all {
		if carving.Id == before.Id {
			for _, spot := range carving.BoltedTo {
				if domain.SpotBoat == spot {
					t.Fatalf("the superseded carving still holds the boat spot: %+v", carving)
				}
			}
		}
	}

	// and every other spot never felt a thing
	if after[domain.SpotBottle].Id != carvingBySpot(t, carvings)[domain.SpotBottle].Id {
		t.Fatalf("bolting the boat spot must not touch the bottle spot")
	}
}

func TestBoltRejectsUnknownSpot(t *testing.T) {
	carvings, _ := newCarvings(t)
	seed := carvingBySpot(t, carvings)[domain.SpotBoat]

	if _, err := carvings.Bolt(seed.Id, "crows-nest"); nil == err {
		t.Fatalf("expected an unknown spot id to be rejected")
	}
}

func TestBoltRejectsEmptySvg(t *testing.T) {
	carvings, _ := newCarvings(t)

	blank, err := carvings.Create(domain.Carving{Name: "blank"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := carvings.Bolt(blank.Id, domain.SpotPaw); nil == err {
		t.Fatalf("expected a carving with no svg to be rejected for bolting")
	}
}

func TestEveryCarvingMutationWritesAKeepersLogLine(t *testing.T) {
	carvings, activity := newCarvings(t)

	created, _ := carvings.Create(domain.Carving{Name: "oilskin", Svg: "<svg></svg>"})
	carvings.Update(domain.Carving{Id: created.Id, Name: "oilskin mk2", Svg: "<svg></svg>"})
	carvings.Bolt(created.Id, domain.SpotPaw)

	// a bolted carving cannot be deleted; free the spot first
	spare, _ := carvings.Create(domain.Carving{Name: "spare", Svg: "<svg></svg>"})
	carvings.Bolt(spare.Id, domain.SpotPaw)

	carvings.Delete(created.Id)

	entries, err := activity.Recent(100)

	if nil != err {
		t.Fatalf("activity read failed: %v", err)
	}

	for _, want := range []string{"created", "edited", "bolted", "deleted", "seeded"} {
		found := false

		for _, entry := range entries {
			if domain.EntityCarving == entry.EntityType && strings.Contains(entry.Message, want) {
				found = true
			}
		}

		if !found {
			t.Fatalf("no %q line reached the keeper's log: %+v", want, entries)
		}
	}
}
