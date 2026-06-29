package shell

import "fmt"

// Kind identifies a supported shell.
type Kind string

const (
	Zsh  Kind = "zsh"
	Bash Kind = "bash"
)

// ParseKind validates a shell name.
func ParseKind(s string) (Kind, error) {
	switch Kind(s) {
	case Zsh:
		return Zsh, nil
	case Bash:
		return Bash, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (use zsh or bash)", s)
	}
}

// Init returns the shell integration snippet for `eval "$(eg init <shell>)"`.
// binPath is the absolute path to the eg binary, baked in so the shim never
// re-invokes the function that shadows it.
//
// The shim captures stdout (the code to eval) only for the env-mutating verbs
// `use` and `off`, and evals it only on success — so a failed resolve never
// partially mutates the shell. Every other verb execs the binary directly so
// its stdout/stderr and exit code pass through untouched. All human-facing
// messages are written to stderr by the binary, keeping stdout eval-clean.
//
// After defining the function it applies the default profile (if one is set and
// the shell isn't already in a profile), so a single `eval "$(eg init zsh)"`
// line both installs the shim and lands every new shell on the default. Because
// the line re-runs `eg init` each startup, upgrades and default changes are
// picked up with no need to re-edit the rc file.
//
// The body is pure POSIX sh (no arrays, no [[ ]]), valid in zsh and bash 3.2.
func Init(k Kind, binPath string) string {
	q := SingleQuote(binPath)
	return fmt.Sprintf(`# env-garden shell integration (%s)
__EG_BIN=%s
eg() {
  case "$1" in
    use|off)
      __eg_out="$("$__EG_BIN" "$@")"; __eg_status=$?
      [ $__eg_status -eq 0 ] && eval "$__eg_out"
      unset __eg_out
      return $__eg_status ;;
    *)
      "$__EG_BIN" "$@" ;;
  esac
}
if [ -z "${EG_ACTIVE:-}" ]; then
  __eg_boot="$("$__EG_BIN" boot 2>/dev/null)"
  [ -n "$__eg_boot" ] && eval "$__eg_boot"
  unset __eg_boot
fi
`, k, q)
}
