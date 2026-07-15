package service_test

import (
	"strings"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newWatch(t *testing.T) (in_port.WatchService, in_port.ActivityService) {
	t.Helper()

	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	watch := service.NewWatchService(out_adapter.NewWatchFakeOutAdapter(), activity)

	return watch, activity
}

func TestSaveUpsertsTheSingleton(t *testing.T) {
	watch, _ := newWatch(t)

	first, err := watch.Save(domain.Watch{Letter: "First entry.", Rotation: "out of the rotation: side quests"})

	if nil != err {
		t.Fatalf("first save failed: %v", err)
	}

	second, err := watch.Save(domain.Watch{Letter: "Second entry."})

	if nil != err {
		t.Fatalf("second save failed: %v", err)
	}

	if first.Id != second.Id {
		t.Fatalf("the singleton grew a second document: %q then %q", first.Id, second.Id)
	}

	// the write replaces the whole document; the first record's rotation line
	// must not bleed into the second
	kept := watch.Get()

	if "Second entry." != kept.Letter || "" != kept.Rotation {
		t.Fatalf("save must replace the record whole: %+v", kept)
	}
}

func TestSaveStampsKeptAtAndIgnoresTheClientValue(t *testing.T) {
	watch, _ := newWatch(t)

	saved, err := watch.Save(domain.Watch{Letter: "All quiet.", KeptAt: "1999-12-31T23:59:59.000000000Z"})

	if nil != err {
		t.Fatalf("save failed: %v", err)
	}

	if "" == saved.KeptAt {
		t.Fatalf("save must stamp keptAt: %+v", saved)
	}

	if "1999-12-31T23:59:59.000000000Z" == saved.KeptAt {
		t.Fatalf("a client-sent keptAt must be ignored, but it stuck: %+v", saved)
	}
}

func TestSaveTruncatesTheBearingsPastThree(t *testing.T) {
	watch, _ := newWatch(t)

	saved, err := watch.Save(domain.Watch{
		Letter: "Busy week at the light.",
		Bearings: []domain.WatchBearing{
			{Verb: "wrangling", Kind: "light", TargetId: "fastnet", Name: "Fastnet Rock"},
			{Verb: "logging", Kind: "note", TargetId: "fog-season", Name: "Fog season"},
			{Verb: "tinkering", Kind: "hobby", TargetId: "lens-work", Name: "Lens work"},
			{Verb: "sweeping", Kind: "none", TargetId: "", Name: "the gallery"},
			{Verb: "polishing", Kind: "none", TargetId: "", Name: "the brass"},
		},
	})

	if nil != err {
		t.Fatalf("save failed: %v", err)
	}

	if 3 != len(saved.Bearings) {
		t.Fatalf("the strip caps at three bearings, got %d", len(saved.Bearings))
	}

	// the cap keeps the first three in the order they arrived
	if "wrangling" != saved.Bearings[0].Verb || "logging" != saved.Bearings[1].Verb || "tinkering" != saved.Bearings[2].Verb {
		t.Fatalf("truncation must keep the first three in order: %+v", saved.Bearings)
	}
}

func TestAClearedWatchIsAValidSave(t *testing.T) {
	watch, _ := newWatch(t)

	if _, err := watch.Save(domain.Watch{Letter: "Off on the mail boat.", PostcardMediaId: "print-7"}); nil != err {
		t.Fatalf("seed save failed: %v", err)
	}

	// clearing is an authed write of an empty record; there is no delete route
	cleared, err := watch.Save(domain.Watch{})

	if nil != err {
		t.Fatalf("clearing the watch must be a valid save: %v", err)
	}

	kept := watch.Get()

	if "" != kept.Letter || "" != kept.PostcardMediaId {
		t.Fatalf("the cleared watch still carries the old record: %+v", kept)
	}

	if "" == cleared.KeptAt {
		t.Fatalf("even a cleared watch is stamped: %+v", cleared)
	}
}

func TestEveryWatchSaveWritesAKeepersLogLine(t *testing.T) {
	watch, activity := newWatch(t)

	if _, err := watch.Save(domain.Watch{Letter: "All quiet."}); nil != err {
		t.Fatalf("save failed: %v", err)
	}

	entries, err := activity.Recent(100)

	if nil != err {
		t.Fatalf("activity read failed: %v", err)
	}

	found := false

	for _, entry := range entries {
		if domain.EntityWatch == entry.EntityType && strings.Contains(entry.Message, "updated") {
			found = true
		}
	}

	if !found {
		t.Fatalf("no watch line reached the keeper's log: %+v", entries)
	}
}

func TestANeverKeptWatchAnswersWithEmptyHolds(t *testing.T) {
	watch, _ := newWatch(t)

	got := watch.Get()

	if nil == got.Bearings || nil == got.Quips {
		t.Fatalf("nil holds would go over the wire as null: %+v", got)
	}
}

func TestSaveNeverReturnsNilHolds(t *testing.T) {
	watch, _ := newWatch(t)

	saved, err := watch.Save(domain.Watch{Letter: "All quiet."})

	if nil != err {
		t.Fatalf("save failed: %v", err)
	}

	if nil == saved.Bearings || nil == saved.Quips {
		t.Fatalf("nil holds would go over the wire as null: %+v", saved)
	}
}
