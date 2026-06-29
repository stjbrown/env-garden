package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/config"
	"github.com/stjbrown/env-garden/internal/shell"
)

const (
	setupBegin = "# >>> env-garden >>>"
	setupEnd   = "# <<< env-garden <<<"
)

func newSetupCmd() *cobra.Command {
	var shellName string
	c := &cobra.Command{
		Use:   "setup [zsh|bash]",
		Short: "Add the env-garden integration line to your shell rc file",
		Long: "Writes `eval \"$(eg init <shell>)\"` into your rc file (idempotently, with a\n" +
			"backup). The written line re-runs `eg init` on every shell start, so future\n" +
			"upgrades and `eg default` changes are picked up without re-running setup.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := shellName
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				name = detectShell()
			}
			kind, err := shell.ParseKind(name)
			if err != nil {
				return err
			}
			rc, err := rcFile(kind)
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			// Absolute path → immune to PATH ordering at rc time; still re-runs
			// `eg init` each startup so upgrades/default changes are picked up.
			line := fmt.Sprintf(`eval "$(%s init %s)"`, bin, kind)
			block := setupBegin + "\n" + line + "\n" + setupEnd

			existing, err := os.ReadFile(rc)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			text := string(existing)

			if strings.Contains(text, setupBegin) {
				updated, changed := replaceBlock(text, block)
				if !changed {
					fmt.Fprintf(os.Stderr, "eg: already configured in %s\n", rcShort(rc))
					return nil
				}
				if err := backupRC(rc, existing); err != nil {
					return err
				}
				if err := os.WriteFile(rc, []byte(updated), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "eg: updated env-garden block in %s\n", rcShort(rc))
			} else {
				if len(existing) > 0 {
					if err := backupRC(rc, existing); err != nil {
						return err
					}
				}
				if len(text) > 0 && !strings.HasSuffix(text, "\n") {
					text += "\n"
				}
				text += "\n" + block + "\n"
				if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(rc, []byte(text), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "eg: added env-garden to %s\n", rcShort(rc))
			}

			fmt.Fprintf(os.Stderr, "    restart your shell:  exec %s\n", kind)
			if config.ReadDefault() == "" {
				fmt.Fprintln(os.Stderr, "    set a default provider:  eg default <profile>")
			}
			return nil
		},
	}
	c.Flags().StringVar(&shellName, "shell", "", "shell to configure (zsh|bash); default: $SHELL")
	return c
}

// detectShell guesses the shell from $SHELL, defaulting to zsh.
func detectShell() string {
	base := filepath.Base(os.Getenv("SHELL"))
	switch base {
	case "bash":
		return "bash"
	default:
		return "zsh"
	}
}

// rcFile returns the rc file path for a shell (honoring $ZDOTDIR for zsh).
func rcFile(k shell.Kind) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch k {
	case shell.Bash:
		return filepath.Join(home, ".bashrc"), nil
	default:
		if z := os.Getenv("ZDOTDIR"); z != "" {
			return filepath.Join(z, ".zshrc"), nil
		}
		return filepath.Join(home, ".zshrc"), nil
	}
}

// replaceBlock swaps the existing env-garden block for newBlock, reporting
// whether anything changed.
func replaceBlock(text, newBlock string) (string, bool) {
	start := strings.Index(text, setupBegin)
	end := strings.Index(text, setupEnd)
	if start < 0 || end < 0 || end < start {
		return text, false
	}
	end += len(setupEnd)
	old := text[start:end]
	if old == newBlock {
		return text, false
	}
	return text[:start] + newBlock + text[end:], true
}

func backupRC(rc string, data []byte) error {
	bak := fmt.Sprintf("%s.eg-bak-%s", rc, time.Now().Format("20060102-150405"))
	if err := os.WriteFile(bak, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "eg: backed up %s -> %s\n", rcShort(rc), filepath.Base(bak))
	return nil
}

func rcShort(rc string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(rc, home) {
		return "~" + rc[len(home):]
	}
	return rc
}
