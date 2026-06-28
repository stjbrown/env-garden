package cli

import (
	"fmt"
	"os"

	"github.com/stjbrown/env-garden/internal/shell"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <zsh|bash>",
		Short: "Print the shell integration snippet",
		Long: "Print the shell function to evaluate in your rc file:\n\n" +
			"  eval \"$(eg init zsh)\"   # ~/.zshrc\n" +
			"  eval \"$(eg init bash)\"  # ~/.bashrc",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := shell.ParseKind(args[0])
			if err != nil {
				return err
			}
			bin, err := os.Executable()
			if err != nil {
				return fmt.Errorf("locating eg binary: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), shell.Init(kind, bin))
			return nil
		},
	}
}
