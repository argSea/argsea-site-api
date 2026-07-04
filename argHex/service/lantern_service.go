package service

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// lanternOutputTailLines bounds how much build output the status payload
// carries — the admin sees the tail of the log, not the whole thing.
const lanternOutputTailLines = 100

// LanternConfig is everything the hoist pipeline needs to know about the box:
// where the site checkout lives, how to build it, and how deep the release
// history is kept.
type LanternConfig struct {
	SiteDir  string            // the Astro checkout — the build's working directory
	BuildCmd []string          // argv array, never a shell string
	DistDir  string            // build output, relative to SiteDir
	Keep     int               // generations to retain after a hoist
	Timeout  time.Duration     // hard cap on the build command
	Env      map[string]string // merged over the process env for the build
}

type lanternService struct {
	cfg      LanternConfig
	runner   out_port.BuildRunner
	releases out_port.ReleaseStore
	state    out_port.LanternStateRepo
	activity in_port.ActivityService

	mu     sync.Mutex
	status domain.LanternStatus
}

// NewLanternService wires the hoist pipeline onto its seams and loads the
// persisted lastHoistedAt so the status is right from the first poll.
func NewLanternService(cfg LanternConfig, runner out_port.BuildRunner, releases out_port.ReleaseStore, state out_port.LanternStateRepo, activity in_port.ActivityService) in_port.LanternService {
	lastHoistedAt, err := state.LastHoistedAt()

	if nil != err {
		log.Printf("lantern could not load lastHoistedAt: %v\n", err)
	}

	return &lanternService{
		cfg:      cfg,
		runner:   runner,
		releases: releases,
		state:    state,
		activity: activity,
		status: domain.LanternStatus{
			State:         domain.LanternIdle,
			LastHoistedAt: lastHoistedAt,
		},
	}
}

// Hoist starts a hoist in the background and returns the fresh status. It is
// single-flight: while one is running it returns the current status with
// in_port.ErrHoistAlreadyRunning and starts nothing.
func (l *lanternService) Hoist() (domain.LanternStatus, error) {
	l.mu.Lock()

	if domain.LanternBuilding == l.status.State || domain.LanternSwapping == l.status.State {
		status := l.status
		l.mu.Unlock()

		return status, in_port.ErrHoistAlreadyRunning
	}

	l.status = domain.LanternStatus{
		State:         domain.LanternBuilding,
		StartedAt:     nowStamp(),
		LastHoistedAt: l.status.LastHoistedAt,
	}

	status := l.status
	l.mu.Unlock()

	// the activity write happens outside the lock so a slow store never blocks
	// status polling
	l.record("lantern hoist started")

	go l.run()

	return status, nil
}

// Status returns a copy of the current hoist status for polling.
func (l *lanternService) Status() domain.LanternStatus {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.status
}

// run is the hoist pipeline: build → stage → swap → prune → persist. A build
// or filesystem failure leaves the live link exactly as it was.
func (l *lanternService) run() {
	output, buildErr := l.runner.Run(l.cfg.SiteDir, l.cfg.BuildCmd, l.cfg.Env, l.cfg.Timeout)
	tail := tailLines(output, lanternOutputTailLines)

	if nil != buildErr {
		l.fail(tail+"\nbuild failed: "+buildErr.Error(), "lantern hoist failed: build error")
		return
	}

	l.transition(domain.LanternSwapping, tail)

	generation, stageErr := l.releases.Stage(filepath.Join(l.cfg.SiteDir, l.cfg.DistDir))

	if nil != stageErr {
		l.fail(tail+"\nstage failed: "+stageErr.Error(), "lantern hoist failed: stage error")
		return
	}

	if swapErr := l.releases.Swap(generation); nil != swapErr {
		l.fail(tail+"\nswap failed: "+swapErr.Error(), "lantern hoist failed: swap error")
		return
	}

	if pruneErr := l.releases.Prune(l.cfg.Keep); nil != pruneErr {
		// the swap already landed — a prune failure leaves extra generations on
		// disk, it doesn't un-ship the site
		log.Printf("lantern prune failed: %v\n", pruneErr)
	}

	stamp := nowStamp()

	if saveErr := l.state.SaveLastHoistedAt(stamp); nil != saveErr {
		log.Printf("lantern could not persist lastHoistedAt: %v\n", saveErr)
	}

	l.mu.Lock()
	l.status.State = domain.LanternSucceeded
	l.status.FinishedAt = stamp
	l.status.LastHoistedAt = stamp
	l.status.Output = tail
	l.mu.Unlock()

	l.record("lantern hoisted")
}

// transition moves the running hoist to a new non-terminal state.
func (l *lanternService) transition(state string, output string) {
	l.mu.Lock()
	l.status.State = state
	l.status.Output = output
	l.mu.Unlock()
}

// fail terminates the hoist as failed, keeping the output tail plus the reason.
func (l *lanternService) fail(output string, message string) {
	l.mu.Lock()
	l.status.State = domain.LanternFailed
	l.status.FinishedAt = nowStamp()
	l.status.Output = output
	l.mu.Unlock()

	l.record(message)
}

// record writes a ship's-log entry; a logging failure never blocks the hoist.
func (l *lanternService) record(message string) {
	if err := l.activity.Record(message, domain.EntityLantern, ""); nil != err {
		log.Printf("activity record failed for lantern: %v\n", err)
	}
}

// tailLines returns at most the last n lines of s.
func tailLines(s string, n int) string {
	s = strings.TrimRight(s, "\n")

	if "" == s {
		return ""
	}

	lines := strings.Split(s, "\n")

	if len(lines) <= n {
		return s
	}

	return strings.Join(lines[len(lines)-n:], "\n")
}
