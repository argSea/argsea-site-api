package out_adapter

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// lanternGenerationPrefix names generation directories so Prune can tell them
// apart from anything else living under releasesDir.
const lanternGenerationPrefix = "gen-"

// lanternGenerationStamp is fixed-width so a plain string sort of generation
// names is a chronological sort.
const lanternGenerationStamp = "20060102-150405.000000000"

type lanternFSReleaseAdapter struct {
	releasesDir string
	liveLink    string
}

// NewLanternFSReleaseAdapter returns the real ReleaseStore over the release
// layout: timestamped generation directories under releasesDir and a liveLink
// symlink that nginx serves from. releasesDir is assumed to be on the same
// filesystem as the build output (renames must be atomic).
func NewLanternFSReleaseAdapter(releasesDir string, liveLink string) out_port.ReleaseStore {
	return lanternFSReleaseAdapter{
		releasesDir: releasesDir,
		liveLink:    liveLink,
	}
}

// Stage moves the built dist into a fresh generation directory and returns its
// path. The rename means a second build starts from an empty dist.
func (l lanternFSReleaseAdapter) Stage(distDir string) (string, error) {
	if err := os.MkdirAll(l.releasesDir, 0755); nil != err {
		return "", err
	}

	generation := filepath.Join(l.releasesDir, lanternGenerationPrefix+time.Now().UTC().Format(lanternGenerationStamp))

	if err := os.Rename(distDir, generation); nil != err {
		return "", err
	}

	return generation, nil
}

// Swap re-points the live link at generationDir atomically: a temp symlink is
// renamed over the live one, so readers always see either the old target or
// the new — never a missing link. A failure leaves the previous target live.
func (l lanternFSReleaseAdapter) Swap(generationDir string) error {
	temp := l.liveLink + ".next"

	// clear any leftover from an interrupted earlier swap
	os.Remove(temp)

	if err := os.Symlink(generationDir, temp); nil != err {
		return err
	}

	return os.Rename(temp, l.liveLink)
}

// Previous returns the generation immediately older than the live one, or ""
// when the live link doesn't exist, points outside the kept generations, or
// already points at the oldest.
func (l lanternFSReleaseAdapter) Previous() (string, error) {
	current, err := os.Readlink(l.liveLink)

	if nil != err {
		// no live link yet means nothing to roll back from — not an error
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}

		return "", err
	}

	current = filepath.Clean(current)

	entries, err := os.ReadDir(l.releasesDir)

	if nil != err {
		return "", err
	}

	var generations []string

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), lanternGenerationPrefix) {
			generations = append(generations, entry.Name())
		}
	}

	sort.Strings(generations) // fixed-width stamps: string sort == chronological

	for i, name := range generations {
		if filepath.Clean(filepath.Join(l.releasesDir, name)) == current {
			if 0 == i {
				return "", nil
			}

			return filepath.Join(l.releasesDir, generations[i-1]), nil
		}
	}

	return "", nil
}

// Prune removes all but the keep newest generations. The generation the live
// link points at is never deleted, whatever its age.
func (l lanternFSReleaseAdapter) Prune(keep int) error {
	entries, err := os.ReadDir(l.releasesDir)

	if nil != err {
		return err
	}

	// empty when the live link doesn't exist yet — then nothing is protected
	current, _ := os.Readlink(l.liveLink)
	current = filepath.Clean(current)

	var generations []string

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), lanternGenerationPrefix) {
			generations = append(generations, entry.Name())
		}
	}

	sort.Strings(generations) // fixed-width stamps: string sort == chronological

	if keep < 1 {
		keep = 1
	}

	for i := 0; i < len(generations)-keep; i++ {
		full := filepath.Join(l.releasesDir, generations[i])

		if filepath.Clean(full) == current {
			continue
		}

		if err := os.RemoveAll(full); nil != err {
			return err
		}
	}

	return nil
}
