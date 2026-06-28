package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func probeAnthropic(env map[string]string, model string, insecure bool) Result {
	res := Result{Provider: "anthropic"}
	base := strings.TrimRight(env["ANTHROPIC_BASE_URL"], "/")
	setAuth, ok := anthropicAuth(env)
	if !ok {
		res.Detail = "no credential set (need ANTHROPIC_AUTH_TOKEN or ANTHROPIC_API_KEY)"
		return res
	}
	client := httpClient(insecure)

	// Pick a model: explicit flag, else discover from /v1/models.
	if model == "" {
		model = discoverAnthropicModel(client, base, setAuth)
	}
	if model == "" {
		res.Detail = "no model specified and none discovered from /v1/models"
		return res
	}

	body := fmt.Sprintf(`{"model":%q,"max_tokens":32,"messages":[{"role":"user","content":%q}]}`, model, testPrompt)
	req, _ := http.NewRequest(http.MethodPost, base+"/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	setAuth(req)
	resp, err := client.Do(req)
	if err != nil {
		res.Detail = err.Error()
		return res
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		res.Detail = fmt.Sprintf("HTTP %d (model %s): %s", resp.StatusCode, model, strings.TrimSpace(string(data)))
		return res
	}
	var reply anthropicReply
	if err := json.Unmarshal(data, &reply); err != nil {
		res.Detail = "could not parse response: " + err.Error()
		return res
	}
	res.OK = true
	res.Reply = reply.text()
	return res
}

// anthropicAuth returns a function that sets auth headers. To be robust across
// proxies/gateways (which variously check x-api-key or Authorization: Bearer for
// different endpoints — e.g. some want Bearer on /v1/models but x-api-key on
// /v1/messages), it sets whichever it can: x-api-key from ANTHROPIC_API_KEY, and
// Authorization: Bearer from ANTHROPIC_AUTH_TOKEN (falling back to the api key).
func anthropicAuth(env map[string]string) (func(*http.Request), bool) {
	key := env["ANTHROPIC_API_KEY"]
	tok := env["ANTHROPIC_AUTH_TOKEN"]
	if key == "" && tok == "" {
		return nil, false
	}
	bearer := tok
	if bearer == "" {
		bearer = key
	}
	return func(r *http.Request) {
		if key != "" {
			r.Header.Set("x-api-key", key)
		}
		if bearer != "" {
			r.Header.Set("Authorization", "Bearer "+bearer)
		}
	}, true
}

// discoverAnthropicModel lists models and returns the first containing "claude"
// (or the first of any), best-effort.
func discoverAnthropicModel(client *http.Client, base string, setAuth func(*http.Request)) string {
	req, _ := http.NewRequest(http.MethodGet, base+"/v1/models", nil)
	setAuth(req)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	data, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data []struct {
			ID        string `json:"id"`
			ModelName string `json:"model_name"`
		} `json:"data"`
		Models []struct {
			ID        string `json:"id"`
			ModelName string `json:"model_name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ""
	}
	ids := []string{}
	for _, m := range parsed.Data {
		ids = append(ids, firstNonEmpty(m.ID, m.ModelName))
	}
	for _, m := range parsed.Models {
		ids = append(ids, firstNonEmpty(m.ID, m.ModelName))
	}
	for _, id := range ids {
		if strings.Contains(strings.ToLower(id), "claude") {
			return id
		}
	}
	if len(ids) > 0 {
		return ids[0]
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
