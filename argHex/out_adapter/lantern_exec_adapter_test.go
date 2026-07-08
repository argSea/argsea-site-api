package out_adapter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/out_adapter"
)

func TestExecRunnerReportsSuccessAndFailure(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	if _, err := runner.Run(t.TempDir(), []string{"true"}, nil, time.Second); nil != err {
		t.Fatalf("expected a zero exit to succeed, got %v", err)
	}

	if _, err := runner.Run(t.TempDir(), []string{"false"}, nil, time.Second); nil == err {
		t.Fatal("expected a nonzero exit to error")
	}
}

func TestExecRunnerCapturesOutput(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	output, err := runner.Run(t.TempDir(), []string{"echo", "ahoy"}, nil, time.Second)

	if nil != err || !strings.Contains(output, "ahoy") {
		t.Fatalf("expected the command output captured, got %q (%v)", output, err)
	}
}

func TestExecRunnerMergesConfiguredEnv(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	env := []string{"LANTERN_TEST_SIGNAL=harbor"}
	output, err := runner.Run(t.TempDir(), []string{"printenv", "LANTERN_TEST_SIGNAL"}, env, time.Second)

	if nil != err || !strings.Contains(output, "harbor") {
		t.Fatalf("expected the configured env visible to the build, got %q (%v)", output, err)
	}
}

func TestExecRunnerEnforcesTimeout(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	start := time.Now()
	_, err := runner.Run(t.TempDir(), []string{"sleep", "5"}, nil, 100*time.Millisecond)

	if nil == err || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected a timeout error, got %v", err)
	}

	if time.Since(start) > 2*time.Second {
		t.Fatal("the timeout did not actually kill the command")
	}
}

// A real build (npm → sh → node) spawns descendants that inherit the output
// pipe. Killing only argv[0] would leave the backgrounded child holding the
// pipe and CombinedOutput blocked for its full 30s; the process-group kill
// must take the whole tree down promptly.
func TestExecRunnerTimeoutKillsDescendants(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	start := time.Now()
	_, err := runner.Run(t.TempDir(), []string{"sh", "-c", "sleep 30 & sleep 30"}, nil, 200*time.Millisecond)

	if nil == err || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected a timeout error, got %v", err)
	}

	// well under the WaitDelay backstop: this proves the group kill worked,
	// not just that the pipe wait eventually gave up
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("runner blocked on an orphaned descendant for %v after the timeout", elapsed)
	}
}

func TestExecRunnerRejectsEmptyCommand(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	if _, err := runner.Run(t.TempDir(), nil, nil, time.Second); nil == err {
		t.Fatal("expected an empty argv to be rejected")
	}
}
