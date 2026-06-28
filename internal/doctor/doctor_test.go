package doctor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"CLAUDE_CODE_USE_BEDROCK": "1"}, "bedrock"},
		{map[string]string{"CLAUDE_CODE_USE_VERTEX": "1"}, "vertex"},
		{map[string]string{"ANTHROPIC_VERTEX_PROJECT_ID": "p"}, "vertex"},
		{map[string]string{"ANTHROPIC_BASE_URL": "https://x"}, "anthropic"},
		{map[string]string{"OPENAI_BASE_URL": "https://x"}, "openai"},
		{map[string]string{}, ""},
	}
	for _, c := range cases {
		if got := detect(c.env); got != c.want {
			t.Errorf("detect(%v) = %q, want %q", c.env, got, c.want)
		}
	}
}

func TestProbeAnthropicSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("x-api-key") != "secret" {
				w.WriteHeader(401)
				return
			}
			w.Write([]byte(`{"data":[{"id":"claude-sonnet-4-6"},{"id":"gpt-4"}]}`))
		case "/v1/messages":
			w.Write([]byte(`{"content":[{"text":"Hi there friend"}]}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	res := Run(map[string]string{
		"ANTHROPIC_BASE_URL": srv.URL,
		"ANTHROPIC_API_KEY":  "secret",
	}, "", false)
	if !res.OK || res.Reply != "Hi there friend" {
		t.Fatalf("got %+v", res)
	}
	if res.Provider != "anthropic" {
		t.Errorf("provider = %q", res.Provider)
	}
}

func TestProbeAnthropicAuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	res := Run(map[string]string{
		"ANTHROPIC_BASE_URL":   srv.URL,
		"ANTHROPIC_AUTH_TOKEN": "bad",
	}, "explicit-model", false)
	if res.OK {
		t.Fatalf("expected failure, got %+v", res)
	}
}

func TestProbeOpenAI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" && r.Header.Get("Authorization") == "Bearer k" {
			w.Write([]byte(`{"data":[]}`))
			return
		}
		w.WriteHeader(403)
	}))
	defer srv.Close()

	res := Run(map[string]string{"OPENAI_BASE_URL": srv.URL, "OPENAI_API_KEY": "k"}, "", false)
	if !res.OK || res.Provider != "openai" {
		t.Fatalf("got %+v", res)
	}
}

func TestNoProvider(t *testing.T) {
	if res := Run(map[string]string{}, "", false); res.OK {
		t.Errorf("expected failure for empty env")
	}
}
