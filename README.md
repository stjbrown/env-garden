# env-garden (`eg`)

A per-shell **environment profile switcher** with batteries. Point your tools at
named profiles and switch them **independently in each terminal** — Claude Code
on Vertex in one window, Bedrock in another, a customer endpoint in a third.

`eg` started as an AI-provider switcher and grew into a general tool: a profile is
just a named set of environment variables (with optional 1Password secrets), and
you can project it onto your **shell**, a **subprocess**, or a **project file**.

```sh
eg use bedrock-dev          # switch THIS shell
eg exec myproxy -- python agent.py   # run a command with a profile's env
eg render staging -o .env   # write a project .env file
eg add claude-code vertex   # scaffold a profile from a recipe
eg doctor bedrock-dev       # smoke-test that it actually works
```

## Why per-shell?

Switching providers is fundamentally an environment-variable problem: env vars are
per-process, so each terminal/tmux pane is naturally independent. Anything that
stores one global "active provider" (a GUI, a daemon, a config file) can't give
you a different provider per window. `eg` is a tiny Go binary plus a shell shim,
so switching happens in *your* shell.

## Install

**Homebrew**

```sh
brew install stjbrown/tap/eg
```

**Script**

```sh
curl -fsSL https://raw.githubusercontent.com/stjbrown/env-garden/main/install.sh | sh
```

**From source**

```sh
go install github.com/stjbrown/env-garden/cmd/eg@latest
```

Then add the shell integration (once):

```sh
echo 'eval "$(eg init zsh)"'  >> ~/.zshrc     # zsh
echo 'eval "$(eg init bash)"' >> ~/.bashrc    # bash
```

The shim is required because only a function sourced into your shell can change
that shell's environment — a plain binary runs in a child process and can't.

## Profiles

Profiles live in `~/.config/env-garden/` (honors `$XDG_CONFIG_HOME`) as
`.env.<name>` files:

```sh
# desc: Claude on Bedrock (dev)
export CLAUDE_CODE_USE_BEDROCK=1
export AWS_REGION=us-east-1
export AWS_PROFILE=dev
```

- An optional `# desc:` header shows up in `eg list`.
- `$VAR` / `${VAR}` are expanded against earlier lines (and the environment).
- A value containing a 1Password reference (`op://Vault/Item/field`) is treated as
  a secret and resolved on demand — see below.

Create them by hand, or scaffold from a built-in recipe:

```sh
eg recipes
eg add claude-code bedrock --name bedrock-dev --param aws_profile=dev
```

## Secrets (1Password)

Store the secret in 1Password and reference it in the profile:

```sh
export ANTHROPIC_AUTH_TOKEN=op://Private/myproxy/credential
```

- `eg use` / `eg exec` resolve refs **in memory** via the `op` CLI — never to disk.
- Profiles with no refs never touch `op` at all.
- If `op` is missing or signed out, ref profiles fail cleanly (the shim never
  applies a half-set environment).

## Commands

| Command | What it does |
|---|---|
| `eg use <profile>` | Load a profile into the current shell (replaces the previous one). |
| `eg off` | Clear the active profile from the current shell. |
| `eg exec <profile> -- <cmd>` | Run a command with the profile's env injected (subprocess only). |
| `eg render <profile> [-o .env]` | Write a project env file. Default: `op://` refs. `--resolve`: real values (must be git-ignored). `--force-resolve`: real values, no check. |
| `eg add <tool> <provider>` | Create a profile (and tool config) from a recipe. |
| `eg recipes` | List built-in recipes. |
| `eg doctor [profile]` | Send a tiny real request to verify a provider works. |
| `eg list` / `eg status` | List profiles / show the active profile in this shell. |
| `eg edit <profile>` | Open a profile in `$EDITOR`. |
| `eg init <zsh\|bash>` | Print the shell integration snippet. |

## Per-project providers (optional)

`eg` switches per shell, on command. If you'd rather bind a provider to a
*project directory*, compose with [direnv](https://direnv.net): put
`eg use vertex-acme` (or `source`) logic in the repo's `.envrc`.

## File-configured tools (Codex)

Some tools read a config file instead of env vars. `eg add` handles those too —
e.g. Codex needs a provider block in `~/.codex/config.toml` plus its `OPENAI_API_KEY`:

```sh
eg add codex openai-compat \
  --param provider_id=myproxy \
  --param base_url=https://proxy.example.com/v1 \
  --param model=claude-sonnet-4-6 \
  --param key_ref=op://Private/myproxy/credential
# then:  codex --profile myproxy
```

`eg` writes the `[model_providers.myproxy]` block and a `myproxy.config.toml`
overlay atomically, keeping a timestamped backup and preserving the rest of your
config (including comments and other providers).

## Use from an AI agent

`skills/eg/SKILL.md` ships a skill so an agent can run commands against a provider
(`eg exec <profile> -- <cmd>`) or smoke-test it (`eg doctor`) without touching your
shell or seeing secrets. Copy it to `~/.claude/skills/eg/`.

## Security model

- Emitted shell code single-quote-escapes every value; variable names are
  validated — crafted values/keys can't inject shell commands.
- `use`/`off` write shell code to stdout only; all messages go to stderr, so what
  the shim `eval`s is always clean.
- `render` defaults to writing `op://` references, not secrets; writing real
  values is gated on the target being git-ignored unless you opt out explicitly.

## License

MIT.
