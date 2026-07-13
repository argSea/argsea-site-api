package service_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/out_port"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newSightings(t *testing.T) (in_port.SightingService, out_port.SightingRepo) {
	t.Helper()

	repo := out_adapter.NewSightingFakeOutAdapter()

	return service.NewSightingService(repo, "test-salt"), repo
}

func addSail(t *testing.T, repo out_port.SightingRepo, day string, visitor string, port string) {
	t.Helper()

	if _, err := repo.Add(domain.Sighting{Kind: domain.SightingSail, Day: day, Path: "/", Port: port, Visitor: visitor, At: time.Now().UTC()}); nil != err {
		t.Fatalf("add sail failed: %v", err)
	}
}

func addEvent(t *testing.T, repo out_port.SightingRepo, day string, kind string, subject string) {
	t.Helper()

	if _, err := repo.Add(domain.Sighting{Kind: kind, Day: day, Path: "/projects/x", Subject: subject, Visitor: "v", At: time.Now().UTC()}); nil != err {
		t.Fatalf("add event failed: %v", err)
	}
}

func addFlare(t *testing.T, repo out_port.SightingRepo, day string, subject string, visitor string) {
	t.Helper()

	if _, err := repo.Add(domain.Sighting{Kind: domain.SightingFlare, Day: day, Path: "/hobbies", Subject: subject, Visitor: visitor, At: time.Now().UTC()}); nil != err {
		t.Fatalf("add flare failed: %v", err)
	}
}

func TestRecordStoresEachKind(t *testing.T) {
	svc, repo := newSightings(t)

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingSail, Path: "/projects/foo", Ref: "https://www.google.com/search"}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a sail failed: %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingFlip, Path: "/projects/foo", Subject: "cat-cascade", Ref: ""}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a flip failed: %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingRead, Path: "/journal/fog", Subject: "note-fog", Ref: ""}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a read failed: %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingVisit, Path: "/hobbies", Subject: "graveyard-chess", Ref: ""}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a visit failed: %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingBottle, Path: "/", Subject: "", Ref: ""}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a bottle failed: %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingFlare, Path: "/hobbies", Subject: "piano", Ref: ""}, "203.0.113.7", "Mozilla/5.0"); nil != err {
		t.Fatalf("recording a flare failed: %v", err)
	}

	window, err := repo.Window("")

	if nil != err {
		t.Fatalf("window read failed: %v", err)
	}

	if 6 != len(window) {
		t.Fatalf("expected all six kinds stored, got %d", len(window))
	}

	flare := findKind(t, window, domain.SightingFlare)

	if "piano" != flare.Subject {
		t.Fatalf("a stored flare must carry its hobby subject, got %q", flare.Subject)
	}

	sail := findKind(t, window, domain.SightingSail)
	today := time.Now().UTC().Format("2006-01-02")

	if domain.PortSearch != sail.Port {
		t.Fatalf("a google referrer must bucket to search, got %q", sail.Port)
	}

	if "" == sail.Visitor {
		t.Fatalf("a stored sail must carry a derived visitor hash")
	}

	if today != sail.Day {
		t.Fatalf("expected the sail stamped with today %q, got %q", today, sail.Day)
	}

	if sail.At.IsZero() {
		t.Fatalf("a stored sail must carry an at timestamp for the TTL")
	}

	if "/projects/foo" != sail.Path {
		t.Fatalf("the sail path did not round-trip, got %q", sail.Path)
	}
}

func TestRecordDropsBots(t *testing.T) {
	svc, repo := newSightings(t)

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingSail, Path: "/", Ref: ""}, "203.0.113.7", "Googlebot/2.1"); nil != err {
		t.Fatalf("a dropped bot must not error, got %v", err)
	}

	window, _ := repo.Window("")

	if 0 != len(window) {
		t.Fatalf("a bot ping must store nothing, got %d", len(window))
	}
}

func TestRecordRejectsJunk(t *testing.T) {
	svc, repo := newSightings(t)

	if err := svc.Record(domain.SightingBeacon{Kind: "click", Path: "/"}, "203.0.113.7", "Mozilla/5.0"); !errors.Is(err, in_port.ErrSightingRejected) {
		t.Fatalf("an unknown kind must be rejected, got %v", err)
	}

	if err := svc.Record(domain.SightingBeacon{Kind: domain.SightingSail, Path: "no-slash"}, "203.0.113.7", "Mozilla/5.0"); !errors.Is(err, in_port.ErrSightingRejected) {
		t.Fatalf("a junk path must be rejected, got %v", err)
	}

	window, _ := repo.Window("")

	if 0 != len(window) {
		t.Fatalf("a rejected ping must store nothing, got %d", len(window))
	}
}

func TestTrafficAggregateShape(t *testing.T) {
	svc, repo := newSightings(t)
	today := time.Now().UTC()
	day := func(k int) string { return today.AddDate(0, 0, -k).Format("2006-01-02") }

	// today: three sails from two visitors, ports two search one fediverse
	addSail(t, repo, day(0), "v1", domain.PortSearch)
	addSail(t, repo, day(0), "v1", domain.PortSearch)
	addSail(t, repo, day(0), "v2", domain.PortFediverse)
	// two days ago: one sail from a third visitor, direct
	addSail(t, repo, day(2), "v3", domain.PortDirect)
	// flips and a read, resolving to their tops
	addEvent(t, repo, day(0), domain.SightingFlip, "cat-cascade")
	addEvent(t, repo, day(0), domain.SightingFlip, "cat-cascade")
	addEvent(t, repo, day(0), domain.SightingFlip, "otherbook")
	addEvent(t, repo, day(1), domain.SightingRead, "note-fog")
	// visits resolving to a top hobby, and bottles that only count in total
	addEvent(t, repo, day(0), domain.SightingVisit, "graveyard-chess")
	addEvent(t, repo, day(0), domain.SightingVisit, "graveyard-chess")
	addEvent(t, repo, day(1), domain.SightingVisit, "abandoned-origami")
	addEvent(t, repo, day(0), domain.SightingBottle, "")
	addEvent(t, repo, day(2), domain.SightingBottle, "")
	// flares from distinct visitors: piano ahead of chess in the roll call
	addFlare(t, repo, day(0), "piano", "fv1")
	addFlare(t, repo, day(0), "piano", "fv2")
	addFlare(t, repo, day(1), "chess", "fv1")

	report, err := svc.Traffic(7)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 7 != len(report.Days) {
		t.Fatalf("expected a seven-day zero-filled series, got %d", len(report.Days))
	}

	if day(6) != report.Days[0].Day || day(0) != report.Days[6].Day {
		t.Fatalf("days must run oldest to newest, got %q .. %q", report.Days[0].Day, report.Days[6].Day)
	}

	if 4 != report.Sails {
		t.Fatalf("expected four total sails, got %d", report.Sails)
	}

	if 3 != report.Uniques {
		t.Fatalf("expected three unique visitors, got %d", report.Uniques)
	}

	if 3 != report.Days[6].Sails || 2 != report.Days[6].Uniques {
		t.Fatalf("today should show 3 sails / 2 uniques, got %d / %d", report.Days[6].Sails, report.Days[6].Uniques)
	}

	if 0 != report.Days[5].Sails {
		t.Fatalf("a day with only a read must zero-fill its sails, got %d", report.Days[5].Sails)
	}

	if strings.ToLower(today.Weekday().String()) != report.Busiest {
		t.Fatalf("busiest should be today's weekday %q, got %q", strings.ToLower(today.Weekday().String()), report.Busiest)
	}

	if nil == report.TopPostcard || "cat-cascade" != report.TopPostcard.Subject || 2 != report.TopPostcard.Flips {
		t.Fatalf("expected cat-cascade with 2 flips as top postcard, got %+v", report.TopPostcard)
	}

	if nil == report.TopNote || "note-fog" != report.TopNote.Subject || 1 != report.TopNote.Reads {
		t.Fatalf("expected note-fog with 1 read as top note, got %+v", report.TopNote)
	}

	if 2 != report.Bottles {
		t.Fatalf("expected two bottles served, got %d", report.Bottles)
	}

	if nil == report.TopHobby || "graveyard-chess" != report.TopHobby.Subject || 2 != report.TopHobby.Visits {
		t.Fatalf("expected graveyard-chess with 2 visits as top hobby, got %+v", report.TopHobby)
	}

	if 3 != report.Flares {
		t.Fatalf("expected three distinct-visitor flares, got %d", report.Flares)
	}

	if 2 != len(report.FlareRolls) {
		t.Fatalf("expected two ships in the roll call, got %+v", report.FlareRolls)
	}

	if "piano" != report.FlareRolls[0].Subject || 2 != report.FlareRolls[0].Flares {
		t.Fatalf("expected piano leading the roll call with 2, got %+v", report.FlareRolls[0])
	}

	if "chess" != report.FlareRolls[1].Subject || 1 != report.FlareRolls[1].Flares {
		t.Fatalf("expected chess trailing with 1, got %+v", report.FlareRolls[1])
	}

	shares := map[string]int{}
	sum := 0

	for _, p := range report.Ports {
		shares[p.Port] = p.Share
		sum += p.Share
	}

	if 50 != shares[domain.PortSearch] || 25 != shares[domain.PortFediverse] || 25 != shares[domain.PortDirect] {
		t.Fatalf("port shares off: %+v", report.Ports)
	}

	if sum < 95 || sum > 100 {
		t.Fatalf("port shares should sum to ~100, got %d", sum)
	}
}

func TestTrafficTopHobbyIsNullWithoutVisits(t *testing.T) {
	svc, repo := newSightings(t)
	today := time.Now().UTC()
	day := func(k int) string { return today.AddDate(0, 0, -k).Format("2006-01-02") }

	// bottles are served but no hobby graveyard record is ever opened
	addSail(t, repo, day(0), "v1", domain.PortDirect)
	addEvent(t, repo, day(0), domain.SightingBottle, "")
	addEvent(t, repo, day(0), domain.SightingBottle, "")

	report, err := svc.Traffic(7)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 2 != report.Bottles {
		t.Fatalf("expected two bottles counted without any visit, got %d", report.Bottles)
	}

	if nil != report.TopHobby {
		t.Fatalf("no visits means a null top hobby, got %+v", report.TopHobby)
	}
}

func TestTrafficIsEmptyButShapedWhenNothingHappened(t *testing.T) {
	svc, _ := newSightings(t)

	report, err := svc.Traffic(7)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 7 != len(report.Days) {
		t.Fatalf("an empty window still zero-fills its days, got %d", len(report.Days))
	}

	if 0 != report.Sails || 0 != report.Uniques {
		t.Fatalf("an empty window has no sails or uniques, got %d / %d", report.Sails, report.Uniques)
	}

	if 0 != report.Bottles {
		t.Fatalf("an empty window has no bottles, got %d", report.Bottles)
	}

	if "" != report.Busiest {
		t.Fatalf("an empty window has no busiest day, got %q", report.Busiest)
	}

	if nil != report.TopPostcard || nil != report.TopNote || nil != report.TopHobby {
		t.Fatalf("an empty window has null tops, got %+v / %+v / %+v", report.TopPostcard, report.TopNote, report.TopHobby)
	}

	if nil == report.Ports || 0 != len(report.Ports) {
		t.Fatalf("an empty window has an empty, non-null ports list, got %+v", report.Ports)
	}

	if 0 != report.Flares || nil == report.FlareRolls || 0 != len(report.FlareRolls) {
		t.Fatalf("an empty window has no flares and an empty, non-null roll call, got %d / %+v", report.Flares, report.FlareRolls)
	}

	body, err := json.Marshal(report)

	if nil != err {
		t.Fatalf("report did not marshal: %v", err)
	}

	if !strings.Contains(string(body), `"topPostcard":null`) || !strings.Contains(string(body), `"ports":[]`) {
		t.Fatalf("wire shape must carry null tops and an empty ports array, got %s", body)
	}

	if !strings.Contains(string(body), `"bottles":0`) || !strings.Contains(string(body), `"topHobby":null`) {
		t.Fatalf("wire shape must carry a zero bottle count and a null top hobby, got %s", body)
	}

	if !strings.Contains(string(body), `"flares":0`) || !strings.Contains(string(body), `"flareRolls":[]`) {
		t.Fatalf("wire shape must carry a zero flare count and an empty roll call array, got %s", body)
	}
}

func TestTrafficFlareRollCountsDistinctVisitors(t *testing.T) {
	svc, repo := newSightings(t)
	today := time.Now().UTC()
	day := func(k int) string { return today.AddDate(0, 0, -k).Format("2006-01-02") }

	// piano: the same visitor flares twice (counts once) plus a second visitor
	addFlare(t, repo, day(0), "piano", "fv1")
	addFlare(t, repo, day(0), "piano", "fv1")
	addFlare(t, repo, day(1), "piano", "fv2")
	// chess: two distinct visitors
	addFlare(t, repo, day(0), "chess", "fv1")
	addFlare(t, repo, day(0), "chess", "fv3")
	// kite: a single visitor
	addFlare(t, repo, day(0), "kite", "fv9")

	report, err := svc.Traffic(7)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	// piano 2 + chess 2 + kite 1: the total is the sum over the roll call
	if 5 != report.Flares {
		t.Fatalf("expected five distinct-visitor flares total, got %d", report.Flares)
	}

	// descending by count, ties broken on the lower subject: chess before piano
	want := []domain.FlareRoll{
		{Subject: "chess", Flares: 2},
		{Subject: "piano", Flares: 2},
		{Subject: "kite", Flares: 1},
	}

	if len(want) != len(report.FlareRolls) {
		t.Fatalf("expected three ships in the roll call, got %+v", report.FlareRolls)
	}

	for i, roll := range want {
		if roll != report.FlareRolls[i] {
			t.Fatalf("roll call off at %d: got %+v, want %+v", i, report.FlareRolls[i], roll)
		}
	}
}

func TestTrafficFlareRollsAreEmptyWithoutFlares(t *testing.T) {
	svc, repo := newSightings(t)
	today := time.Now().UTC()
	day := func(k int) string { return today.AddDate(0, 0, -k).Format("2006-01-02") }

	addSail(t, repo, day(0), "v1", domain.PortDirect)

	report, err := svc.Traffic(7)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 0 != report.Flares {
		t.Fatalf("expected no flares without any flare ping, got %d", report.Flares)
	}

	if nil == report.FlareRolls || 0 != len(report.FlareRolls) {
		t.Fatalf("expected an empty, non-null roll call, got %+v", report.FlareRolls)
	}
}

func TestTrafficClampsTheWindow(t *testing.T) {
	svc, _ := newSightings(t)

	wide, err := svc.Traffic(1000)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 90 != len(wide.Days) {
		t.Fatalf("an over-wide window must clamp to 90 days, got %d", len(wide.Days))
	}

	narrow, err := svc.Traffic(0)

	if nil != err {
		t.Fatalf("traffic read failed: %v", err)
	}

	if 1 != len(narrow.Days) {
		t.Fatalf("a zero window must clamp to a single day, got %d", len(narrow.Days))
	}
}

func findKind(t *testing.T, window domain.Sightings, kind string) domain.Sighting {
	t.Helper()

	for _, sighting := range window {
		if kind == sighting.Kind {
			return sighting
		}
	}

	t.Fatalf("no %q sighting in the window", kind)

	return domain.Sighting{}
}
