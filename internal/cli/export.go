package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <profile>",
		Short: "Print a profile's environment as shell export statements",
		Long: "Emit `export KEY=value` lines for a profile (op:// refs resolved). Useful in\n" +
			"scripts or a direnv .envrc:\n\n  eval \"$(eg export vertex)\"",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeProfiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := profile.Load(args[0])
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no profile %q (try: eg list)", args[0])
				}
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitUse("", args[0], pairs))
			return nil
		},
	}
}
