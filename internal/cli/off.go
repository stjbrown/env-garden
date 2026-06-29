package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newOffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "off",
		Short: "Clear the active profile from the current shell",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prev := os.Getenv(shell.ManagedVar)
			fmt.Fprint(cmd.OutOrStdout(), shell.EmitOff(prev))
			fmt.Fprintln(os.Stderr, "eg: cleared active profile")
			return nil
		},
	}
}
