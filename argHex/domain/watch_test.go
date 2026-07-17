package domain_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// A cleared watch is a valid state, not a missing one: the keeper writes an
// empty record and the site collapses the section. The empty Letter has to
// survive the Replace-based Save without the doc growing blocks it never had.

func TestAClearedWatchRoundTrips(t *testing.T) {
	raw, err := bson.Marshal(domain.Watch{KeptAt: "2026-07-15T00:00:00.000000000Z"})

	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}
	for _, key := range []string{"letter", "rotation", "bearings", "postcardMediaId", "postcard2MediaId", "quips"} {
		if _, present := doc[key]; present {
			t.Fatalf("a cleared watch grew a %q block it never had: %v", key, doc)
		}
	}

	var back domain.Watch
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to Watch failed: %v", err)
	}
	if "" != back.Letter || nil != back.Bearings {
		t.Fatalf("a cleared watch should load empty, got letter=%q bearings=%+v", back.Letter, back.Bearings)
	}
	if "2026-07-15T00:00:00.000000000Z" != back.KeptAt {
		t.Fatalf("keptAt did not round-trip: %q", back.KeptAt)
	}
}

// A "none" bearing carries no target on purpose; its empty TargetId must
// persist rather than vanish, or a reload would leave the strip unable to
// tell an unlinked line from a broken one.

func TestBearingsKeepTheirKeys(t *testing.T) {
	watch := domain.Watch{
		Letter: "All quiet at the light.",
		Bearings: []domain.WatchBearing{
			{Verb: "wrangling", Kind: "light", TargetId: "fastnet", Name: "Fastnet Rock"},
			{Verb: "tinkering", Kind: "none", TargetId: "", Name: "the workbench"},
		},
	}

	raw, err := bson.Marshal(watch)

	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	bearings, ok := doc["bearings"].(bson.A)
	if !ok || 2 != len(bearings) {
		t.Fatalf("bearings block missing from the persisted doc: %v", doc)
	}

	unlinked, ok := bearings[1].(bson.M)
	if !ok {
		t.Fatalf("the none bearing did not persist as a doc: %v", bearings[1])
	}

	targetId, present := unlinked["targetId"]
	if !present {
		t.Fatalf("the none bearing's empty targetId went missing: %v", unlinked)
	}
	if "" != targetId {
		t.Fatalf("the none bearing grew a target: %v", unlinked)
	}

	var back domain.Watch
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to Watch failed: %v", err)
	}
	if 2 != len(back.Bearings) || "none" != back.Bearings[1].Kind || "" != back.Bearings[1].TargetId {
		t.Fatalf("bearings did not round-trip: %+v", back.Bearings)
	}
}
