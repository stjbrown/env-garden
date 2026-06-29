package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/config"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/shell"
)

// newBootCmd is the hidden command the shell integration calls on startup to
// apply the default profile. It must never break shell startup: any problem is
// reported to stderr and it emits nothing (the shim then evals nothing).
func newBootCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "boot",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := config.ReadDefault()
			if name == "" {
				return nil // no default → nothing to apply
			}
			p, err := profile.Load(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "eg: default profile %q not found (eg default <profile>)\n", name)
				return nil
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "eg: default %q: %v\n", name, err)
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitUse("", name, pairs))
			return nil
		},
	}
}
