package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the active profile in this shell",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			active := os.Getenv(shell.ActiveVar)
			if active == "" {
				fmt.Fprintln(out, "No active profile in this shell. Load one with: eg use <profile>")
				return nil
			}
			fmt.Fprintf(out, "Active profile: %s\n", active)

			managed := strings.Fields(os.Getenv(shell.ManagedVar))
			if len(managed) == 0 {
				return nil
			}
			sort.Strings(managed)
			fmt.Fprintln(out, "Managed variables:")
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			for _, k := range managed {
				fmt.Fprintf(w, "  %s\t%s\n", k, mask(k, os.Getenv(k)))
			}
			return w.Flush()
		},
	}
}

// mask hides the value of variables that look like secrets.
func mask(key, val string) string {
	if val == "" {
		return "(empty)"
	}
	k := strings.ToUpper(key)
	for _, hint := range []string{"KEY", "TOKEN", "SECRET", "PASSWORD"} {
		if strings.Contains(k, hint) {
			if len(val) <= 4 {
				return "****"
			}
			return val[:4] + "…"
		}
	}
	return val
}
