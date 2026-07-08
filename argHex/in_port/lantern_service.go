package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrHoistAlreadyRunning is returned by Hoist while a hoist is in flight; the
// adapter maps it to a 409 carrying the current status.
var ErrHoistAlreadyRunning = errors.New("a hoist is already running")

// ErrNoPreviousBuild is returned by Rollback when no earlier kept generation
// exists to point the live link at; the adapter maps it to a 409.
var ErrNoPreviousBuild = errors.New("no previous build to roll back to")

// LanternService is the deploy seam: Hoist kicks off the build → stage → swap
// pipeline in the background and returns immediately; Status is what the admin
// polls to watch it; Rollback re-points the live link at the previous kept
// build without rebuilding anything.
type LanternService interface {
	Hoist() (domain.LanternStatus, error)
	Status() domain.LanternStatus
	Rollback() (domain.LanternStatus, error)
}
