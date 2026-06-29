package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Load a profile into the current shell",
		Long: "Emit shell code (evaluated by the eg shim) that switches the current\n" +
			"shell to the named profile, replacing any previously-loaded profile.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeProfiles,
		// stdout is the code the shim evals — keep it clean; everything else
		// goes to stderr.
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			p, err := profile.Load(name)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no profile %q (try: eg list)", name)
				}
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}
			prev := os.Getenv(shell.ManagedVar)
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitUse(prev, name, pairs))
			fmt.Fprintf(os.Stderr, "eg: switched to %s\n", name)
			return nil
		},
	}
}
