package out_adapter

import (
	"time"
)

// LanternFakeRunner is a scriptable BuildRunner for tests: it returns the
// configured output/error, optionally blocking on Gate first so a test can
// observe the building state and exercise the single-flight guard.
type LanternFakeRunner struct {
	Output string
	Err    error
	Gate   chan struct{} // when non-nil, Run blocks until the channel closes
}

func (f *LanternFakeRunner) Run(dir string, argv []string, env []string, timeout time.Duration) (string, error) {
	if nil != f.Gate {
		<-f.Gate
	}

	return f.Output, f.Err
}

// LanternFakeStateRepo is an in-memory LanternStateRepo for tests.
type LanternFakeStateRepo struct {
	Stamp string
}

func (f *LanternFakeStateRepo) LastHoistedAt() (string, error) {
	return f.Stamp, nil
}

func (f *LanternFakeStateRepo) SaveLastHoistedAt(stamp string) error {
	f.Stamp = stamp

	return nil
}

// LanternFakeReleaseStore records the pipeline's filesystem calls without
// touching disk, so state-machine tests stay pure.
type LanternFakeReleaseStore struct {
	Staged  []string
	Swapped []string
	Pruned  []int
	Prev    string // returned by Previous
	Err     error  // returned by Stage when set
}

func (f *LanternFakeReleaseStore) Stage(distDir string) (string, error) {
	if nil != f.Err {
		return "", f.Err
	}

	f.Staged = append(f.Staged, distDir)

	return distDir + "-generation", nil
}

func (f *LanternFakeReleaseStore) Swap(generationDir string) error {
	f.Swapped = append(f.Swapped, generationDir)

	return nil
}

func (f *LanternFakeReleaseStore) Prune(keep int) error {
	f.Pruned = append(f.Pruned, keep)

	return nil
}

func (f *LanternFakeReleaseStore) Previous() (string, error) {
	return f.Prev, nil
}
