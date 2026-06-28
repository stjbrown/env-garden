package toolconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestApplyCodexFresh(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	if _, err := ApplyCodex(cfg, CodexParams{ProviderID: "myproxy", BaseURL: "https://z/v1", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	assertProvider(t, cfg, "myproxy", "https://z/v1")

	overlay := filepath.Join(dir, "myproxy.config.toml")
	var ov map[string]any
	mustParse(t, overlay, &ov)
	if ov["model_provider"] != "myproxy" || ov["model"] != "m" {
		t.Errorf("overlay wrong: %+v", ov)
	}
}

// The critical regression: appending to a config that already has a provider
// must remain valid TOML with BOTH providers intact.
func TestApplyCodexPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	seed := "# hand-written\nmodel = \"gpt-5\"\n\n[model_providers.openrouter]\nname = \"OpenRouter\"\nbase_url = \"https://openrouter.ai/api/v1\"\n"
	if err := os.WriteFile(cfg, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyCodex(cfg, CodexParams{ProviderID: "acme", BaseURL: "https://acme/v1", Model: "big"}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(cfg)
	if !strings.Contains(string(data), "# hand-written") {
		t.Error("lost original comment")
	}
	assertProvider(t, cfg, "openrouter", "https://openrouter.ai/api/v1")
	assertProvider(t, cfg, "acme", "https://acme/v1")
}

func TestApplyCodexIdempotent(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	p := CodexParams{ProviderID: "myproxy", BaseURL: "https://z/v1", Model: "m"}
	if _, err := ApplyCodex(cfg, p); err != nil {
		t.Fatal(err)
	}
	// Second apply with a different base_url must NOT duplicate or change the block.
	p.BaseURL = "https://changed/v1"
	if _, err := ApplyCodex(cfg, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(cfg)
	if n := strings.Count(string(data), "[model_providers.myproxy]"); n != 1 {
		t.Errorf("expected exactly one myproxy block, got %d", n)
	}
	assertProvider(t, cfg, "myproxy", "https://z/v1") // original kept
}

func TestApplyCodexRejectsReserved(t *testing.T) {
	if _, err := ApplyCodex(filepath.Join(t.TempDir(), "c.toml"), CodexParams{ProviderID: "openai", BaseURL: "x"}); err == nil {
		t.Error("expected reserved id rejection")
	}
}

func assertProvider(t *testing.T, path, id, wantBase string) {
	t.Helper()
	var doc map[string]any
	mustParse(t, path, &doc)
	mp, ok := doc["model_providers"].(map[string]any)
	if !ok {
		t.Fatalf("no model_providers table in %s", path)
	}
	prov, ok := mp[id].(map[string]any)
	if !ok {
		t.Fatalf("no provider %q in %s", id, path)
	}
	if prov["base_url"] != wantBase {
		t.Errorf("provider %q base_url = %v, want %v", id, prov["base_url"], wantBase)
	}
}

func mustParse(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := toml.Unmarshal(data, v); err != nil {
		t.Fatalf("invalid TOML in %s: %v\n%s", path, err, data)
	}
}
