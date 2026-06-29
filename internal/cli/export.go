package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <profile> [profile...]",
		Short: "Print profiles' environment as shell export statements",
		Long: "Emit `export KEY=value` lines for one or more profiles (op:// refs\n" +
			"resolved). Multiple profiles are merged in order (later values win).\n" +
			"Useful in scripts or a direnv .envrc:\n\n" +
			"  eval \"$(eg export dev-vertex zscaler slack)\"",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeProfiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := loadMerged(args)
			if err != nil {
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitUse("", p.Name, pairs))
			return nil
		},
	}
}
