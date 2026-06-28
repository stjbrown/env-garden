package cli

import (
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/secret"
	"github.com/stjbrown/env-garden/internal/shell"
)

// resolvePairs turns a profile's variables into concrete key/value pairs ready
// to export. Literal values pass through verbatim; op:// references are resolved
// in-memory via the 1Password CLI.
//
// Profiles with no references never touch op. When references are present but op
// is unavailable or signed out, this fails before any value is produced — so
// callers (e.g. the use shim) never partially apply a profile.
func resolvePairs(p *profile.Profile) ([]shell.KV, error) {
	if p.HasRefs() {
		if err := secret.ErrUnavailable(); err != nil {
			return nil, err
		}
	}
	pairs := make([]shell.KV, 0, len(p.Vars))
	for _, v := range p.Vars {
		val := v.Raw
		if v.IsRef {
			resolved, err := secret.Read(v.Raw)
			if err != nil {
				return nil, err
			}
			val = resolved
		}
		pairs = append(pairs, shell.KV{Key: v.Key, Value: val})
	}
	return pairs, nil
}
