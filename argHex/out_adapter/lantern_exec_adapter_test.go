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

	env := map[string]string{"LANTERN_TEST_SIGNAL": "harbor"}
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

func TestExecRunnerRejectsEmptyCommand(t *testing.T) {
	runner := out_adapter.NewLanternExecAdapter()

	if _, err := runner.Run(t.TempDir(), nil, nil, time.Second); nil == err {
		t.Fatal("expected an empty argv to be rejected")
	}
}
