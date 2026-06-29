package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/config"
	"github.com/stjbrown/env-garden/internal/profile"
)

func newDefaultCmd() *cobra.Command {
	var clear bool
	c := &cobra.Command{
		Use:   "default [profile]",
		Short: "Get or set the default profile applied to new shells",
		Long: "With no argument, prints the current default. With a profile, sets it.\n" +
			"The default is applied automatically by the shell integration when a new\n" +
			"shell starts (and no profile is already active). Stored as data in the\n" +
			"config dir, so it survives upgrades and can be synced across machines.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeProfile,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clear {
				if err := config.ClearDefault(); err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "eg: cleared default profile")
				return nil
			}
			if len(args) == 0 {
				if d := config.ReadDefault(); d != "" {
					fmt.Fprintln(cmd.OutOrStdout(), d)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "(no default set)")
				}
				return nil
			}
			name := args[0]
			exists, err := profile.Exists(name)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("no profile %q (try: eg list)", name)
			}
			if err := config.WriteDefault(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "eg: default profile set to %s (applies to new shells)\n", name)
			return nil
		},
	}
	c.Flags().BoolVar(&clear, "clear", false, "remove the default profile")
	return c
}
