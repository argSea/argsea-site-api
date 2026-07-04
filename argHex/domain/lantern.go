package domain

// Lantern hoist states. A hoist is single-flight: building and swapping are the
// running states, succeeded/failed are terminal until the next hoist starts.
const (
	LanternIdle      = "idle"
	LanternBuilding  = "building"
	LanternSwapping  = "swapping"
	LanternSucceeded = "succeeded"
	LanternFailed    = "failed"
)

// LanternStatus is the polled hoist status the admin reads. Timestamps are the
// fixed-width RFC3339 UTC strings the rest of the API uses; Output is a bounded
// tail of build output (never the whole log, never the build environment).
type LanternStatus struct {
	State         string `json:"state"`
	StartedAt     string `json:"startedAt"`
	FinishedAt    string `json:"finishedAt"`
	LastHoistedAt string `json:"lastHoistedAt"`
	Output        string `json:"output"`
}

// LanternState is the small singleton document that keeps lastHoistedAt across
// restarts.
type LanternState struct {
	Id            string `json:"id" bson:"_id,omitempty"`
	LastHoistedAt string `json:"lastHoistedAt" bson:"lastHoistedAt,omitempty"`
}
