package cli

import (
	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
)

// completeProfile offers profile names for single-profile commands: it suggests
// nothing once the one positional argument is present.
func completeProfile(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeProfiles(cmd, args, toComplete)
}

// completeProfiles offers profile names for shell completion. Commands that
// accept several profiles call this at every position; already-chosen names are
// filtered out so each profile is suggested only once.
func completeProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	summaries, err := profile.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	chosen := make(map[string]bool, len(args))
	for _, a := range args {
		chosen[a] = true
	}
	names := make([]string, 0, len(summaries))
	for _, s := range summaries {
		if !chosen[s.Name] {
			names = append(names, s.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
