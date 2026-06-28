// Package shell emits POSIX shell code (valid in both zsh and bash 3.2) for the
// eg shim to eval, plus the shim itself.
package shell

import "strings"

// SingleQuote wraps s in single quotes, escaping embedded single quotes via the
// classic '\” sequence. This is the only fully shell-agnostic, injection-safe
// way to emit an arbitrary value: inside single quotes $, backticks, ", spaces,
// newlines, and globs are all inert.
//
//	it's $HOME  ->  'it'\''s $HOME'
func SingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
