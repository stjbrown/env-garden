// Package secret resolves 1Password op:// references via the `op` CLI, with
// graceful degradation when op is absent or not signed in.
package secret

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Ref reports whether a value carries a 1Password reference.
func Ref(value string) bool {
	return strings.Contains(value, "op://")
}

// Available reports whether the `op` CLI is installed.
func Available() bool {
	_, err := exec.LookPath("op")
	return err == nil
}

// Read resolves a single op:// reference to its value.
func Read(ref string) (string, error) {
	var out, errBuf bytes.Buffer
	cmd := exec.Command("op", "read", "--no-newline", ref)
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("op read %s: %s", ref, msg)
	}
	return out.String(), nil
}

// ErrUnavailable reports whether op-backed resolution is possible at all. It
// only checks that the CLI is installed — auth/session state is intentionally
// NOT pre-checked here, because `op whoami` reports "not signed in" under the
// 1Password desktop app integration even when `op read` works fine. The actual
// read is the source of truth: Read surfaces op's own error (which carries the
// sign-in hint) when a reference can't be resolved.
func ErrUnavailable() error {
	if !Available() {
		return fmt.Errorf("this profile uses op:// references but the 1Password CLI is not installed (brew install --cask 1password-cli)")
	}
	return nil
}
