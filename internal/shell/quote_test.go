package shell

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSingleQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "''"},
		{"plain", "'plain'"},
		{"with space", "'with space'"},
		{"it's", `'it'\''s'`},
		{"$HOME", "'$HOME'"},
		{"`whoami`", "'`whoami`'"},
		{`"x"`, `'"x"'`},
	}
	for _, c := range cases {
		if got := SingleQuote(c.in); got != c.want {
			t.Errorf("SingleQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSingleQuoteRoundTrip verifies that adversarial values survive a real
// shell round-trip with their bytes intact — no command substitution, no
// expansion, no splitting.
func TestSingleQuoteRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	values := []string{
		"",
		"simple",
		"a b c",
		"it's a 'test'",
		"$(rm -rf /)",
		"`reboot`",
		"$HOME and ${PATH}",
		"line1\nline2",
		"semi; colon && echo pwned",
		"-leading-dash",
		"emoji 🌱 unicode",
		`mixed "double" and 'single'`,
	}
	for _, v := range values {
		// printf '%s' <quoted> should reproduce v exactly.
		out, err := exec.Command("sh", "-c", "printf '%s' "+SingleQuote(v)).Output()
		if err != nil {
			t.Fatalf("sh failed for %q: %v", v, err)
		}
		if string(out) != v {
			t.Errorf("round-trip mismatch: got %q, want %q", string(out), v)
		}
	}
}

// TestEmittedUseIsValidShell verifies emitted `use` code parses as valid shell.
func TestEmittedUseIsValidShell(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	code := EmitUse("OLD_A OLD_B", "weird-name", []KV{
		{Key: "FOO", Value: "$(rm -rf /)"},
		{Key: "BAR", Value: "it's fine"},
	})
	cmd := exec.Command("sh", "-n")
	cmd.Stdin = strings.NewReader(code)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sh -n rejected emitted code: %v\n%s\n--- code ---\n%s", err, out, code)
	}
}
