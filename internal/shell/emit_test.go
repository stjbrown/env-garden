package shell

import (
	"strings"
	"testing"
)

func TestEmitUseOrderAndTracking(t *testing.T) {
	code := EmitUse("VERTEX_ONE VERTEX_TWO", "bedrock", []KV{
		{Key: "CLAUDE_CODE_USE_BEDROCK", Value: "1"},
		{Key: "AWS_REGION", Value: "us-east-1"},
	})

	// Old vars are unset before new exports.
	unsetIdx := strings.Index(code, "unset VERTEX_ONE VERTEX_TWO 2>/dev/null")
	exportIdx := strings.Index(code, "export CLAUDE_CODE_USE_BEDROCK=")
	if unsetIdx < 0 || exportIdx < 0 {
		t.Fatalf("missing unset or export line:\n%s", code)
	}
	if unsetIdx > exportIdx {
		t.Errorf("unset must precede export:\n%s", code)
	}

	// Managed list is sorted and lists only the new keys.
	if !strings.Contains(code, "export __EG_MANAGED='AWS_REGION CLAUDE_CODE_USE_BEDROCK'") {
		t.Errorf("managed list wrong/unsorted:\n%s", code)
	}
	if !strings.Contains(code, "export EG_ACTIVE='bedrock'") {
		t.Errorf("missing EG_ACTIVE:\n%s", code)
	}
}

func TestEmitUseNoPrev(t *testing.T) {
	code := EmitUse("", "vertex", []KV{{Key: "X", Value: "y"}})
	if strings.Contains(code, "unset ") && !strings.Contains(code, "unset __EG_MANAGED") {
		// the only unset allowed here would be none for managed vars
		t.Errorf("unexpected unset with no prev:\n%s", code)
	}
}

func TestEmitOff(t *testing.T) {
	code := EmitOff("A B C")
	if !strings.Contains(code, "unset A B C 2>/dev/null") {
		t.Errorf("missing managed unset:\n%s", code)
	}
	if !strings.Contains(code, "unset __EG_MANAGED EG_ACTIVE 2>/dev/null") {
		t.Errorf("missing tracking unset:\n%s", code)
	}
}

// A corrupted managed list must not be able to inject shell code: only valid
// identifier tokens survive (as harmless `unset` targets), and every shell
// metacharacter is dropped.
func TestFieldsRejectsNonIdent(t *testing.T) {
	code := EmitOff("GOOD ; rm -rf / BAD-DASH 123abc $(reboot)")
	// Only valid identifiers survive, as harmless unset targets. "rm" is a
	// valid name; the metacharacters (;, -rf, /, -, $( ) are all dropped.
	wantLine := "unset GOOD rm 2>/dev/null"
	if !strings.Contains(code, wantLine+"\n") {
		t.Errorf("managed unset not sanitized to %q:\n%s", wantLine, code)
	}
}
