package service_test

import (
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newDoodles(t *testing.T) (in_port.DoodleService, in_port.ActivityService) {
	t.Helper()

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	doodles := service.NewDoodleService(out_adapter.NewDoodleFakeOutAdapter(), activity)

	return doodles, activity
}

func TestCreateDoodleStampsIdAndTimestamps(t *testing.T) {
	doodles, _ := newDoodles(t)

	saved, err := doodles.Create(domain.Doodle{Name: "anchor", ViewBox: "0 0 10 10"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if "" == saved.Id || "" == saved.CreatedAt || "" == saved.UpdatedAt {
		t.Fatalf("create must stamp id and timestamps: %+v", saved)
	}
}

func TestCreateRejectsUnknownShapeType(t *testing.T) {
	doodles, _ := newDoodles(t)

	bad := domain.Doodle{
		Name:   "smuggler",
		Shapes: []domain.Shape{{Type: "script", D: "alert(1)"}},
	}

	if _, err := doodles.Create(bad); nil == err {
		t.Fatalf("expected an unknown shape type to be rejected")
	}
}

func TestUpdateIsAFullReplacePreservingCreatedAt(t *testing.T) {
	doodles, _ := newDoodles(t)

	draft, _ := doodles.Create(domain.Doodle{Name: "anchor"})

	saved, err := doodles.Update(domain.Doodle{Id: draft.Id, Name: "anchor mk2", ViewBox: "0 0 20 20"})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if "anchor mk2" != saved.Name || draft.CreatedAt != saved.CreatedAt {
		t.Fatalf("update must replace fields and keep createdAt: %+v", saved)
	}
}

func TestUpdateMissingDoodleErrors(t *testing.T) {
	doodles, _ := newDoodles(t)

	if _, err := doodles.Update(domain.Doodle{Id: "nope", Name: "ghost"}); nil == err {
		t.Fatalf("expected an update against a missing doodle to fail")
	}
}

func TestDeleteMissingDoodleErrors(t *testing.T) {
	doodles, _ := newDoodles(t)

	if err := doodles.Delete("nope"); nil == err {
		t.Fatalf("expected deleting a missing doodle to fail")
	}
}

func TestEveryDoodleMutationWritesAShipsLogLine(t *testing.T) {
	doodles, activity := newDoodles(t)

	draft, _ := doodles.Create(domain.Doodle{Name: "anchor"})
	doodles.Update(domain.Doodle{Id: draft.Id, Name: "anchor mk2"})
	doodles.Delete(draft.Id)

	entries, err := activity.Recent(100)

	if nil != err {
		t.Fatalf("activity read failed: %v", err)
	}

	for _, want := range []string{"created", "edited", "deleted"} {
		found := false

		for _, entry := range entries {
			if domain.EntityDoodle == entry.EntityType && strings.Contains(entry.Message, want) {
				found = true
			}
		}

		if !found {
			t.Fatalf("no %q line reached the keeper's log: %+v", want, entries)
		}
	}
}
