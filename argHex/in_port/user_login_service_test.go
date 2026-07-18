package in_port_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/in_port"
)

// TestErrLoginBarredMessagePinsTheKeeperLine locks the exact ratified line: the
// admin console couples on the word "barred" in this string, so a rewrite here
// is a cross-repo break, not a wording tweak.
func TestErrLoginBarredMessagePinsTheKeeperLine(t *testing.T) {
	const wantLine = "the door is barred for the night. come back with the tide."

	if wantLine != in_port.ErrLoginBarred.Error() {
		t.Fatalf("ErrLoginBarred must pin the ratified keeper line, got %q", in_port.ErrLoginBarred.Error())
	}
}
