package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newRollbackHarness wires the lantern over a scriptable fake release store,
// so rollback can be exercised without a filesystem.
func newRollbackHarness(gate chan struct{}, previous string) (in_port.LanternService, *out_adapter.LanternFakeReleaseStore, in_port.ActivityService) {
	releases := &out_adapter.LanternFakeReleaseStore{Prev: previous}
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		&out_adapter.LanternFakeRunner{Gate: gate},
		releases,
		&out_adapter.LanternFakeStateRepo{},
		activity,
	)

	return lantern, releases, activity
}

func TestRollbackSwapsToThePreviousGeneration(t *testing.T) {
	lantern, releases, activity := newRollbackHarness(nil, "gen-older")

	status, err := lantern.Rollback()

	if nil != err {
		t.Fatalf("rollback failed: %v", err)
	}

	if 1 != len(releases.Swapped) || "gen-older" != releases.Swapped[0] {
		t.Fatalf("expected the live link swapped to gen-older, got %v", releases.Swapped)
	}

	// no rebuild: nothing staged, status untouched by the swap
	if 0 != len(releases.Staged) {
		t.Fatalf("rollback must not stage anything, got %v", releases.Staged)
	}

	if "" != status.StartedAt {
		t.Fatalf("rollback must not start a hoist, got %+v", status)
	}

	entries, _ := activity.Recent(10)

	if 1 != len(entries) || !strings.Contains(entries[0].Message, "rolled back") {
		t.Fatalf("expected a rolled-back entry in the ship's log, got %+v", entries)
	}
}

func TestRollbackWithNoPreviousBuildIsRefused(t *testing.T) {
	lantern, releases, _ := newRollbackHarness(nil, "")

	if _, err := lantern.Rollback(); !errors.Is(err, in_port.ErrNoPreviousBuild) {
		t.Fatalf("expected ErrNoPreviousBuild, got %v", err)
	}

	if 0 != len(releases.Swapped) {
		t.Fatalf("a refused rollback must not move the live link, got %v", releases.Swapped)
	}
}

func TestRollbackDuringHoistIsRefused(t *testing.T) {
	gate := make(chan struct{})
	lantern, releases, _ := newRollbackHarness(gate, "gen-older")

	defer close(gate)

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("hoist failed to start: %v", err)
	}

	// the build is blocked on the gate — the rollback must bounce off it
	if _, err := lantern.Rollback(); !errors.Is(err, in_port.ErrHoistAlreadyRunning) {
		t.Fatalf("expected ErrHoistAlreadyRunning, got %v", err)
	}

	if 0 != len(releases.Swapped) {
		t.Fatalf("a refused rollback must not move the live link, got %v", releases.Swapped)
	}
}
