package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// newLanternHarness lays out a temp release area (site dir with a built dist,
// releases dir, live link path) and wires the service with the REAL exec
// runner and REAL filesystem release store, so a stub argv command drives the
// whole pipeline end-to-end without Mongo or a Node toolchain.
func newLanternHarness(t *testing.T, buildCmd []string, keep int) (in_port.LanternService, *out_adapter.LanternFakeStateRepo, in_port.ActivityService, string, string) {
	t.Helper()

	root := t.TempDir()
	siteDir := filepath.Join(root, "site")
	releasesDir := filepath.Join(root, "releases")
	liveLink := filepath.Join(root, "live")

	seedDist(t, siteDir)

	state := &out_adapter.LanternFakeStateRepo{}
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	lantern := service.NewLanternService(
		service.LanternConfig{
			SiteDir:  siteDir,
			BuildCmd: buildCmd,
			DistDir:  "dist",
			Keep:     keep,
			Timeout:  5 * time.Second,
		},
		out_adapter.NewLanternExecAdapter(),
		out_adapter.NewLanternFSReleaseAdapter(releasesDir, liveLink),
		state,
		activity,
	)

	return lantern, state, activity, releasesDir, liveLink
}

// seedDist creates site/dist/index.html: the "build output" the stub command
// pretends to have produced.
func seedDist(t *testing.T, siteDir string) {
	t.Helper()

	dist := filepath.Join(siteDir, "dist")

	if err := os.MkdirAll(dist, 0755); nil != err {
		t.Fatalf("could not seed dist: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte("<h1>ahoy</h1>"), 0644); nil != err {
		t.Fatalf("could not seed index.html: %v", err)
	}
}

// waitTerminal polls Status until the hoist reaches a terminal state.
func waitTerminal(t *testing.T, lantern in_port.LanternService) domain.LanternStatus {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		status := lantern.Status()

		if domain.LanternSucceeded == status.State || domain.LanternFailed == status.State {
			return status
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("hoist never reached a terminal state, stuck at %q", lantern.Status().State)

	return domain.LanternStatus{}
}

func TestHoistSuccessSwapsLiveLink(t *testing.T) {
	lantern, state, activity, _, liveLink := newLanternHarness(t, []string{"true"}, 2)

	status, err := lantern.Hoist()

	if nil != err {
		t.Fatalf("hoist failed to start: %v", err)
	}

	if domain.LanternBuilding != status.State || "" == status.StartedAt {
		t.Fatalf("expected an immediate building status, got %+v", status)
	}

	final := waitTerminal(t, lantern)

	if domain.LanternSucceeded != final.State {
		t.Fatalf("expected succeeded, got %q (output: %s)", final.State, final.Output)
	}

	if "" == final.FinishedAt || final.LastHoistedAt != final.FinishedAt {
		t.Fatalf("expected finishedAt and lastHoistedAt stamped together, got %+v", final)
	}

	if state.Stamp != final.LastHoistedAt {
		t.Fatalf("lastHoistedAt was not persisted (repo has %q)", state.Stamp)
	}

	// the live link must point at a generation containing the built site
	target, readErr := os.Readlink(liveLink)

	if nil != readErr {
		t.Fatalf("live link missing after a successful hoist: %v", readErr)
	}

	if _, statErr := os.Stat(filepath.Join(target, "index.html")); nil != statErr {
		t.Fatalf("live generation does not contain the built site: %v", statErr)
	}

	// start, then success; both in the keeper's log
	entries, _ := activity.Recent(10)

	if 2 != len(entries) || domain.EntityLantern != entries[0].EntityType {
		t.Fatalf("expected 2 lantern activity entries, got %+v", entries)
	}
}

func TestHoistBuildFailureLeavesLiveLinkUntouched(t *testing.T) {
	lantern, state, activity, _, liveLink := newLanternHarness(t, []string{"false"}, 2)

	// plant a pre-existing live target to prove a failed build can't move it
	previous := filepath.Join(filepath.Dir(liveLink), "previous-generation")

	if err := os.MkdirAll(previous, 0755); nil != err {
		t.Fatalf("could not plant previous generation: %v", err)
	}

	if err := os.Symlink(previous, liveLink); nil != err {
		t.Fatalf("could not plant live link: %v", err)
	}

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("hoist failed to start: %v", err)
	}

	final := waitTerminal(t, lantern)

	if domain.LanternFailed != final.State {
		t.Fatalf("expected failed, got %q", final.State)
	}

	if !strings.Contains(final.Output, "build failed") {
		t.Fatalf("expected the failure reason in the output tail, got %q", final.Output)
	}

	if "" != state.Stamp || "" != final.LastHoistedAt {
		t.Fatalf("a failed hoist must not stamp lastHoistedAt")
	}

	target, _ := os.Readlink(liveLink)

	if previous != target {
		t.Fatalf("live link moved on a failed build: %q", target)
	}

	entries, _ := activity.Recent(10)

	if 2 != len(entries) || !strings.Contains(entries[0].Message, "failed") {
		t.Fatalf("expected a failure entry in the keeper's log, got %+v", entries)
	}
}

func TestHoistIsSingleFlight(t *testing.T) {
	gate := make(chan struct{})
	runner := &out_adapter.LanternFakeRunner{Gate: gate}
	releases := &out_adapter.LanternFakeReleaseStore{}
	state := &out_adapter.LanternFakeStateRepo{}
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		runner,
		releases,
		state,
		activity,
	)

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("first hoist failed to start: %v", err)
	}

	// the build is blocked on the gate; a second hoist must bounce
	status, err := lantern.Hoist()

	if !errors.Is(err, in_port.ErrHoistAlreadyRunning) {
		t.Fatalf("expected ErrHoistAlreadyRunning, got %v", err)
	}

	if domain.LanternBuilding != status.State {
		t.Fatalf("the conflict must return the running status, got %q", status.State)
	}

	close(gate)
	final := waitTerminal(t, lantern)

	if domain.LanternSucceeded != final.State {
		t.Fatalf("expected the gated hoist to succeed, got %q", final.State)
	}

	// and once terminal, a new hoist may start again
	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("hoist after completion must be allowed: %v", err)
	}

	waitTerminal(t, lantern)
}

func TestHoistPrunesOldGenerationsButNeverTheLiveOne(t *testing.T) {
	lantern, _, _, releasesDir, liveLink := newLanternHarness(t, []string{"true"}, 1)

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("first hoist failed to start: %v", err)
	}

	waitTerminal(t, lantern)

	// the stage consumed dist; rebuild it for the second hoist
	seedDist(t, filepath.Join(filepath.Dir(releasesDir), "site"))

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("second hoist failed to start: %v", err)
	}

	final := waitTerminal(t, lantern)

	if domain.LanternSucceeded != final.State {
		t.Fatalf("expected the second hoist to succeed, got %q", final.State)
	}

	entries, err := os.ReadDir(releasesDir)

	if nil != err {
		t.Fatalf("could not list releases: %v", err)
	}

	if 1 != len(entries) {
		t.Fatalf("expected keep=1 to leave exactly one generation, found %d", len(entries))
	}

	target, _ := os.Readlink(liveLink)

	if filepath.Join(releasesDir, entries[0].Name()) != target {
		t.Fatalf("the surviving generation must be the live one (live=%q, kept=%q)", target, entries[0].Name())
	}
}

func TestStatusOutputIsBoundedTail(t *testing.T) {
	long := strings.Repeat("line\n", 150)
	runner := &out_adapter.LanternFakeRunner{Output: long}
	releases := &out_adapter.LanternFakeReleaseStore{}

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		runner,
		releases,
		&out_adapter.LanternFakeStateRepo{},
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("hoist failed to start: %v", err)
	}

	final := waitTerminal(t, lantern)

	if lines := strings.Count(final.Output, "\n") + 1; 100 < lines {
		t.Fatalf("status output must be bounded to ~100 lines, got %d", lines)
	}
}

// The failure path appends a reason to the tail; the bound must hold at
// publication, not just after a successful build.
func TestFailureOutputIsBoundedTail(t *testing.T) {
	long := strings.Repeat("line\n", 150)
	runner := &out_adapter.LanternFakeRunner{Output: long, Err: errors.New("boom")}

	lantern := service.NewLanternService(
		service.LanternConfig{BuildCmd: []string{"stub"}, Keep: 2, Timeout: time.Second},
		runner,
		&out_adapter.LanternFakeReleaseStore{},
		&out_adapter.LanternFakeStateRepo{},
		service.NewActivityService(out_adapter.NewActivityFakeOutAdapter()),
	)

	if _, err := lantern.Hoist(); nil != err {
		t.Fatalf("hoist failed to start: %v", err)
	}

	final := waitTerminal(t, lantern)

	if domain.LanternFailed != final.State {
		t.Fatalf("expected failed, got %q", final.State)
	}

	if lines := strings.Count(final.Output, "\n") + 1; 100 < lines {
		t.Fatalf("failure output must stay bounded to ~100 lines, got %d", lines)
	}

	if !strings.Contains(final.Output, "build failed: boom") {
		t.Fatalf("the failure reason must survive the re-bounding, got tail %q", final.Output[len(final.Output)-80:])
	}
}

// the fakes must actually satisfy the ports they stand in for
var _ out_port.BuildRunner = &out_adapter.LanternFakeRunner{}
var _ out_port.ReleaseStore = &out_adapter.LanternFakeReleaseStore{}
var _ out_port.LanternStateRepo = &out_adapter.LanternFakeStateRepo{}
