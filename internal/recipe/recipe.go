// Package recipe provides the embedded catalog of (tool, provider) templates
// used by `eg add` to scaffold profiles (and, for file-configured tools, to
// patch tool config files).
package recipe

import "fmt"

// Param is a value the user supplies when materializing a recipe.
type Param struct {
	Name    string `toml:"name"`
	Prompt  string `toml:"prompt"`
	Default string `toml:"default"`
	// Secret marks params whose value is a 1Password reference (op://...); used
	// only to tailor the prompt, not to change handling.
	Secret bool `toml:"secret"`
}

// EnvVar is one line of the generated profile. Value may contain {{param}}
// placeholders resolved against the supplied params.
type EnvVar struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

// FileArtifact describes a tool config file the recipe manages (e.g. Codex's
// config.toml). Handled by package toolconfig in Phase 4; declared here so the
// schema is stable.
type FileArtifact struct {
	Tool string `toml:"tool"` // e.g. "codex"
	Path string `toml:"path"` // e.g. "~/.codex/config.toml"
}

// Recipe is a single catalog entry.
type Recipe struct {
	Tool        string         `toml:"tool"`
	Provider    string         `toml:"provider"`
	Description string         `toml:"description"`
	DefaultName string         `toml:"default_name"`
	Notes       string         `toml:"notes"`
	Params      []Param        `toml:"param"`
	Env         []EnvVar       `toml:"env"`
	Files       []FileArtifact `toml:"file"`
}

// Key is the "tool/provider" identifier.
func (r Recipe) Key() string { return r.Tool + "/" + r.Provider }

// validate checks a recipe is well-formed (used at load time).
func (r Recipe) validate() error {
	if r.Tool == "" || r.Provider == "" {
		return fmt.Errorf("recipe missing tool/provider")
	}
	if len(r.Env) == 0 && len(r.Files) == 0 {
		return fmt.Errorf("recipe %s has no env or files", r.Key())
	}
	return nil
}
