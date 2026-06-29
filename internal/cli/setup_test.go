package cli

import (
	"strings"
	"testing"
)

func TestReplaceBlock(t *testing.T) {
	newBlock := setupBegin + "\neval \"$(/new/eg init zsh)\"\n" + setupEnd

	// No existing block → no change.
	if _, changed := replaceBlock("export FOO=1\n", newBlock); changed {
		t.Error("expected no change when no block present")
	}

	// Existing identical block → no change.
	text := "top\n" + newBlock + "\nbottom\n"
	if _, changed := replaceBlock(text, newBlock); changed {
		t.Error("expected no change for identical block")
	}

	// Existing different block → replaced, surrounding text preserved.
	old := "top\n" + setupBegin + "\neval \"$(/old/eg init zsh)\"\n" + setupEnd + "\nbottom\n"
	out, changed := replaceBlock(old, newBlock)
	if !changed {
		t.Fatal("expected change for differing block")
	}
	if !strings.Contains(out, "/new/eg") || strings.Contains(out, "/old/eg") {
		t.Errorf("block not swapped:\n%s", out)
	}
	if !strings.HasPrefix(out, "top\n") || !strings.HasSuffix(out, "bottom\n") {
		t.Errorf("surrounding text lost:\n%s", out)
	}
}

func TestDetectShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if detectShell() != "bash" {
		t.Error("expected bash")
	}
	t.Setenv("SHELL", "/usr/local/bin/zsh")
	if detectShell() != "zsh" {
		t.Error("expected zsh")
	}
	t.Setenv("SHELL", "")
	if detectShell() != "zsh" {
		t.Error("expected zsh default")
	}
}
