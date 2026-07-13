package out_adapter

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// The boot migration is a mongo $rename/$set concern with no live mongo in this
// suite, so the tests drive the migration's real update documents over in-memory
// raw docs, modeling the operators it uses ($exists, $ne, bare equality,
// $rename, $set) the same way the fakes mirror mongo elsewhere.

func applyMigration(docs []bson.M) int {
	modified := 0

	for _, pass := range hobbyMigrationPasses() {
		for _, doc := range docs {
			if !matchesFilter(doc, pass.filter) {
				continue
			}

			applyUpdate(doc, pass.update)
			modified++
		}
	}

	return modified
}

func matchesFilter(doc bson.M, filter bson.M) bool {
	for field, cond := range filter {
		val, present := doc[field]

		clause, isClause := cond.(bson.M)

		if !isClause {
			if !present || val != cond {
				return false
			}

			continue
		}

		for op, operand := range clause {
			switch op {
			case "$exists":
				if present != operand.(bool) {
					return false
				}
			case "$ne":
				if present && val == operand {
					return false
				}
			default:
				return false
			}
		}
	}

	return true
}

func applyUpdate(doc bson.M, update bson.M) {
	if rename, ok := update["$rename"].(bson.M); ok {
		for src, dst := range rename {
			if val, present := doc[src]; present {
				doc[dst.(string)] = val
				delete(doc, src)
			}
		}
	}

	if set, ok := update["$set"].(bson.M); ok {
		for key, val := range set {
			doc[key] = val
		}
	}
}

func oldActiveHobby() bson.M {
	return bson.M{
		"name":        "Piano",
		"active":      true,
		"service":     "2023 · 2024",
		"log":         "still playing",
		"cause":       "",
		"found":       "a cheap upright",
		"return":      "maybe",
		"lastLog":     "last week",
		"disposition": "keen",
		"marker":      "buoy",
		"char":        "Fl W 3s",
		"wear":        0.4,
		"order":       1,
	}
}

func TestMigrationLiftsAnActiveDocToTheShipsLogShape(t *testing.T) {
	doc := oldActiveHobby()

	if 1 != applyMigration([]bson.M{doc}) {
		t.Fatalf("expected the one old-shape doc migrated")
	}

	if domain.StateMoored != doc["state"] {
		t.Fatalf("an active doc must derive state moored, got %v", doc["state"])
	}

	// the four prose fields are renamed with their values preserved
	if "still playing" != doc["bearing"] || "a cheap upright" != doc["floats"] || "maybe" != doc["odds"] {
		t.Fatalf("renamed prose lost its value: %+v", doc)
	}

	// the source names are gone, moved by $rename
	for _, gone := range []string{"log", "found", "return", "cause"} {
		if _, present := doc[gone]; present {
			t.Fatalf("the source field %q must be moved by the rename", gone)
		}
	}

	// coord/from land as null, seasons defaults to empty
	if coord, ok := doc["coord"]; !ok || nil != coord {
		t.Fatalf("coord must be set to null, got %v (present %v)", coord, ok)
	}

	if from, ok := doc["from"]; !ok || nil != from {
		t.Fatalf("from must be set to null, got %v (present %v)", from, ok)
	}

	if "" != doc["seasons"] {
		t.Fatalf("seasons must default to empty, got %v", doc["seasons"])
	}

	// name/service/lastLog/order ride through unchanged
	if "Piano" != doc["name"] || "2023 · 2024" != doc["service"] || "last week" != doc["lastLog"] || 1 != doc["order"] {
		t.Fatalf("carried fields changed: %+v", doc)
	}

	// legacy fields are left as dead data, never $unset
	if true != doc["active"] || "keen" != doc["disposition"] || "buoy" != doc["marker"] || "Fl W 3s" != doc["char"] || 0.4 != doc["wear"] {
		t.Fatalf("legacy fields must survive untouched: %+v", doc)
	}
}

func TestMigrationDerivesAdriftForInactive(t *testing.T) {
	inactive := bson.M{"name": "Kite", "active": false, "log": "drifted off", "order": 2}
	noFlag := bson.M{"name": "Origami", "order": 3}

	if 2 != applyMigration([]bson.M{inactive, noFlag}) {
		t.Fatalf("expected both docs migrated")
	}

	if domain.StateAdrift != inactive["state"] {
		t.Fatalf("an inactive doc must derive state adrift, got %v", inactive["state"])
	}

	// a doc with no active flag at all still lands adrift, never lost
	if domain.StateAdrift != noFlag["state"] {
		t.Fatalf("a doc missing active must default to adrift, got %v", noFlag["state"])
	}

	if "drifted off" != inactive["bearing"] {
		t.Fatalf("the inactive doc's log must rename to bearing, got %v", inactive["bearing"])
	}
}

func TestMigrationLeavesMissingOptionalFieldsAlone(t *testing.T) {
	// an old doc that never carried a found/return/cause: the rename touches only
	// the fields it has, and never invents floats/odds/offCourse
	sparse := bson.M{"name": "Chess", "active": true, "log": "won a club game", "order": 4}

	if 1 != applyMigration([]bson.M{sparse}) {
		t.Fatalf("expected the sparse doc migrated")
	}

	if "won a club game" != sparse["bearing"] {
		t.Fatalf("the present prose field must still rename, got %v", sparse["bearing"])
	}

	for _, absent := range []string{"floats", "odds", "offCourse"} {
		if _, present := sparse[absent]; present {
			t.Fatalf("a missing source must not create %q", absent)
		}
	}

	if domain.StateMoored != sparse["state"] {
		t.Fatalf("state must still derive, got %v", sparse["state"])
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	doc := oldActiveHobby()

	applyMigration([]bson.M{doc})
	bearing := doc["bearing"]

	// a second boot touches nothing: the docs already carry state
	if 0 != applyMigration([]bson.M{doc}) {
		t.Fatalf("a second migration run must modify nothing")
	}

	if bearing != doc["bearing"] || domain.StateMoored != doc["state"] {
		t.Fatalf("the re-run must leave the migrated doc unchanged: %+v", doc)
	}
}

func TestMigrationSkipsDocsAlreadyCarryingState(t *testing.T) {
	// a doc already in the ship's-log shape must not be re-derived or renamed
	fresh := bson.M{"name": "Sailing", "state": domain.StatePort, "bearing": "made harbor", "coord": bson.M{"lat": 1.0, "lon": 2.0}, "order": 5}

	if 0 != applyMigration([]bson.M{fresh}) {
		t.Fatalf("a doc with state must be skipped")
	}

	if domain.StatePort != fresh["state"] || "made harbor" != fresh["bearing"] {
		t.Fatalf("a stated doc must ride through untouched: %+v", fresh)
	}
}

func TestMigrationCountsOnlyOldShapeDocs(t *testing.T) {
	docs := []bson.M{
		oldActiveHobby(),
		{"name": "Kite", "active": false, "order": 2},
		{"name": "Sailing", "state": domain.StateInkspill, "order": 3},
	}

	if 2 != applyMigration(docs) {
		t.Fatalf("expected only the two old-shape docs counted")
	}
}
