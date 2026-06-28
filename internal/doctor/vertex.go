package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func probeVertex(env map[string]string, model string, insecure bool) Result {
	res := Result{Provider: "vertex"}
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	project := env["ANTHROPIC_VERTEX_PROJECT_ID"]
	if project == "" {
		res.Detail = "ANTHROPIC_VERTEX_PROJECT_ID is not set"
		return res
	}
	if !inPath("gcloud") {
		res.Detail = "gcloud CLI not found (required for Vertex)"
		return res
	}
	region := get(env, "CLOUD_ML_REGION", "us-east5")
	if region == "global" {
		// The Anthropic REST endpoint isn't served on the global region.
		region = "us-east5"
	}

	tokenCmd := exec.Command("gcloud", "auth", "print-access-token")
	tokenCmd.Env = childEnv(env)
	tokenOut, err := tokenCmd.Output()
	if err != nil {
		res.Detail = "could not get gcloud access token (try: gcloud auth login)"
		return res
	}
	token := strings.TrimSpace(string(tokenOut))

	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:rawPredict",
		region, project, region, model)
	body := `{"anthropic_version":"vertex-2023-10-16","max_tokens":32,` +
		`"messages":[{"role":"user","content":"` + testPrompt + `"}]}`

	req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient(insecure).Do(req)
	if err != nil {
		res.Detail = err.Error()
		return res
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		res.Detail = fmt.Sprintf("HTTP %d (region %s): %s", resp.StatusCode, region, strings.TrimSpace(string(data)))
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
