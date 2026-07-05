package out_adapter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/argSea/argsea-site-api/argHex/out_adapter"
)

// seedGenerations lays out a releases dir with the named generations and
// points the live link at target (when non-empty).
func seedGenerations(t *testing.T, generations []string, target string) (string, string) {
	t.Helper()

	root := t.TempDir()
	releasesDir := filepath.Join(root, "releases")
	liveLink := filepath.Join(root, "live")

	for _, name := range generations {
		if err := os.MkdirAll(filepath.Join(releasesDir, name), 0755); nil != err {
			t.Fatalf("could not seed generation %q: %v", name, err)
		}
	}

	if "" != target {
		if err := os.Symlink(filepath.Join(releasesDir, target), liveLink); nil != err {
			t.Fatalf("could not plant live link: %v", err)
		}
	}

	return releasesDir, liveLink
}

func TestPreviousReturnsTheGenerationBeforeTheLiveOne(t *testing.T) {
	releasesDir, liveLink := seedGenerations(t, []string{"gen-a", "gen-b", "gen-c"}, "gen-c")

	store := out_adapter.NewLanternFSReleaseAdapter(releasesDir, liveLink)
	previous, err := store.Previous()

	if nil != err {
		t.Fatalf("previous failed: %v", err)
	}

	if filepath.Join(releasesDir, "gen-b") != previous {
		t.Fatalf("expected gen-b, got %q", previous)
	}
}

func TestPreviousIsEmptyOnTheOldestGeneration(t *testing.T) {
	releasesDir, liveLink := seedGenerations(t, []string{"gen-a", "gen-b"}, "gen-a")

	store := out_adapter.NewLanternFSReleaseAdapter(releasesDir, liveLink)
	previous, err := store.Previous()

	if nil != err || "" != previous {
		t.Fatalf("expected no previous below the oldest generation, got %q / %v", previous, err)
	}
}

func TestPreviousIsEmptyWithoutALiveLink(t *testing.T) {
	releasesDir, liveLink := seedGenerations(t, []string{"gen-a"}, "")

	store := out_adapter.NewLanternFSReleaseAdapter(releasesDir, liveLink)
	previous, err := store.Previous()

	if nil != err || "" != previous {
		t.Fatalf("expected no previous without a live link, got %q / %v", previous, err)
	}
}
