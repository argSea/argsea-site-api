package out_adapter

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// lanternWaitDelay is the backstop after a timeout kill: if any descendant
// survives and holds the output pipe open, stop waiting on it after this long
// and return anyway instead of hanging the hoist forever.
const lanternWaitDelay = 5 * time.Second

type lanternExecAdapter struct {
}

// NewLanternExecAdapter returns the real BuildRunner: it execs the argv
// directly — no shell ever sees the command line — with a hard timeout.
func NewLanternExecAdapter() out_port.BuildRunner {
	return lanternExecAdapter{}
}

// Run executes argv in dir with env (KEY=VALUE entries) merged over the
// process environment and returns the combined stdout+stderr. The output never
// includes the environment itself, so nothing secret can leak into the status
// payload.
func (l lanternExecAdapter) Run(dir string, argv []string, env []string, timeout time.Duration) (string, error) {
	if 0 == len(argv) {
		return "", errors.New("build command is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir

	// configured entries win over inherited ones because they come last
	cmd.Env = append(os.Environ(), env...)

	// a build spawns descendants (npm → sh → node) that inherit the output
	// pipe. Run the whole tree in its own process group and kill the group on
	// timeout — killing only argv[0] would leave a grandchild holding the pipe
	// and CombinedOutput blocked until it exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = lanternWaitDelay

	output, err := cmd.CombinedOutput()

	if context.DeadlineExceeded == ctx.Err() {
		return string(output), errors.New("build timed out after " + timeout.String())
	}

	return string(output), err
}
