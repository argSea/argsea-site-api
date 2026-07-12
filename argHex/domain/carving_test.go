package domain_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// A non-builtin carving is builtin:false; that explicit false must survive a
// Replace-based Set the same way figurehead's published/seed flags do, or a
// created carving could come back reading as a permanent seed.

func TestBuiltinFalseSurvivesBsonRoundTrip(t *testing.T) {
	carving := domain.Carving{
		Name:    "hand-carved",
		Svg:     "<svg></svg>",
		Builtin: false,
	}

	raw, err := bson.Marshal(carving)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	if false != doc["builtin"] {
		t.Fatalf("the carving is not builtin but the doc says %v: omitempty ate the false", doc["builtin"])
	}

	var back domain.Carving
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to Carving failed: %v", err)
	}
	if back.Builtin {
		t.Fatalf("builtin did not round-trip: %v", back.Builtin)
	}
}

func TestBoltedToRoundTripsByteForByte(t *testing.T) {
	carving := domain.Carving{
		Name:     "The lighthouse",
		Svg:      "<svg></svg>",
		BoltedTo: []string{domain.SpotLighthouseLogo, domain.SpotBoat},
	}

	raw, err := bson.Marshal(carving)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var back domain.Carving
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if 2 != len(back.BoltedTo) || domain.SpotLighthouseLogo != back.BoltedTo[0] || domain.SpotBoat != back.BoltedTo[1] {
		t.Fatalf("boltedTo did not round-trip: %+v", back.BoltedTo)
	}
}
