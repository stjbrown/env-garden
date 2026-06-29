package cli

import (
	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
)

// completeProfiles offers profile names for shell completion.
func completeProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	summaries, err := profile.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, 0, len(summaries))
	for _, s := range summaries {
		names = append(names, s.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
