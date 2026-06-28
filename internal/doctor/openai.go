package doctor

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// probeOpenAI does a connectivity check against an OpenAI-compatible endpoint by
// listing models. (Codex itself speaks the Responses API; a models listing is a
// lightweight reachability/auth check.)
func probeOpenAI(env map[string]string, insecure bool) Result {
	res := Result{Provider: "openai"}
	base := strings.TrimRight(env["OPENAI_BASE_URL"], "/")
	key := env["OPENAI_API_KEY"]
	if base == "" {
		res.Detail = "OPENAI_BASE_URL is not set"
		return res
	}

	req, _ := http.NewRequest(http.MethodGet, base+"/v1/models", nil)
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := httpClient(insecure).Do(req)
	if err != nil {
		res.Detail = err.Error()
		return res
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		res.Detail = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
		return res
	}
	res.OK = true
	res.Reply = "endpoint reachable, /v1/models OK"
	return res
}
