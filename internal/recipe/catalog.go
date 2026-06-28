package recipe

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"

	toml "github.com/pelletier/go-toml/v2"
)

//go:embed data/*.toml
var catalogFS embed.FS

// Catalog loads and returns all embedded recipes, sorted by key.
func Catalog() ([]Recipe, error) {
	entries, err := fs.ReadDir(catalogFS, "data")
	if err != nil {
		return nil, err
	}
	var out []Recipe
	for _, e := range entries {
		b, err := catalogFS.ReadFile("data/" + e.Name())
		if err != nil {
			return nil, err
		}
		var r Recipe
		if err := toml.Unmarshal(b, &r); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if err := r.validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out, nil
}

// Lookup finds a recipe by tool and provider.
func Lookup(tool, provider string) (Recipe, error) {
	all, err := Catalog()
	if err != nil {
		return Recipe{}, err
	}
	for _, r := range all {
		if r.Tool == tool && r.Provider == provider {
			return r, nil
		}
	}
	return Recipe{}, fmt.Errorf("no recipe %s/%s (run: eg recipes)", tool, provider)
}
