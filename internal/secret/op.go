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

// SignedIn reports whether an `op` session is usable.
func SignedIn() bool {
	if !Available() {
		return false
	}
	return exec.Command("op", "whoami").Run() == nil
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

// ErrUnavailable describes why op-backed resolution cannot proceed, or returns
// nil if it can. The hint is suitable for printing to the user.
func ErrUnavailable() error {
	if !Available() {
		return fmt.Errorf("this profile uses op:// references but the 1Password CLI is not installed (brew install --cask 1password-cli)")
	}
	if !SignedIn() {
		return fmt.Errorf("this profile uses op:// references but you are not signed in to 1Password (run: op signin, or enable the CLI in the 1Password app)")
	}
	return nil
}
