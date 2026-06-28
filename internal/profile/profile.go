// Package profile parses and represents env-garden profiles: ordered sets of
// environment variables stored as ".env.<name>" files, supporting literal
// values, $VAR interpolation, and 1Password op:// secret references.
package profile

import "strings"

// Var is a single environment variable from a profile, in file order.
type Var struct {
	Key string
	// Raw is the post-expansion literal value, or—when IsRef is true—the raw
	// op:// reference text (never expanded, never resolved at parse time).
	Raw   string
	IsRef bool
}

// Profile is a parsed profile file.
type Profile struct {
	Name string
	Desc string
	Path string
	Vars []Var
}

// HasRefs reports whether any variable is a 1Password reference.
func (p *Profile) HasRefs() bool {
	for _, v := range p.Vars {
		if v.IsRef {
			return true
		}
	}
	return false
}

// Keys returns the variable names in file order.
func (p *Profile) Keys() []string {
	keys := make([]string, len(p.Vars))
	for i, v := range p.Vars {
		keys[i] = v.Key
	}
	return keys
}

// isRefValue reports whether a value carries a 1Password reference.
func isRefValue(s string) bool {
	return strings.Contains(s, "op://")
}
