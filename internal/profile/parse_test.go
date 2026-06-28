package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.test")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseLiteralsAndExport(t *testing.T) {
	p, err := Parse(writeTemp(t, "# desc: My profile\nexport FOO=bar\nBAZ=qux\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Desc != "My profile" {
		t.Errorf("desc = %q", p.Desc)
	}
	want := []Var{{Key: "FOO", Raw: "bar"}, {Key: "BAZ", Raw: "qux"}}
	if len(p.Vars) != len(want) {
		t.Fatalf("got %d vars, want %d", len(p.Vars), len(want))
	}
	for i, v := range want {
		if p.Vars[i] != v {
			t.Errorf("var[%d] = %+v, want %+v", i, p.Vars[i], v)
		}
	}
}

// The real myproxy profile relies on $VAR interpolation against an earlier var.
func TestParseInterpolation(t *testing.T) {
	src := "export MYPROXY_BASE_URL=https://myproxy.example\n" +
		"export ANTHROPIC_BASE_URL=$MYPROXY_BASE_URL\n" +
		"export NESTED=${MYPROXY_BASE_URL}/v1\n"
	p, err := Parse(writeTemp(t, src), "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Vars[1].Raw != "https://myproxy.example" {
		t.Errorf("$VAR not expanded: %q", p.Vars[1].Raw)
	}
	if p.Vars[2].Raw != "https://myproxy.example/v1" {
		t.Errorf("${VAR} not expanded: %q", p.Vars[2].Raw)
	}
}

func TestParseOpRefNotExpanded(t *testing.T) {
	p, err := Parse(writeTemp(t, `export ANTHROPIC_API_KEY="op://Private/myproxy/credential"`+"\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasRefs() {
		t.Fatal("expected ref detected")
	}
	v := p.Vars[0]
	if !v.IsRef || v.Raw != "op://Private/myproxy/credential" {
		t.Errorf("ref var = %+v", v)
	}
}

func TestParseSingleQuoteIsLiteral(t *testing.T) {
	p, err := Parse(writeTemp(t, "FOO='$HOME literal'\n"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Vars[0].Raw != "$HOME literal" {
		t.Errorf("single-quoted value should not expand: %q", p.Vars[0].Raw)
	}
}

func TestParseRejectsBadKey(t *testing.T) {
	if _, err := Parse(writeTemp(t, "1BAD=x\n"), "test"); err == nil {
		t.Error("expected error for invalid key")
	}
	if _, err := Parse(writeTemp(t, "no equals here\n"), "test"); err == nil {
		t.Error("expected error for missing =")
	}
}
