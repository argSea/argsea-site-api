package out_adapter

import (
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// The drawer backfill is a mongo upsert concern with no live mongo in this
// suite, so the tests drive backfillUpsert's real filter+update pair directly:
// the same seam hobby_migration_test.go tests its passes through.

func staleFlare(id string) domain.Sighting {
	return domain.Sighting{
		Id:      id,
		Kind:    domain.SightingFlare,
		Day:     "2026-06-01",
		Path:    "/hobbies",
		Subject: "piano",
		Port:    domain.PortDirect,
		Visitor: "fv1",
		At:      time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestBackfillUpsertKeysTheFilterOnTheRowsOwnId(t *testing.T) {
	id := primitive.NewObjectID()
	filter, _, ok := backfillUpsert(staleFlare(id.Hex()))

	if !ok {
		t.Fatalf("a well-formed id must be accepted")
	}

	if id != filter["_id"] {
		t.Fatalf("expected the filter keyed on the row's own id %v, got %+v", id, filter)
	}

	if 1 != len(filter) {
		t.Fatalf("expected the filter to carry only _id, got %+v", filter)
	}
}

func TestBackfillUpsertSetsEveryFieldButId(t *testing.T) {
	flare := staleFlare(primitive.NewObjectID().Hex())
	_, update, ok := backfillUpsert(flare)

	if !ok {
		t.Fatalf("a well-formed id must be accepted")
	}

	if 1 != len(update) || "$set" != update[0].Key {
		t.Fatalf("expected a single $set update, got %+v", update)
	}

	set, isM := update[0].Value.(bson.M)

	if !isM {
		t.Fatalf("expected the $set value to be a bson.M, got %T", update[0].Value)
	}

	if flare.Kind != set["kind"] || flare.Day != set["day"] || flare.Path != set["path"] {
		t.Fatalf("kind/day/path did not carry into the $set: %+v", set)
	}

	if flare.Subject != set["subject"] || flare.Port != set["port"] || flare.Visitor != set["visitor"] {
		t.Fatalf("subject/port/visitor did not carry into the $set: %+v", set)
	}

	if flare.At != set["at"] {
		t.Fatalf("at did not carry into the $set: %+v", set)
	}

	// _id is immutable and already rides the filter; setting it again is at
	// best redundant and at worst an update mongo rejects
	if _, present := set["_id"]; present {
		t.Fatalf("the $set must not carry _id, got %+v", set)
	}
}

func TestBackfillUpsertSkipsAMalformedId(t *testing.T) {
	for _, badId := range []string{"", "not-an-object-id", "sighting-4"} {
		filter, update, ok := backfillUpsert(staleFlare(badId))

		if ok {
			t.Fatalf("a malformed id %q must be skipped, not upserted", badId)
		}

		if nil != filter || nil != update {
			t.Fatalf("a skipped row must not hand back a filter or update, got %+v / %+v", filter, update)
		}
	}
}
