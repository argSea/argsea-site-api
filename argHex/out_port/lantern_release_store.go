package out_port

// ReleaseStore owns the on-disk release layout for the lantern: staging a
// freshly built dist directory into a new timestamped generation, atomically
// pointing the live link at a generation, and pruning old generations.
type ReleaseStore interface {
	// Stage moves distDir into a new generation and returns the generation's
	// absolute path.
	Stage(distDir string) (string, error)
	// Swap atomically re-points the live link at generationDir. The previously
	// linked generation stays on disk for instant rollback.
	Swap(generationDir string) error
	// Prune deletes all but the keep newest generations, never touching the one
	// the live link currently points at.
	Prune(keep int) error
	// Previous returns the kept generation immediately older than the one the
	// live link points at, or "" when there is nothing to roll back to.
	Previous() (string, error)
}
