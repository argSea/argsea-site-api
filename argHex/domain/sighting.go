package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"time"
)

type Sightings []Sighting

// Sighting is one anonymous ping from the shore: a page view, a light overlay
// opened, or a journal note opened. There is no visitor identity here, only a
// per-day hash that lets the tally count uniques without ever knowing who. At
// is a real date so a mongo TTL can sweep old sightings; Day is the UTC date
// string the aggregate groups on.
type Sighting struct {
	Id      string    `json:"id" bson:"_id,omitempty"`
	Kind    string    `json:"kind" bson:"kind"`
	Day     string    `json:"day" bson:"day"`
	Path    string    `json:"path" bson:"path"`
	Subject string    `json:"subject" bson:"subject,omitempty"`
	Port    string    `json:"port" bson:"port"`
	Visitor string    `json:"visitor" bson:"visitor"`
	At      time.Time `json:"at" bson:"at"`
}

// SightingBeacon is the client-supplied part of an ingest ping. The server
// trusts none of it for the tally: kind and path are validated, ref is only a
// hint the port bucket is derived from, and everything else about the stored
// sighting is derived server-side.
type SightingBeacon struct {
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Subject string `json:"subject"`
	Ref     string `json:"ref"`
}

// the three kinds of ping the shore sends: a page view, a light overlay
// opened on a project, a journal note opened.
const (
	SightingSail = "sail"
	SightingFlip = "flip"
	SightingRead = "read"
)

// the port a visitor came through, bucketed from the referrer so the raw
// referrer is never stored.
const (
	PortDirect    = "direct"
	PortSearch    = "search"
	PortFediverse = "fediverse"
	PortOther     = "other"
)

// visitorHashLen keeps 64 bits of the sha-256 as hex: plenty to tell daily
// visitors apart, short enough that the stored value is never a fingerprint.
const visitorHashLen = 16

// pathMaxLen bounds a stored path; a real site path is far shorter, anything
// longer is junk or an attempt to bloat the ledger.
const pathMaxLen = 512

// searchNeedles and fediNeedles are coarse on purpose: the tally only wants a
// bucket, not attribution, so a substring match on the referrer host is enough.
var searchNeedles = []string{"google", "bing", "duckduckgo", "kagi", "ecosia", "yandex"}

var fediNeedles = []string{"mastodon", "mstdn", "toot", "fosstodon", "hachyderm", "pleroma", "misskey", "pixelfed", "lemmy", "infosec.exchange", "mas.to"}

// botNeedles catch the obvious non-humans by user agent. The list is not
// exhaustive; it only has to keep the loudest crawlers out of the tally.
var botNeedles = []string{"bot", "crawl", "spider", "preview", "fetch", "curl", "wget", "headless"}

// ValidKind reports whether kind is one the shore is allowed to send.
func ValidKind(kind string) bool {
	return SightingSail == kind || SightingFlip == kind || SightingRead == kind
}

// ValidPath rejects anything that is not a plain site path: it must be rooted,
// bounded, and free of whitespace or control bytes.
func ValidPath(path string) bool {
	if "" == path {
		return false
	}

	if '/' != path[0] {
		return false
	}

	if len(path) > pathMaxLen {
		return false
	}

	for _, r := range path {
		if r < 0x20 || ' ' == r {
			return false
		}
	}

	return true
}

// IsBot reports whether a user agent looks like a crawler or an empty-agent
// fetch, so the ingest can drop it before it ever reaches the ledger.
func IsBot(userAgent string) bool {
	if "" == userAgent {
		return true
	}

	agent := strings.ToLower(userAgent)

	for _, needle := range botNeedles {
		if strings.Contains(agent, needle) {
			return true
		}
	}

	return false
}

// PortBucket sorts a referrer into a coarse bucket. An empty referrer is a
// direct arrival; a host that matches a known search or fediverse needle gets
// its bucket; everything else is other.
func PortBucket(ref string) string {
	if "" == ref {
		return PortDirect
	}

	host := refHost(ref)

	if "" == host {
		return PortOther
	}

	for _, needle := range searchNeedles {
		if strings.Contains(host, needle) {
			return PortSearch
		}
	}

	for _, needle := range fediNeedles {
		if strings.Contains(host, needle) {
			return PortFediverse
		}
	}

	return PortOther
}

// VisitorHash is the anonymous per-day identity: a truncated sha-256 over the
// salt, the day, the ip, and the user agent. The salt and the day rotate the
// value so it cannot be joined across days or back to a person; the same
// visitor on the same day lands on the same hash, which is all uniques needs.
func VisitorHash(salt string, day string, ip string, userAgent string) string {
	sum := sha256.Sum256([]byte(salt + "|" + day + "|" + ip + "|" + userAgent))

	return hex.EncodeToString(sum[:])[:visitorHashLen]
}

func refHost(ref string) string {
	parsed, err := url.Parse(ref)

	if nil != err {
		return ""
	}

	return strings.ToLower(parsed.Hostname())
}

// TrafficReport is the watch room's read of the tally over a window: totals,
// a zero-filled per-day series, the busiest weekday, the top flipped postcard
// and read note, and the port shares. It carries ids only; the admin resolves
// them to titles from its own store.
type TrafficReport struct {
	Uniques     int           `json:"uniques"`
	Sails       int           `json:"sails"`
	Days        []TrafficDay  `json:"days"`
	Busiest     string        `json:"busiest"`
	TopPostcard *TopPostcard  `json:"topPostcard"`
	TopNote     *TopNote      `json:"topNote"`
	Ports       []TrafficPort `json:"ports"`
}

type TrafficDay struct {
	Day     string `json:"day"`
	Sails   int    `json:"sails"`
	Uniques int    `json:"uniques"`
}

type TopPostcard struct {
	Subject string `json:"subject"`
	Flips   int    `json:"flips"`
}

type TopNote struct {
	Subject string `json:"subject"`
	Reads   int    `json:"reads"`
}

type TrafficPort struct {
	Port  string `json:"port"`
	Share int    `json:"share"`
}
