package profile

import "strings"

// Conflict records that a later profile's value for Key replaced an earlier
// one. Winner is the profile whose value survived; Loser is the profile that
// was overridden. Callers surface these as warnings.
type Conflict struct {
	Key    string
	Winner string
	Loser  string
}

// Merge combines profiles in argument order into a single profile. Variables
// keep their first-seen position, but a later profile's value for an existing
// key overrides the earlier one (last wins) — every override is reported in the
// returned Conflict slice, in the order it occurred.
//
// name becomes the merged profile's Name (used for EG_ACTIVE and generated-file
// headers); Path is left empty since the result is synthetic. The description is
// the joined non-empty descriptions of the inputs.
func Merge(name string, ps []*Profile) (*Profile, []Conflict) {
	merged := &Profile{Name: name}
	// pos maps a key to its index in merged.Vars, so an override updates the
	// value in place rather than appending a duplicate.
	pos := map[string]int{}
	// src tracks which profile currently owns each key, for conflict reporting.
	src := map[string]string{}

	var conflicts []Conflict
	var descs []string
	for _, p := range ps {
		if p.Desc != "" {
			descs = append(descs, p.Desc)
		}
		for _, v := range p.Vars {
			if i, ok := pos[v.Key]; ok {
				conflicts = append(conflicts, Conflict{Key: v.Key, Winner: p.Name, Loser: src[v.Key]})
				merged.Vars[i] = v
				src[v.Key] = p.Name
				continue
			}
			pos[v.Key] = len(merged.Vars)
			src[v.Key] = p.Name
			merged.Vars = append(merged.Vars, v)
		}
	}
	merged.Desc = strings.Join(descs, " + ")
	return merged, conflicts
}
