package domain_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// The eggs contract is absent-means-on, so an explicit false is the only way
// to switch one off. These tests guard the trap: an omitempty on the inner
// bools would drop the false during the Replace-based Save and the egg would
// come back on by itself.

func TestEggsFalseSurvivesBsonRoundTrip(t *testing.T) {
	flags := domain.SiteCopy{
		Eggs: &domain.Eggs{Bottle: false, Cat: true, Lights: true},
	}

	raw, err := bson.Marshal(flags)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	eggs, ok := doc["eggs"].(bson.M)
	if !ok {
		t.Fatalf("eggs block missing from the persisted doc: %v", doc)
	}
	if false != eggs["bottle"] {
		t.Fatalf("the bottle was switched off but the doc says %v: omitempty ate the false", eggs["bottle"])
	}

	var back domain.SiteCopy
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to SiteCopy failed: %v", err)
	}
	if nil == back.Eggs || back.Eggs.Bottle || !back.Eggs.Cat {
		t.Fatalf("eggs did not round-trip: %+v", back.Eggs)
	}
}

// The cat catalog moved from a fixed struct to open maps, but the false-off
// contract came with it; a spot switched off has to persist its false. A map
// value dodges the omitempty-on-bool trap, but this guards it end to end.

func TestCatMapFalseSurvivesBsonRoundTrip(t *testing.T) {
	flags := domain.SiteCopy{
		CatPages: map[string]bool{"hello": true, "notes": false},
		CatSpots: map[string]bool{"hello.header": true, "notes.footer": false},
	}

	raw, err := bson.Marshal(flags)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	pages, ok := doc["catPages"].(bson.M)
	if !ok {
		t.Fatalf("catPages block missing from the persisted doc: %v", doc)
	}
	if false != pages["notes"] {
		t.Fatalf("the cat was banned from notes but the doc says %v: the false went missing", pages["notes"])
	}

	spots, ok := doc["catSpots"].(bson.M)
	if !ok {
		t.Fatalf("catSpots block missing from the persisted doc: %v", doc)
	}
	if false != spots["notes.footer"] {
		t.Fatalf("the cat was banned from notes.footer but the doc says %v: the false went missing", spots["notes.footer"])
	}

	var back domain.SiteCopy
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to SiteCopy failed: %v", err)
	}
	if back.CatPages["notes"] || !back.CatPages["hello"] {
		t.Fatalf("catPages did not round-trip: %+v", back.CatPages)
	}
	if back.CatSpots["notes.footer"] || !back.CatSpots["hello.header"] {
		t.Fatalf("catSpots did not round-trip: %+v", back.CatSpots)
	}
}

func TestLegacyDocRoundTripsWithoutEggBlocks(t *testing.T) {
	// a pre-eggs doc has none of the new fields; it must save without growing
	// them and load with nil pointers so consumers can default to everything-on
	raw, err := bson.Marshal(domain.SiteCopy{QuipHello: "ahoy"})
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}
	for _, key := range []string{"eggs", "catPages", "catSpots", "bottleProverbs", "lighthouses", "wallGhost"} {
		if _, present := doc[key]; present {
			t.Fatalf("legacy doc grew a %q block it never had: %v", key, doc)
		}
	}

	var back domain.SiteCopy
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to SiteCopy failed: %v", err)
	}
	if nil != back.Eggs || nil != back.CatPages || nil != back.CatSpots || nil != back.WallGhost {
		t.Fatalf("legacy doc should load with nil egg config, got eggs=%+v catPages=%+v catSpots=%+v wallGhost=%+v", back.Eggs, back.CatPages, back.CatSpots, back.WallGhost)
	}
}

// The wall ghost is off by default in the sense that a disabled placard has
// to stay disabled through a Replace-based Save, so its Enabled bool guards
// the same omitempty trap as Eggs.

func TestWallGhostSurvivesBsonRoundTrip(t *testing.T) {
	flags := domain.SiteCopy{
		WallGhost: &domain.WallGhost{X: 12.5, Y: 80, Rotation: -3, Enabled: false},
	}

	raw, err := bson.Marshal(flags)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	ghost, ok := doc["wallGhost"].(bson.M)
	if !ok {
		t.Fatalf("wallGhost block missing from the persisted doc: %v", doc)
	}
	if false != ghost["enabled"] {
		t.Fatalf("the ghost was switched off but the doc says %v: omitempty ate the false", ghost["enabled"])
	}

	var back domain.SiteCopy
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to SiteCopy failed: %v", err)
	}
	if nil == back.WallGhost {
		t.Fatalf("wallGhost did not round-trip: got nil")
	}
	if 12.5 != back.WallGhost.X || 80 != back.WallGhost.Y || -3 != back.WallGhost.Rotation || back.WallGhost.Enabled {
		t.Fatalf("expected wall ghost {12.5 80 -3 false}, got %+v", back.WallGhost)
	}
}

func TestLighthousesKeepTheirKeys(t *testing.T) {
	flags := domain.SiteCopy{
		BottleProverbs: []string{"a smooth sea never made a skilled sailor"},
		Lighthouses: []domain.Lighthouse{
			{Name: "Fastnet Rock", Pos: "51°23′N 9°36′W", Line: "Ireland's teardrop."},
		},
	}

	raw, err := bson.Marshal(flags)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var back domain.SiteCopy
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if 1 != len(back.BottleProverbs) || "a smooth sea never made a skilled sailor" != back.BottleProverbs[0] {
		t.Fatalf("proverbs did not round-trip: %v", back.BottleProverbs)
	}
	if 1 != len(back.Lighthouses) || "Fastnet Rock" != back.Lighthouses[0].Name || "51°23′N 9°36′W" != back.Lighthouses[0].Pos {
		t.Fatalf("light list did not round-trip: %+v", back.Lighthouses)
	}
}
