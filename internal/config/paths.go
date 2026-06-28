// Package config resolves the on-disk locations env-garden uses, following the
// XDG Base Directory specification.
package config

import (
	"os"
	"path/filepath"
	"strings"
)

const appName = "env-garden"

// ProfilePrefix is the leading token of every profile file: ".env.<name>".
const ProfilePrefix = ".env."

// Dir returns the env-garden config directory, honoring $XDG_CONFIG_HOME and
// falling back to ~/.config/env-garden. It does not create the directory.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appName), nil
}

// EnsureDir returns the config directory, creating it (0700) if absent.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// ProfilePath returns the absolute path to the profile file for name.
func ProfilePath(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ProfilePrefix+name), nil
}

// ExpandTilde expands a leading ~ or ~/ to the user's home directory.
func ExpandTilde(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

// NameFromFile extracts the profile name from a ".env.<name>" filename, or ""
// if base is not a profile file.
func NameFromFile(base string) string {
	if !strings.HasPrefix(base, ProfilePrefix) {
		return ""
	}
	return strings.TrimPrefix(base, ProfilePrefix)
}
