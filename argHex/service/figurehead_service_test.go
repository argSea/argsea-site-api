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

func newFigureheads(t *testing.T) (in_port.FigureheadService, in_port.ActivityService) {
	t.Helper()

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	figureheads := service.NewFigureheadService(out_adapter.NewCatDesignFakeOutAdapter(), activity)

	if err := figureheads.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	return figureheads, activity
}

func publishedByPose(t *testing.T, figureheads in_port.FigureheadService) map[string]domain.CatDesign {
	t.Helper()

	published, err := figureheads.Published()

	if nil != err {
		t.Fatalf("published read failed: %v", err)
	}

	out := map[string]domain.CatDesign{}

	for _, design := range published {
		if _, dup := out[design.Pose]; dup {
			t.Fatalf("two published designs fly the %s pose", design.Pose)
		}

		out[design.Pose] = design
	}

	return out
}

func TestSeedPlantsBothV1CatsPublished(t *testing.T) {
	figureheads, _ := newFigureheads(t)

	current := publishedByPose(t, figureheads)

	if 2 != len(current) {
		t.Fatalf("expected one published design per pose, got %d", len(current))
	}

	for _, pose := range []string{domain.PosePerched, domain.PoseLying} {
		seed, ok := current[pose]

		if !ok || "v1" != seed.Label || !seed.Seed || !seed.Published {
			t.Fatalf("the %s pose should fly the published v1 seed, got %+v", pose, seed)
		}

		if 0 == len(seed.Shapes) || "" == seed.ViewBox {
			t.Fatalf("the %s seed lost its geometry: %+v", pose, seed)
		}
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	figureheads, _ := newFigureheads(t)

	// second boot: the collection is populated, the seed must not touch it
	if err := figureheads.Seed(); nil != err {
		t.Fatalf("re-seed failed: %v", err)
	}

	all, _ := figureheads.List()

	if 2 != len(all) {
		t.Fatalf("re-seeding grew the wardrobe to %d designs", len(all))
	}
}

func TestSeedLeavesANonEmptyCollectionAlone(t *testing.T) {
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	figureheads := service.NewFigureheadService(out_adapter.NewCatDesignFakeOutAdapter(), activity)

	// a lone draft in the collection means a keeper has been here; anything
	// present at all suppresses the seed, not just a full wardrobe
	if _, err := figureheads.Create(domain.CatDesign{Pose: domain.PosePerched, Label: "hand-carved"}); nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if err := figureheads.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	all, _ := figureheads.List()

	if 1 != len(all) {
		t.Fatalf("the seed must not run against a populated collection, got %d designs", len(all))
	}
}

func TestPublishSwapsWithinPoseOnly(t *testing.T) {
	figureheads, _ := newFigureheads(t)
	before := publishedByPose(t, figureheads)

	draft, err := figureheads.Create(domain.CatDesign{Pose: domain.PoseLying, Label: "pirate hat"})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if draft.Published {
		t.Fatalf("a fresh design must arrive as a draft")
	}

	published, err := figureheads.Publish(draft.Id)

	if nil != err || !published.Published {
		t.Fatalf("publish failed: %v %+v", err, published)
	}

	after := publishedByPose(t, figureheads)

	if after[domain.PoseLying].Id != draft.Id {
		t.Fatalf("the lying pose should fly the new design, got %+v", after[domain.PoseLying])
	}

	// the old lying cat was lowered, not deleted
	superseded, _ := figureheads.List()
	for _, design := range superseded {
		if design.Id == before[domain.PoseLying].Id && design.Published {
			t.Fatalf("the superseded lying design is still published")
		}
	}

	// and the perched pose never felt a thing
	if after[domain.PosePerched].Id != before[domain.PosePerched].Id {
		t.Fatalf("publishing a lying design must not touch the perched cat")
	}
}

func TestPublishedPicksTheNewerInTheCrashWindow(t *testing.T) {
	repo := out_adapter.NewCatDesignFakeOutAdapter()
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	figureheads := service.NewFigureheadService(repo, activity)

	// the crash window inside Publish: the new design is hoisted but the old
	// one not yet lowered, so two published designs fly the same pose. The
	// public read must settle on the newer stamp, deterministically.
	olderID, _ := repo.Add(domain.CatDesign{Pose: domain.PoseLying, Label: "old coat", Published: true, UpdatedAt: "2026-01-01T00:00:00.000000000Z"})
	newerID, _ := repo.Add(domain.CatDesign{Pose: domain.PoseLying, Label: "new coat", Published: true, UpdatedAt: "2026-02-01T00:00:00.000000000Z"})

	current := publishedByPose(t, figureheads)

	if 1 != len(current) || current[domain.PoseLying].Id != newerID {
		t.Fatalf("expected the newer design %s on the bow, got %+v (older was %s)", newerID, current, olderID)
	}
}

func TestDeleteGuardsPublishedAndSeeds(t *testing.T) {
	figureheads, _ := newFigureheads(t)
	current := publishedByPose(t, figureheads)

	// the published seed trips the seed guard first; permanent is permanent
	if err := figureheads.Delete(current[domain.PosePerched].Id); !errors.Is(err, in_port.ErrDesignSeeded) {
		t.Fatalf("expected the seed guard, got %v", err)
	}

	// a published non-seed design is guarded too
	draft, _ := figureheads.Create(domain.CatDesign{Pose: domain.PosePerched, Label: "sou'wester"})
	figureheads.Publish(draft.Id)

	if err := figureheads.Delete(draft.Id); !errors.Is(err, in_port.ErrDesignPublished) {
		t.Fatalf("expected the published guard, got %v", err)
	}

	// a superseded seed stays undeletable; that is the whole point of seeding
	supersededSeed := current[domain.PosePerched].Id

	if err := figureheads.Delete(supersededSeed); !errors.Is(err, in_port.ErrDesignSeeded) {
		t.Fatalf("expected the seed guard on a superseded seed, got %v", err)
	}

	// a plain unpublished draft deletes fine
	scrap, _ := figureheads.Create(domain.CatDesign{Pose: domain.PoseLying, Label: "scrap"})

	if err := figureheads.Delete(scrap.Id); nil != err {
		t.Fatalf("deleting a plain draft failed: %v", err)
	}
}

func TestUpdateLeavesLifecycleAndPoseAlone(t *testing.T) {
	figureheads, _ := newFigureheads(t)

	draft, _ := figureheads.Create(domain.CatDesign{Pose: domain.PoseLying, Label: "draft coat"})
	figureheads.Publish(draft.Id)

	// an update smuggling published:false and a pose flip changes neither
	saved, err := figureheads.Update(domain.CatDesign{Id: draft.Id, Pose: domain.PosePerched, Label: "new coat", Published: false})

	if nil != err {
		t.Fatalf("update failed: %v", err)
	}

	if "new coat" != saved.Label || !saved.Published || domain.PoseLying != saved.Pose {
		t.Fatalf("update must keep lifecycle and stance, got %+v", saved)
	}
}

func TestSeedsAreImmutable(t *testing.T) {
	figureheads, _ := newFigureheads(t)
	current := publishedByPose(t, figureheads)

	_, err := figureheads.Update(domain.CatDesign{Id: current[domain.PoseLying].Id, Label: "defaced v1"})

	if !errors.Is(err, in_port.ErrDesignSeeded) {
		t.Fatalf("expected the seed guard on an update, got %v", err)
	}
}

func TestCreateRejectsUnknownPoseAndShapeType(t *testing.T) {
	figureheads, _ := newFigureheads(t)

	if _, err := figureheads.Create(domain.CatDesign{Pose: "standing", Label: "impossible"}); nil == err {
		t.Fatalf("expected an unknown pose to be rejected")
	}

	bad := domain.CatDesign{
		Pose:   domain.PosePerched,
		Label:  "smuggler",
		Shapes: []domain.Shape{{Type: "script", D: "alert(1)"}},
	}

	if _, err := figureheads.Create(bad); nil == err {
		t.Fatalf("expected an unknown shape type to be rejected")
	}
}

func TestEveryMutationWritesAKeepersLogLine(t *testing.T) {
	figureheads, activity := newFigureheads(t)

	draft, _ := figureheads.Create(domain.CatDesign{Pose: domain.PoseLying, Label: "oilskin"})
	figureheads.Update(domain.CatDesign{Id: draft.Id, Label: "oilskin mk2"})
	figureheads.Publish(draft.Id)

	scrap, _ := figureheads.Create(domain.CatDesign{Pose: domain.PosePerched, Label: "scrap"})
	figureheads.Delete(scrap.Id)

	entries, err := activity.Recent(100)

	if nil != err {
		t.Fatalf("activity read failed: %v", err)
	}

	for _, want := range []string{"created", "edited", "published", "deleted", "seeded"} {
		found := false

		for _, entry := range entries {
			if domain.EntityFigurehead == entry.EntityType && strings.Contains(entry.Message, want) {
				found = true
			}
		}

		if !found {
			t.Fatalf("no %q line reached the keeper's log: %+v", want, entries)
		}
	}
}
