package in_port

import (
	"errors"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// ErrSightingRejected marks an ingest the endpoint refuses outright: an unknown
// kind or a junk path. The adapter answers it 400, while a storage failure
// keeps surfacing as a 500.
var ErrSightingRejected = errors.New("sighting rejected")

// SightingService is the harbor's tally seam. The public shore records
// anonymous pings through Record; the watch room reads only aggregates back
// through Traffic, never a single visitor.
type SightingService interface {
	Record(beacon domain.SightingBeacon, ip string, userAgent string) error
	Traffic(days int) (domain.TrafficReport, error)
}
