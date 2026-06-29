---
name: eg
description: Run or test a command against a specific AI/cloud provider profile using env-garden (the `eg` CLI). Use when the user wants to test, smoke-test, or run something against a named provider profile (e.g. bedrock, vertex, myproxy, a customer endpoint), check whether a provider works, or run an agent/script against a particular backend. Triggers include "test the X profile", "run this against bedrock", "does vertex work", "use the myproxy endpoint", "eg exec", "smoke test the provider".
license: MIT
compatibility: Requires the `eg` CLI (env-garden) installed; for op:// secret references, the 1Password CLI (`op`) must be installed and signed in.
---

# Using env-garden (`eg`) from an agent

`eg` switches environments between named profiles (`.env.<name>` in
`~/.config/env-garden`). As an agent you cannot change the user's interactive
shell, so use the subprocess and diagnostic verbs only.

## Discover what's available

```
eg list        # profiles (with descriptions)
eg recipes     # built-in templates for `eg add`
```

## Run a command against a profile (preferred)

`eg exec` injects the profile's environment into a subprocess. Secrets (1Password
`op://` references) are resolved in-memory and never written to disk. The user's
shell is untouched.

```
eg exec <profile> -- <command> [args...]
# e.g.
eg exec myproxy -- python agent.py
eg exec bedrock -- claude -p "hello"
```

## Smoke-test connectivity

```
eg doctor <profile>          # sends a tiny real request; exits nonzero on failure
eg doctor <profile> --insecure   # only for self-signed corporate proxies
```

## Rules

- NEVER run `eg use` — it emits shell code meant for the user's interactive shell
  (via the `eg` shell function) and does nothing useful from a child process.
- NEVER run `eg render --resolve` or `--force-resolve` unless the user explicitly
  asks — those write real secret values to a file on disk.
- Treat any resolved values as secrets: do not print, log, or echo them.
- If a command fails with a 1Password message ("op not installed" / "not signed
  in"), surface that hint to the user instead of retrying — they must run
  `op signin` or install the 1Password CLI.
- To create a new profile, prefer `eg add <tool> <provider>` with `--param k=v`
  and `--non-interactive` so no prompt blocks you.
