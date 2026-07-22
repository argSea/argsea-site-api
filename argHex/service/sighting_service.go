package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// sightingDayLayout is the UTC calendar day a sighting is bucketed under. It
// sorts lexically the same as chronologically, so the window read ranges on it
// as a plain string.
const sightingDayLayout = "2006-01-02"

// the aggregate window is clamped to a sane band: at least a day, at most a
// quarter, defended here so no caller can ask for an unbounded scan.
const (
	sightingMinDays = 1
	sightingMaxDays = 90
)

type sightingService struct {
	repo out_port.SightingRepo
	salt string
}

func NewSightingService(repo out_port.SightingRepo, salt string) in_port.SightingService {
	return sightingService{
		repo: repo,
		salt: salt,
	}
}

// Record validates a ping, drops the bots, derives everything the client is not
// trusted for, and lands the sighting. A dropped bot and a stored ping both
// return nil: the caller never learns which, and only a junk ping or a storage
// failure surfaces an error.
func (s sightingService) Record(beacon domain.SightingBeacon, ip string, userAgent string) error {
	if !domain.ValidKind(beacon.Kind) {
		return fmt.Errorf("unknown sighting kind %q: %w", beacon.Kind, in_port.ErrSightingRejected)
	}

	if !domain.ValidPath(beacon.Path) {
		return fmt.Errorf("junk sighting path: %w", in_port.ErrSightingRejected)
	}

	if domain.IsBot(userAgent) {
		return nil
	}

	now := time.Now().UTC()
	day := now.Format(sightingDayLayout)

	sighting := domain.Sighting{
		Kind:    beacon.Kind,
		Day:     day,
		Path:    beacon.Path,
		Subject: beacon.Subject,
		Port:    domain.PortBucket(beacon.Ref),
		Visitor: domain.VisitorHash(s.salt, day, ip, userAgent),
		At:      now,
	}

	_, err := s.repo.Add(sighting)

	return err
}

// Traffic folds the window of sightings into the watch room's read: totals of
// sails and bottles, a zero-filled per-day series, the busiest weekday, the top
// flipped postcard, read note, and visited hobby, the port shares, and the
// per-ship flare roll call. The roll call and its total ride every flare ever
// sent, not just the window, so the flare tally is a second, unfiltered read.
func (s sightingService) Traffic(days int) (domain.TrafficReport, error) {
	days = clampDays(days)
	now := time.Now().UTC()
	since := now.AddDate(0, 0, -(days - 1)).Format(sightingDayLayout)

	window, err := s.repo.Window(since)

	if nil != err {
		return domain.TrafficReport{}, err
	}

	flares, err := s.repo.Flares()

	if nil != err {
		return domain.TrafficReport{}, err
	}

	return foldTraffic(window, flares, now, days), nil
}

// foldTraffic is the whole aggregate in one pass over the window plus one pass
// over the all-time flares, so it stays a pure function of the rows and the
// clock: easy to read, easy to test. Flares are the one tally counted by
// distinct visitor per ship, not by raw ping, and the one tally that never
// ages out of the window.
func foldTraffic(window domain.Sightings, flares domain.Sightings, now time.Time, days int) domain.TrafficReport {
	order := make([]string, 0, days)
	daySails := map[string]int{}
	dayVisitors := map[string]map[string]bool{}

	for k := days - 1; k >= 0; k-- {
		day := now.AddDate(0, 0, -k).Format(sightingDayLayout)
		order = append(order, day)
		daySails[day] = 0
		dayVisitors[day] = map[string]bool{}
	}

	windowVisitors := map[string]bool{}
	portCounts := map[string]int{}
	flipCounts := map[string]int{}
	readCounts := map[string]int{}
	visitCounts := map[string]int{}
	totalSails := 0
	totalBottles := 0

	for _, sighting := range window {
		switch sighting.Kind {
		case domain.SightingSail:
			totalSails++
			windowVisitors[sighting.Visitor] = true
			portCounts[sighting.Port]++

			if _, ok := daySails[sighting.Day]; ok {
				daySails[sighting.Day]++
				dayVisitors[sighting.Day][sighting.Visitor] = true
			}
		case domain.SightingFlip:
			if "" != sighting.Subject {
				flipCounts[sighting.Subject]++
			}
		case domain.SightingRead:
			if "" != sighting.Subject {
				readCounts[sighting.Subject]++
			}
		case domain.SightingVisit:
			if "" != sighting.Subject {
				visitCounts[sighting.Subject]++
			}
		case domain.SightingBottle:
			totalBottles++
		}
	}

	// flares count distinct visitors per ship, not raw pings, so the roll call
	// holds a visitor set per subject rather than a plain counter. It rides the
	// all-time flares passed in, not the window, so the tally never ages out.
	flareVisitors := map[string]map[string]bool{}

	for _, sighting := range flares {
		if "" != sighting.Subject {
			if nil == flareVisitors[sighting.Subject] {
				flareVisitors[sighting.Subject] = map[string]bool{}
			}

			flareVisitors[sighting.Subject][sighting.Visitor] = true
		}
	}

	flareRolls, totalFlares := flareRollCall(flareVisitors)

	daySeries := make([]domain.TrafficDay, 0, len(order))
	busiest := ""
	busiestSails := 0

	for _, day := range order {
		sails := daySails[day]
		daySeries = append(daySeries, domain.TrafficDay{
			Day:     day,
			Sails:   sails,
			Uniques: len(dayVisitors[day]),
		})

		if sails > busiestSails {
			busiestSails = sails
			busiest = weekdayName(day)
		}
	}

	return domain.TrafficReport{
		Uniques:     len(windowVisitors),
		Sails:       totalSails,
		Bottles:     totalBottles,
		Flares:      totalFlares,
		Days:        daySeries,
		Busiest:     busiest,
		TopPostcard: topPostcard(flipCounts),
		TopNote:     topNote(readCounts),
		TopHobby:    topHobby(visitCounts),
		Ports:       portShares(portCounts, totalSails),
		FlareRolls:  flareRolls,
	}
}

func clampDays(days int) int {
	if days < sightingMinDays {
		return sightingMinDays
	}

	if days > sightingMaxDays {
		return sightingMaxDays
	}

	return days
}

func weekdayName(day string) string {
	parsed, err := time.Parse(sightingDayLayout, day)

	if nil != err {
		return ""
	}

	return strings.ToLower(parsed.Weekday().String())
}

func topPostcard(flipCounts map[string]int) *domain.TopPostcard {
	subject, count := topSubject(flipCounts)

	if 0 == count {
		return nil
	}

	return &domain.TopPostcard{Subject: subject, Flips: count}
}

func topNote(readCounts map[string]int) *domain.TopNote {
	subject, count := topSubject(readCounts)

	if 0 == count {
		return nil
	}

	return &domain.TopNote{Subject: subject, Reads: count}
}

func topHobby(visitCounts map[string]int) *domain.TopHobby {
	subject, count := topSubject(visitCounts)

	if 0 == count {
		return nil
	}

	return &domain.TopHobby{Subject: subject, Visits: count}
}

// topSubject picks the most-counted subject, breaking ties on the lower id so
// the winner never rides mongo's or the map's iteration order.
func topSubject(counts map[string]int) (string, int) {
	top := ""
	best := 0

	for subject, count := range counts {
		if count > best || (count == best && subject < top) {
			best = count
			top = subject
		}
	}

	return top, best
}

// flareRollCall turns the per-ship visitor sets into the roll call: one entry
// per ship carrying its distinct-visitor count, sorted by count and then by the
// lower subject so the order never rides the map's iteration. Ships with no
// distinct visitors never enter the sets, so no zero counts reach the response,
// which stays an empty slice (never null) when no flares were sent. The returned
// total is the sum across the roll call.
func flareRollCall(flareVisitors map[string]map[string]bool) ([]domain.FlareRoll, int) {
	rolls := []domain.FlareRoll{}
	total := 0

	for subject, visitors := range flareVisitors {
		count := len(visitors)

		if 0 == count {
			continue
		}

		rolls = append(rolls, domain.FlareRoll{Subject: subject, Flares: count})
		total += count
	}

	sort.Slice(rolls, func(i, j int) bool {
		if rolls[i].Flares != rolls[j].Flares {
			return rolls[i].Flares > rolls[j].Flares
		}

		return rolls[i].Subject < rolls[j].Subject
	})

	return rolls, total
}

// portShares turns raw per-port sail counts into integer percentages, sorted
// by share and then by port so the list is stable. Buckets with no sails never
// enter the map, so none reach the response.
func portShares(portCounts map[string]int, totalSails int) []domain.TrafficPort {
	shares := []domain.TrafficPort{}

	if 0 == totalSails {
		return shares
	}

	for port, count := range portCounts {
		shares = append(shares, domain.TrafficPort{Port: port, Share: count * 100 / totalSails})
	}

	sort.Slice(shares, func(i, j int) bool {
		if shares[i].Share != shares[j].Share {
			return shares[i].Share > shares[j].Share
		}

		return shares[i].Port < shares[j].Port
	})

	return shares
}
