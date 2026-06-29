package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/recipe"
)

func newRecipesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recipes",
		Short: "List built-in recipes for `eg add`",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := recipe.Catalog()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TOOL\tPROVIDER\tDESCRIPTION")
			for _, r := range all {
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Tool, r.Provider, r.Description)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nCreate one with: eg add <tool> <provider>")
			return nil
		},
	}
}
