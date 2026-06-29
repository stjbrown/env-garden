package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile> [profile...]",
		Short: "Load one or more profiles into the current shell",
		Long: "Emit shell code (evaluated by the eg shim) that switches the current\n" +
			"shell to the named profile(s), replacing any previously-loaded profile.\n\n" +
			"Multiple profiles are merged in order (later values win on conflicts):\n\n" +
			"  eg use dev-vertex zscaler slack",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeProfiles,
		// stdout is the code the shim evals — keep it clean; everything else
		// goes to stderr.
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := loadMerged(args)
			if err != nil {
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}
			prev := os.Getenv(shell.ManagedVar)
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitUse(prev, p.Name, pairs))
			fmt.Fprintf(os.Stderr, "eg: switched to %s\n", p.Name)
			return nil
		},
	}
}
