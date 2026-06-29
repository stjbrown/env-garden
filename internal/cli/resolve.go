package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/secret"
	"github.com/stjbrown/env-garden/internal/shell"
)

// loadMerged loads each named profile and merges them into one (last wins on
// key collisions). With a single name it returns that profile unchanged. Any
// override is reported to stderr so a clobbered variable is never silent; stdout
// is left untouched, keeping `use`/`export` output eval-clean.
//
// The merged profile's Name is the names joined with "+", which flows into
// EG_ACTIVE and generated-file headers.
func loadMerged(names []string) (*profile.Profile, error) {
	ps := make([]*profile.Profile, 0, len(names))
	for _, name := range names {
		p, err := profile.Load(name)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("no profile %q (try: eg list)", name)
			}
			return nil, err
		}
		ps = append(ps, p)
	}
	if len(ps) == 1 {
		return ps[0], nil
	}
	merged, conflicts := profile.Merge(strings.Join(names, "+"), ps)
	for _, c := range conflicts {
		fmt.Fprintf(os.Stderr, "eg: %s from %q overrides %q\n", c.Key, c.Winner, c.Loser)
	}
	return merged, nil
}

// profileExists reports whether a profile of the given name exists, treating any
// lookup error as "no" (it is only used for a best-effort usage hint).
func profileExists(name string) bool {
	ok, err := profile.Exists(name)
	return err == nil && ok
}

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
