package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/in_port"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

func newBlockSets() in_port.BlockSetService {
	activity := service.NewActivityService(out_adapter.NewActivityFakeOutAdapter())
	return service.NewBlockSetService(out_adapter.NewBlockSetFakeOutAdapter(), activity)
}

func TestBlockSetSeedPlantsTheHeaderOnceOnly(t *testing.T) {
	sets := newBlockSets()

	if err := sets.Seed(); nil != err {
		t.Fatalf("seed failed: %v", err)
	}

	seeded, _ := sets.List()

	if 1 != len(seeded) || "header" != seeded[0].Name {
		t.Fatalf("expected one header set seeded, got %+v", seeded)
	}

	if 4 != len(seeded[0].Blocks) {
		t.Fatalf("expected the header template to carry title/subhead/facts/meta, got %d blocks", len(seeded[0].Blocks))
	}

	// a second boot must not plant a second header
	if err := sets.Seed(); nil != err {
		t.Fatalf("second seed failed: %v", err)
	}

	after, _ := sets.List()

	if 1 != len(after) {
		t.Fatalf("expected the seed to be a no-op the second time, got %d sets", len(after))
	}
}

func TestBlockSetCreateAndDelete(t *testing.T) {
	sets := newBlockSets()

	saved, err := sets.Create(domain.BlockSet{
		Name:   "intro",
		Blocks: domain.Blocks{domain.Block{"kind": "paragraph", "text": "hello"}},
	})

	if nil != err {
		t.Fatalf("create failed: %v", err)
	}

	if "" == saved.Id || "intro" != saved.Name {
		t.Fatalf("expected the set saved with an id, got %+v", saved)
	}

	if err := sets.Delete(saved.Id); nil != err {
		t.Fatalf("delete failed: %v", err)
	}

	remaining, _ := sets.List()

	if 0 != len(remaining) {
		t.Fatalf("expected the set removed, got %d", len(remaining))
	}

	// deleting a set that isn't there is an error, not a silent success
	if err := sets.Delete("nope"); nil == err {
		t.Fatalf("expected delete of an unknown set to error")
	}
}
