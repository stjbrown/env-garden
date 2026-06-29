package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available profiles",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summaries, err := profile.List()
			if err != nil {
				return err
			}
			if len(summaries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No profiles yet. Create one with: eg add <tool> <provider>")
				return nil
			}
			active := os.Getenv("EG_ACTIVE")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, s := range summaries {
				marker := "  "
				if s.Name == active {
					marker = "* "
				}
				fmt.Fprintf(w, "%s%s\t%s\n", marker, s.Name, s.Desc)
			}
			return w.Flush()
		},
	}
}
