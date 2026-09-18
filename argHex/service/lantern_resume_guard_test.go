package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newGuardHarness wires the lantern over fakes with the given resume shelf, so
// the guard can be exercised without a filesystem or a build.
func newGuardHarness(resumes in_port.ResumeService) (in_port.LanternService, *out_adapter.LanternFakeReleaseStore) {
	releases := &out_adapter.LanternFakeReleaseStore{}

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		&out_adapter.LanternFakeRunner{},
		releases,
		&out_adapter.LanternFakeStateRepo{},
		resumes,
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	return lantern, releases
}

func TestHoistRefusesWhileNoResumeIsPublished(t *testing.T) {
	lantern, releases := newGuardHarness(emptyShelf())

	status, err := lantern.Hoist()

	if !errors.Is(err, in_port.ErrNoPublishedResume) {
		t.Fatalf("expected ErrNoPublishedResume, got %v", err)
	}

	if domain.LanternIdle != status.State {
		t.Fatalf("a refused hoist must leave the lantern idle, got %q", status.State)
	}

	// the refusal has to land before the build, not after one
	if 0 != len(releases.Staged) {
		t.Fatalf("a refused hoist must stage nothing, staged %+v", releases.Staged)
	}
}

func TestHoistRunsOnceAResumeIsPublished(t *testing.T) {
	lantern, _ := newGuardHarness(publishedShelf())

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("a hoist with a published resume must start: %v", err)
	}

	waitTerminal(t, lantern)
}
