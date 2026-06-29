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

## Install `eg` if it's missing

This skill may be installed before the `eg` CLI itself. If `eg` is not on `PATH`
(`command -v eg` fails), install it before running anything else:

```sh
# macOS / Linux with Homebrew (preferred)
brew install stjbrown/tap/eg

# no Homebrew — official install script
curl -fsSL https://raw.githubusercontent.com/stjbrown/env-garden/main/install.sh | sh

# from source (needs Go)
go install github.com/stjbrown/env-garden/cmd/eg@latest
```

Prefer Homebrew when `brew` is available; fall back to the install script
otherwise. After installing, verify with `eg --version`. Don't try to run `eg`
commands until one of these succeeds.

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

To combine several profiles, list them before the `--`; they merge left-to-right
and a later profile's value wins on any key conflict (overrides are noted on
stderr). The `--` separator is required when passing more than one profile:

```
eg exec dev-vertex zscaler slack -- python agent.py
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
