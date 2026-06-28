package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

const testPrompt = "Say hi in 5 words or less."

func probeBedrock(env map[string]string, model string) Result {
	res := Result{Provider: "bedrock"}
	if model == "" {
		model = "us.anthropic.claude-sonnet-4-6"
	}
	if !inPath("aws") {
		res.Detail = "aws CLI not found (required for Bedrock)"
		return res
	}
	region := get(env, "AWS_REGION", "us-east-1")
	profile := get(env, "AWS_PROFILE", "default")

	body := `{"anthropic_version":"bedrock-2023-05-31","max_tokens":32,` +
		`"messages":[{"role":"user","content":"` + testPrompt + `"}]}`
	in, err := os.CreateTemp("", "eg-bedrock-in-*.json")
	if err != nil {
		res.Detail = err.Error()
		return res
	}
	defer os.Remove(in.Name())
	in.WriteString(body)
	in.Close()
	out, err := os.CreateTemp("", "eg-bedrock-out-*.json")
	if err != nil {
		res.Detail = err.Error()
		return res
	}
	out.Close()
	defer os.Remove(out.Name())

	cmd := exec.Command("aws", "bedrock-runtime", "invoke-model",
		"--profile", profile, "--region", region, "--model-id", model,
		"--body", "fileb://"+in.Name(),
		"--content-type", "application/json", "--accept", "application/json",
		out.Name())
	cmd.Env = childEnv(env)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		res.Detail = strings.TrimSpace(stderr.String())
		if res.Detail == "" {
			res.Detail = err.Error()
		}
		res.Detail += "\n    hint: aws sso login --profile " + profile + " (and check the model id for your region)"
		return res
	}

	data, err := os.ReadFile(out.Name())
	if err != nil {
		res.Detail = err.Error()
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
