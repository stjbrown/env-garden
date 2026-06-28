package recipe

import (
	"strings"
	"testing"
)

func TestSubstitute(t *testing.T) {
	got, err := Substitute("{{a}}/v1 and {{b}}", map[string]string{"a": "http://x", "b": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://x/v1 and y" {
		t.Errorf("got %q", got)
	}
	if _, err := Substitute("{{missing}}", map[string]string{}); err == nil {
		t.Error("expected error for unknown placeholder")
	}
}

func TestRenderProfile(t *testing.T) {
	r := Recipe{
		Tool: "claude-code", Provider: "bedrock", Description: "Bedrock",
		Env: []EnvVar{{Key: "AWS_PROFILE", Value: "{{p}}"}, {Key: "X", Value: "1"}},
	}
	out, err := RenderProfile(r, map[string]string{"p": "dev"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# desc: Bedrock", "export AWS_PROFILE=dev", "export X=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// TestCatalogLoads guards every embedded recipe: they must parse and validate.
func TestCatalogLoads(t *testing.T) {
	all, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("empty catalog")
	}
	for _, r := range all {
		if _, err := Lookup(r.Tool, r.Provider); err != nil {
			t.Errorf("lookup %s: %v", r.Key(), err)
		}
	}
}
