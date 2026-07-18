package domain

// loginBarThreshold is the miss count that bars the door. The sixth bad hail
// from a client IP (its misses reaching this) sets barred; every hail after
// that stays barred. Kept here so nothing outside the domain hardcodes the count.
const loginBarThreshold = 6

// LoginLock is one client IP's standing at the admin login: how many bad hails
// it has sent and whether the door has been barred against it. A barred lock is
// the only gate on the login; the sole reset is deleting the doc on the server.
// barred is stored alongside misses so a glance at the doc reads the state, but
// it is always derivable from misses.
type LoginLock struct {
	Id     string `json:"id" bson:"_id,omitempty"`
	IP     string `json:"ip" bson:"ip"`
	Misses int    `json:"misses" bson:"misses"`
	Barred bool   `json:"barred" bson:"barred"`
}

// IsBarred reports whether this IP has crossed the bar threshold. It reads
// the miss count, not the stored flag, so the derivation is the point of truth
// even for a doc hand-edited to disagree.
func (l LoginLock) IsBarred() bool {
	return l.Misses >= loginBarThreshold
}

// Missed returns the lock after one more bad hail, barring it when the count
// crosses the threshold. It returns the next state rather than mutating in
// place, so a caller decides when the new standing is persisted.
func (l LoginLock) Missed() LoginLock {
	l.Misses++
	l.Barred = l.IsBarred()

	return l
}
