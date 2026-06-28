// Package cli wires up the eg command tree.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set via -ldflags at release time.
var version = "dev"

// errSilent signals a nonzero exit without an extra "eg: ..." line (the command
// has already printed its own user-facing failure).
var errSilent = errors.New("")

// SetVersion lets main inject the build version.
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

// NewRoot builds the root command with all subcommands attached.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "eg",
		Short: "env-garden — a per-shell environment profile switcher",
		Long: "eg switches the current shell, a subprocess, or a project file between\n" +
			"named environment profiles (.env.<name> in ~/.config/env-garden).",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.HiddenDefaultCmd = true

	root.AddCommand(
		newInitCmd(),
		newUseCmd(),
		newOffCmd(),
		newExecCmd(),
		newRenderCmd(),
		newAddCmd(),
		newRecipesCmd(),
		newDoctorCmd(),
		newListCmd(),
		newStatusCmd(),
		newEditCmd(),
	)
	return root
}

// Execute runs the command tree, printing errors to stderr.
func Execute() int {
	if err := NewRoot().Execute(); err != nil {
		if !errors.Is(err, errSilent) {
			fmt.Fprintln(os.Stderr, "eg: "+err.Error())
		}
		return 1
	}
	return 0
}
