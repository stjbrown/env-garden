package profile

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/stjbrown/env-garden/internal/config"
)

// Summary is a lightweight profile listing entry (no variables parsed beyond
// the description header).
type Summary struct {
	Name string
	Desc string
	Path string
}

// List enumerates profiles in the config directory, sorted by name. A missing
// directory yields an empty list, not an error.
func List() ([]Summary, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []Summary
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := config.NameFromFile(e.Name())
		if name == "" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		out = append(out, Summary{Name: name, Desc: descOf(path), Path: path})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Load parses the profile named name from the config directory.
func Load(name string) (*Profile, error) {
	path, err := config.ProfilePath(name)
	if err != nil {
		return nil, err
	}
	return Parse(path, name)
}

// Exists reports whether a profile file exists for name.
func Exists(name string) (bool, error) {
	path, err := config.ProfilePath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// descOf reads just the description header without a full parse; best-effort.
func descOf(path string) string {
	p, err := Parse(path, "")
	if err != nil {
		return ""
	}
	return p.Desc
}
