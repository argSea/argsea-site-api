package out_adapter

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/argSea/argsea-site-api/argHex/out_port"
)

type lanternExecAdapter struct {
}

// NewLanternExecAdapter returns the real BuildRunner: it execs the argv
// directly — no shell ever sees the command line — with a hard timeout.
func NewLanternExecAdapter() out_port.BuildRunner {
	return lanternExecAdapter{}
}

// Run executes argv in dir with env merged over the process environment and
// returns the combined stdout+stderr. The output never includes the
// environment itself, so nothing secret can leak into the status payload.
func (l lanternExecAdapter) Run(dir string, argv []string, env map[string]string, timeout time.Duration) (string, error) {
	if 0 == len(argv) {
		return "", errors.New("build command is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir

	// configured entries win over inherited ones because they come last
	merged := os.Environ()

	for key, value := range env {
		merged = append(merged, key+"="+value)
	}

	cmd.Env = merged

	output, err := cmd.CombinedOutput()

	if context.DeadlineExceeded == ctx.Err() {
		return string(output), errors.New("build timed out after " + timeout.String())
	}

	return string(output), err
}
