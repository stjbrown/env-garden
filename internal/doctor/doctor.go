// Package doctor smoke-tests a resolved profile against its provider by sending
// a tiny real request, mirroring the old test-ai.sh probes in Go.
package doctor

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Result is the outcome of a probe.
type Result struct {
	Provider string // bedrock | vertex | anthropic | openai
	OK       bool
	Reply    string // model reply text on success
	Detail   string // diagnostic detail on failure
}

// Run detects the provider from the resolved env and probes it. model overrides
// the default test model; insecure disables TLS verification (also enabled when
// the profile sets NODE_TLS_REJECT_UNAUTHORIZED=0).
func Run(env map[string]string, model string, insecure bool) Result {
	insecure = insecure || env["NODE_TLS_REJECT_UNAUTHORIZED"] == "0"
	switch detect(env) {
	case "bedrock":
		return probeBedrock(env, model)
	case "vertex":
		return probeVertex(env, model, insecure)
	case "anthropic":
		return probeAnthropic(env, model, insecure)
	case "openai":
		return probeOpenAI(env, insecure)
	default:
		return Result{OK: false, Detail: "could not determine provider from this profile (need CLAUDE_CODE_USE_BEDROCK, CLAUDE_CODE_USE_VERTEX, ANTHROPIC_BASE_URL, or OPENAI_BASE_URL)"}
	}
}

func detect(env map[string]string) string {
	switch {
	case truthy(env["CLAUDE_CODE_USE_BEDROCK"]):
		return "bedrock"
	case truthy(env["CLAUDE_CODE_USE_VERTEX"]) || env["ANTHROPIC_VERTEX_PROJECT_ID"] != "":
		return "vertex"
	case env["ANTHROPIC_BASE_URL"] != "":
		return "anthropic"
	case env["OPENAI_BASE_URL"] != "":
		return "openai"
	default:
		return ""
	}
}

func truthy(s string) bool {
	switch strings.ToLower(s) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func get(env map[string]string, key, def string) string {
	if v := env[key]; v != "" {
		return v
	}
	return def
}

func inPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// childEnv overlays the profile env on the process env for shelling out.
func childEnv(env map[string]string) []string {
	out := os.Environ()
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

func httpClient(insecure bool) *http.Client {
	tr := &http.Transport{}
	if insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in for self-signed corp proxies
	}
	return &http.Client{Timeout: 30 * time.Second, Transport: tr}
}

// anthropicReply is the shared shape of Anthropic Messages responses.
type anthropicReply struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (r anthropicReply) text() string {
	if len(r.Content) > 0 {
		return r.Content[0].Text
	}
	return ""
}
