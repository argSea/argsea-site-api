package domain_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"go.mongodb.org/mongo-driver/bson"
)

// A superseded design is published:false — that explicit false is the whole
// one-published-per-pose bookkeeping. These tests guard the known trap: an
// omitempty on the bool would drop the false during the Replace-based Set and
// two cats would fly the same pose.

func TestPublishedFalseSurvivesBsonRoundTrip(t *testing.T) {
	design := domain.CatDesign{
		Pose:      domain.PoseLying,
		Label:     "superseded",
		Published: false,
		Seed:      false,
	}

	raw, err := bson.Marshal(design)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); nil != err {
		t.Fatalf("unmarshal to doc failed: %v", err)
	}

	if false != doc["published"] {
		t.Fatalf("the design was unpublished but the doc says %v — omitempty ate the false", doc["published"])
	}

	if false != doc["seed"] {
		t.Fatalf("the design is no seed but the doc says %v — omitempty ate the false", doc["seed"])
	}

	var back domain.CatDesign
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal to CatDesign failed: %v", err)
	}
	if back.Published || back.Seed {
		t.Fatalf("flags did not round-trip: published=%v seed=%v", back.Published, back.Seed)
	}
}

func TestShapeGeometryRoundTripsByteForByte(t *testing.T) {
	// the seed fidelity requirement: a shape's path data must come back exactly
	// as it went in, or the shipped cats stop rendering identical
	tail := "M45 55 C57 52 61 62 56 70 C54.5 72.5 51 72.5 50 70 C52.5 64.5 50 60 43 60 Z"

	design := domain.CatDesign{
		Pose:    domain.PosePerched,
		ViewBox: "0 0 64 74",
		Shapes: []domain.Shape{
			{Id: "tail", Type: "path", D: tail, Fill: "#232a4d", Stroke: "#93a0e8", StrokeWidth: 1.4, Linejoin: "round", Role: "tail", Origin: []float64{45, 56}},
			{Id: "eye-left", Type: "ellipse", Cx: 25.9, Cy: 30.8, Rx: 1.9, Ry: 1.9, Fill: "#f0d9a8", Role: "eyes", Origin: []float64{30, 31}},
		},
	}

	raw, err := bson.Marshal(design)
	if nil != err {
		t.Fatalf("marshal failed: %v", err)
	}

	var back domain.CatDesign
	if err := bson.Unmarshal(raw, &back); nil != err {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if 2 != len(back.Shapes) || tail != back.Shapes[0].D {
		t.Fatalf("path data did not round-trip byte-for-byte: %+v", back.Shapes)
	}

	eye := back.Shapes[1]
	if 25.9 != eye.Cx || 30.8 != eye.Cy || 1.9 != eye.Rx || 1.9 != eye.Ry {
		t.Fatalf("ellipse geometry did not round-trip: %+v", eye)
	}
	if "eyes" != eye.Role || 2 != len(eye.Origin) || 30 != eye.Origin[0] || 31 != eye.Origin[1] {
		t.Fatalf("role/origin did not round-trip: %+v", eye)
	}
}
