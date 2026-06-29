package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if got := ReadDefault(); got != "" {
		t.Errorf("fresh default = %q, want empty", got)
	}
	if err := WriteDefault("vertex"); err != nil {
		t.Fatal(err)
	}
	if got := ReadDefault(); got != "vertex" {
		t.Errorf("ReadDefault = %q, want vertex", got)
	}
	// stored under the config dir
	dp, _ := DefaultPath()
	if filepath.Base(dp) != "default" {
		t.Errorf("DefaultPath base = %q", filepath.Base(dp))
	}
	if err := ClearDefault(); err != nil {
		t.Fatal(err)
	}
	if got := ReadDefault(); got != "" {
		t.Errorf("after clear = %q, want empty", got)
	}
	// clearing again is not an error
	if err := ClearDefault(); err != nil {
		t.Errorf("clear when absent: %v", err)
	}
}
