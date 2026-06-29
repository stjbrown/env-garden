# env-garden (`eg`)

**A per-shell environment profile switcher with batteries.** Keep named sets of
environment variables (with secrets in 1Password), and switch them
**independently in each terminal** — Claude Code on Vertex in one window, Bedrock
in another, a customer endpoint in a third.

```sh
eg use bedrock                  # switch THIS shell to the bedrock profile
eg exec vertex -- python app.py # run one command with the vertex profile's env
eg render staging -o .env       # write a project .env file from a profile
eg doctor bedrock               # send a tiny real request to check it works
```

A *profile* is just a named set of env vars. You can project it onto your
**current shell** (`use`), a **subprocess** (`exec`), or a **project file**
(`render`).

> **Supported:** macOS and Linux (Intel & Apple Silicon / arm64), shells **zsh**
> and **bash**. Windows isn't supported (works under WSL). See
> [Requirements](#requirements) for optional dependencies.

---

## Quickstart

```sh
# 1. install
brew install stjbrown/tap/eg

# 2. wire it into your shell (writes one line to ~/.zshrc, with a backup)
eg setup

# 3. create a profile from a recipe (see `eg recipes` for the list)
eg add claude-code vertex --param project=my-gcp-project

# 4. make it the default for new shells, then reload
eg default vertex
exec zsh
```

That's it. Every new terminal now starts on `vertex`. Switch any single window
with `eg use bedrock`; check the current one with `eg status`.

> **Why the `eg setup` step?** Changing your *current* shell's environment is only
> possible from a function sourced into that shell — a plain binary runs in a
> child process and can't reach back. `eg setup` adds `eval "$(eg init zsh)"` to
> your rc file, which installs that function. (Same mechanism as `direnv`,
> `nvm`, `zoxide`.)

---

## Requirements

| | |
|---|---|
| **OS** | macOS or Linux, `amd64` or `arm64`. Windows via WSL. |
| **Shell** | `zsh` or `bash` (interactive). fish isn't supported yet. |
| **Secrets** (optional) | [1Password CLI](https://developer.1password.com/docs/cli/) (`op`) — only needed for profiles that use `op://` references. |
| **`eg doctor`** (optional) | `aws` CLI for Bedrock, `gcloud` for Vertex — only for live-testing those providers. |

`eg` itself has no runtime dependencies; the optional tools above are only used by
the features that need them.

---

## The three ways to use a profile

| You want to… | Use | Effect |
|---|---|---|
| Switch the shell you're typing in | `eg use <profile>` | Sets the vars in **this shell** (and child processes you launch from it). |
| Run one command on a profile | `eg exec <profile> -- <cmd>` | Injects the vars into **just that subprocess**. Your shell is untouched; secrets never hit disk. |
| Give an app a `.env` file | `eg render <profile>` | Writes a **file** the app reads (e.g. `docker compose`, a framework). |

`use` is for interactive work, `exec` is great for scripts/agents/CI, `render`
is for tools that only read files.

---

## Profiles

Profiles live in `~/.config/env-garden/` as `.env.<name>` files (honors
`$XDG_CONFIG_HOME`):

```sh
# desc: Claude on Bedrock (dev)        ← shown in `eg list`
export CLAUDE_CODE_USE_BEDROCK=1
export AWS_REGION=us-east-1
export AWS_PROFILE=dev
```

- `$VAR` / `${VAR}` are expanded against earlier lines and the environment.
- A value containing `op://Vault/Item/field` is a **1Password reference** —
  resolved on demand, never stored in plaintext (see [Secrets](#secrets-1password)).

Make them with `eg add` (recommended) or by hand:

```sh
eg recipes                                   # list built-in (tool, provider) recipes
eg add claude-code bedrock --param aws_profile=dev   # scaffold ~/.config/env-garden/.env.bedrock
eg edit bedrock                              # open it in $EDITOR
eg list                                      # see all profiles
```

Built-in recipes include `claude-code/{bedrock,vertex,anthropic}`,
`codex/openai-compat`, and `cursor/cli`. (`anthropic` covers any
OpenAI/Anthropic-compatible proxy or gateway.)

### Secrets (1Password)

Put the secret in 1Password and reference it from the profile:

```sh
export ANTHROPIC_AUTH_TOKEN=op://Private/my-proxy/credential
```

- `use` / `exec` resolve refs **in memory** via the `op` CLI — never to disk.
- Profiles with no `op://` refs never invoke `op` at all.
- If `op` isn't installed or signed in, a profile that needs it **fails cleanly**
  rather than applying a half-set environment.

Requires the [1Password CLI](https://developer.1password.com/docs/cli/) (`brew
install --cask 1password-cli`), ideally with the desktop-app integration enabled
(Settings → Developer → *Integrate with 1Password CLI*).

### Combining profiles

Keep each concern in its own small profile and **stack them** at use time, rather
than maintaining one big copy per combination. `use`, `exec`, `export`, and
`render` all accept several profiles and merge them left-to-right:

```sh
# one .env with your dev Vertex + a customer's Zscaler proxy + Slack creds
eg render dev-vertex zscaler slack -o .env --resolve

eg use dev-vertex zscaler slack            # load the merged set into this shell
eg exec dev-vertex zscaler slack -- python agent.py
eval "$(eg export dev-vertex zscaler slack)"
```

- **Order matters:** when two profiles set the same key, the **later** one wins.
  Each override is reported to stderr (e.g. `eg: HTTPS_PROXY from "zscaler"
  overrides "dev-vertex"`) so a clobbered value is never silent.
- `EG_ACTIVE` becomes the joined name (`dev-vertex+zscaler+slack`), and `eg
  status` shows the full merged variable set.
- With `eg exec`, the `--` separator is **required** when passing more than one
  profile, so profile names and the command stay unambiguous.

---

## Switching

### Per window (default)

Each terminal is independent — set a different profile in each:

```sh
# terminal 1
eg use vertex   && claude        # this window talks to Vertex

# terminal 2
eg use bedrock  && claude        # this one talks to Bedrock, at the same time
```

New shells start on whatever `eg default` is set to; `eg use` overrides for the
current shell only; `eg off` clears it.

### Per project (optional)

Bind a provider to a directory with [direnv](https://direnv.net) — put this in
the repo's `.envrc`:

```sh
eval "$(eg export vertex)"   # (or: eg use vertex)
```

---

## Command reference

| Command | What it does |
|---|---|
| `eg setup [zsh\|bash]` | Add the integration line to your rc file (idempotent, with backup). |
| `eg default [profile]` | Get/set the provider new shells start on (auto-applied by the shim). |
| `eg use <profile>…` | Load one or more profiles into the current shell (merged left-to-right). |
| `eg off` | Clear the active profile from the current shell. |
| `eg exec <profile>… -- <cmd>` | Run a command with the profile(s)' env injected (subprocess only). `--` required for 2+ profiles. |
| `eg render <profile>… [-o .env]` | Write a project env file (merging multiple profiles). Default: `op://` refs. `--resolve`: real values (output must be git-ignored). `--force-resolve`: skip the check. |
| `eg export <profile>…` | Print the profile(s) as `export` statements (for scripts / `eval "$(eg export x)"` in a direnv `.envrc`). |
| `eg add <tool> <provider>` | Create a profile (and any tool config) from a recipe. |
| `eg recipes` | List built-in recipes. |
| `eg doctor [profile]` | Send a tiny real request to verify a provider works. |
| `eg list` / `eg status` | List profiles / show the active profile in this shell. |
| `eg edit <profile>` | Open a profile in `$EDITOR`. |
| `eg init <zsh\|bash>` | Print the shell-integration snippet (what `setup` writes). |

---

## Troubleshooting

**`eg use` prints `export …` lines instead of switching.**
The shell function isn't loaded — `eg` is resolving to the bare binary. Check:

```sh
whence -w eg     # zsh   (bash: type eg)
```

If it doesn't say `eg: function`, run `eg setup`, then `exec zsh`. (A binary can
only *print* the exports; the function is what `eval`s them into your shell.)

**A new machine's shells/agents have no provider vars.**
Same cause — `~/.zshrc` on that machine is missing the integration line. Run
`eg setup` there, set `eg default <profile>`, and copy/recreate your profiles
(`~/.config/env-garden/` — they hold only `op://` refs, no secrets, so they're
safe to sync).

**A profile with secrets fails with an `op` message.**
Install the 1Password CLI and enable the app integration (or `op signin`).

**A GUI tool (IDE, cmux, etc.) launches an agent without the env.**
GUI-launched agents inherit the environment of the **shell that started them**.
If that's an interactive shell sourcing `~/.zshrc`, `eg default` is enough; if the
tool execs a binary directly, point its launcher at `eg exec <profile> -- <cmd>`.

---

## Using providers that need a config file (Codex)

Some tools read a config file instead of env vars. `eg add` handles those too —
Codex needs a provider block in `~/.codex/config.toml`:

```sh
eg add codex openai-compat \
  --param provider_id=my-proxy \
  --param base_url=https://proxy.example.com/v1 \
  --param model=claude-sonnet-4-6 \
  --param key_ref=op://Private/my-proxy/credential
# then run Codex with:  codex --profile my-proxy
```

`eg` writes the `[model_providers.my-proxy]` block and a `my-proxy.config.toml`
overlay atomically — with a timestamped backup, preserving the rest of your
config (comments and other providers included).

## Using `eg` from an AI agent

`skills/eg/SKILL.md` is an [agent skill](https://github.com/vercel-labs/skills)
so an agent can run commands against a provider (`eg exec <profile> -- <cmd>`) or
smoke-test one (`eg doctor`) without touching your shell or seeing secrets.

Install it with the [`skills`](https://github.com/vercel-labs/skills) CLI:

```sh
# install just the eg skill, globally, for Claude Code
npx skills add stjbrown/env-garden --skill eg -g -a claude-code

# or interactively pick from the repo
npx skills add stjbrown/env-garden
```

It also works with any other agent the CLI supports (`-a opencode`, etc.). To
install by hand instead, copy `skills/eg/SKILL.md` to `~/.claude/skills/eg/`.

The skill is self-bootstrapping: if it's installed before the `eg` binary, it
tells the agent how to install `eg` first (`brew install stjbrown/tap/eg`).

---

## How it works

`eg` is a small Go binary plus a shell shim. The binary prints shell code; the
shim (`eval "$(eg init zsh)"`, installed by `eg setup`) defines an `eg` function
that runs the binary and `eval`s its output into your shell. That's the only way
a tool can change your live shell — and it's why switching is naturally
per-shell. `eg default` is applied automatically by the shim on each new shell.

**Security:** emitted shell code single-quote-escapes every value and validates
variable names, so crafted values/keys can't inject commands; `use`/`off` write
only shell code to stdout (messages go to stderr) so what's `eval`'d is always
clean; `render` defaults to `op://` references and won't write real secrets to a
non-git-ignored file unless you pass `--force-resolve`.

## Install (other methods)

```sh
# Homebrew (recommended)
brew install stjbrown/tap/eg

# install script (no Homebrew)
curl -fsSL https://raw.githubusercontent.com/stjbrown/env-garden/main/install.sh | sh

# from source
go install github.com/stjbrown/env-garden/cmd/eg@latest
```

## License

MIT
