package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrHoistAlreadyRunning is returned by Hoist while a hoist is in flight — the
// adapter maps it to a 409 carrying the current status.
var ErrHoistAlreadyRunning = errors.New("a hoist is already running")

// LanternService is the deploy seam: Hoist kicks off the build → stage → swap
// pipeline in the background and returns immediately; Status is what the admin
// polls to watch it.
type LanternService interface {
	Hoist() (domain.LanternStatus, error)
	Status() domain.LanternStatus
}
