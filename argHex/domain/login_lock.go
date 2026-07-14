package domain

// loginStrikeThreshold is the miss count that strikes the light. The sixth bad
// hail from a client IP (its misses reaching this) sets struck; every hail after
// that stays struck. Kept here so nothing outside the domain hardcodes the count.
const loginStrikeThreshold = 6

// LoginLock is one client IP's standing at the admin login: how many bad hails
// it has sent and whether the light has been struck against it. A struck lock is
// the only gate on the login; the sole reset is deleting the doc on the server.
// struck is stored alongside misses so a glance at the doc reads the state, but
// it is always derivable from misses.
type LoginLock struct {
	Id     string `json:"id" bson:"_id,omitempty"`
	IP     string `json:"ip" bson:"ip"`
	Misses int    `json:"misses" bson:"misses"`
	Struck bool   `json:"struck" bson:"struck"`
}

// IsStruck reports whether this IP has crossed the strike threshold. It reads
// the miss count, not the stored flag, so the derivation is the point of truth
// even for a doc hand-edited to disagree.
func (l LoginLock) IsStruck() bool {
	return l.Misses >= loginStrikeThreshold
}

// Missed returns the lock after one more bad hail, striking it when the count
// crosses the threshold. It returns the next state rather than mutating in
// place, so a caller decides when the new standing is persisted.
func (l LoginLock) Missed() LoginLock {
	l.Misses++
	l.Struck = l.IsStruck()

	return l
}
