package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stjbrown/env-garden/internal/config"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/recipe"
	"github.com/stjbrown/env-garden/internal/toolconfig"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		name           string
		paramFlags     []string
		nonInteractive bool
		force          bool
	)
	c := &cobra.Command{
		Use:   "add <tool> <provider>",
		Short: "Create a profile from a built-in recipe",
		Long: "Scaffold a profile (and any tool config files) from a recipe.\n" +
			"List recipes with: eg recipes",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := recipe.Lookup(args[0], args[1])
			if err != nil {
				return err
			}
			flagParams, err := parseParams(paramFlags)
			if err != nil {
				return err
			}
			params, err := resolveParams(r, flagParams, nonInteractive)
			if err != nil {
				return err
			}

			profName := name
			if profName == "" {
				profName = r.DefaultName
			}
			exists, err := profile.Exists(profName)
			if err != nil {
				return err
			}
			if exists && !force {
				return fmt.Errorf("profile %q already exists (use --force to overwrite)", profName)
			}

			content, err := recipe.RenderProfile(r, params)
			if err != nil {
				return err
			}
			dir, err := config.EnsureDir()
			if err != nil {
				return err
			}
			path := filepath.Join(dir, config.ProfilePrefix+profName)
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				return err
			}

			if err := applyFiles(r, params); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "eg: created profile %q\n", profName)
			if r.Notes != "" {
				fmt.Fprintf(os.Stderr, "    %s\n", r.Notes)
			}
			fmt.Fprintf(os.Stderr, "    use it:  eg use %s   |   run a command:  eg exec %s -- <cmd>\n", profName, profName)
			return nil
		},
	}
	c.Flags().StringVar(&name, "name", "", "profile name (default: recipe's default)")
	c.Flags().StringArrayVar(&paramFlags, "param", nil, "set a recipe parameter as key=value (repeatable)")
	c.Flags().BoolVar(&nonInteractive, "non-interactive", false, "do not prompt; use --param values and defaults")
	c.Flags().BoolVar(&force, "force", false, "overwrite an existing profile")
	return c
}

// applyFiles materializes a recipe's tool-config file artifacts (e.g. Codex's
// config.toml), dispatching by artifact tool.
func applyFiles(r recipe.Recipe, params map[string]string) error {
	for _, f := range r.Files {
		path, err := config.ExpandTilde(f.Path)
		if err != nil {
			return err
		}
		switch f.Tool {
		case "codex":
			summary, err := toolconfig.ApplyCodex(path, toolconfig.CodexParams{
				ProviderID: params["provider_id"],
				BaseURL:    params["base_url"],
				Model:      params["model"],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "eg: configured codex:\n    %s\n", summary)
		default:
			return fmt.Errorf("recipe %s references unknown tool-config %q", r.Key(), f.Tool)
		}
	}
	return nil
}

// parseParams turns ["k=v", ...] into a map.
func parseParams(flags []string) (map[string]string, error) {
	m := make(map[string]string, len(flags))
	for _, f := range flags {
		eq := strings.IndexByte(f, '=')
		if eq <= 0 {
			return nil, fmt.Errorf("invalid --param %q (want key=value)", f)
		}
		m[f[:eq]] = f[eq+1:]
	}
	return m, nil
}

// resolveParams fills every recipe param from flags, interactive prompts, or
// defaults — in that order.
func resolveParams(r recipe.Recipe, flags map[string]string, nonInteractive bool) (map[string]string, error) {
	interactive := !nonInteractive && isInteractive()
	reader := bufio.NewReader(os.Stdin)
	out := make(map[string]string, len(r.Params))

	for _, p := range r.Params {
		if v, ok := flags[p.Name]; ok {
			out[p.Name] = v
			continue
		}
		if interactive {
			out[p.Name] = prompt(reader, p)
			continue
		}
		if p.Default == "" {
			return nil, fmt.Errorf("missing required parameter %q (pass --param %s=... or run interactively)", p.Name, p.Name)
		}
		out[p.Name] = p.Default
	}
	// Unknown flag params are a likely typo.
	for k := range flags {
		if !hasParam(r, k) {
			return nil, fmt.Errorf("recipe %s has no parameter %q", r.Key(), k)
		}
	}
	return out, nil
}

func prompt(r *bufio.Reader, p recipe.Param) string {
	label := p.Prompt
	if label == "" {
		label = p.Name
	}
	if p.Default != "" {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", label, p.Default)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", label)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return p.Default
	}
	return line
}

func hasParam(r recipe.Recipe, name string) bool {
	for _, p := range r.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

// isInteractive reports whether stdin is a terminal.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
