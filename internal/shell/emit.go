package shell

import (
	"sort"
	"strings"
)

// Managed tracking variables exported into the shell so a later `use`/`off` can
// tear down exactly what a prior `use` set, with no global state file.
const (
	ManagedVar = "__EG_MANAGED"
	ActiveVar  = "EG_ACTIVE"
)

// KV is a resolved key/value pair ready to export.
type KV struct {
	Key   string
	Value string
}

// EmitUse returns shell code that unsets the previously-managed variables
// (prev), exports the new pairs, and records the new managed set + active name.
// prev is the space-separated value of $__EG_MANAGED from the calling shell.
//
// Order matters: unset-old happens before export-new, so a variable present in
// both the old and new profile ends up with the new value rather than being
// dropped.
func EmitUse(prev string, name string, pairs []KV) string {
	var b strings.Builder
	emitUnset(&b, fields(prev))

	keys := make([]string, 0, len(pairs))
	for _, p := range pairs {
		b.WriteString("export ")
		b.WriteString(p.Key)
		b.WriteByte('=')
		b.WriteString(SingleQuote(p.Value))
		b.WriteByte('\n')
		keys = append(keys, p.Key)
	}
	sort.Strings(keys)
	b.WriteString("export " + ManagedVar + "=" + SingleQuote(strings.Join(keys, " ")) + "\n")
	b.WriteString("export " + ActiveVar + "=" + SingleQuote(name) + "\n")
	return b.String()
}

// EmitOff returns shell code that unsets the previously-managed variables along
// with the tracking variables themselves.
func EmitOff(prev string) string {
	var b strings.Builder
	emitUnset(&b, fields(prev))
	b.WriteString("unset " + ManagedVar + " " + ActiveVar + " 2>/dev/null\n")
	return b.String()
}

func emitUnset(b *strings.Builder, names []string) {
	if len(names) == 0 {
		return
	}
	b.WriteString("unset")
	for _, n := range names {
		b.WriteByte(' ')
		b.WriteString(n)
	}
	b.WriteString(" 2>/dev/null\n")
}

// fields splits a space-separated managed list, dropping empties. Only valid
// identifier names are kept, so a corrupted $__EG_MANAGED cannot inject code via
// the unset line.
func fields(s string) []string {
	var out []string
	for _, f := range strings.Fields(s) {
		if isIdent(f) {
			out = append(out, f)
		}
	}
	return out
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
