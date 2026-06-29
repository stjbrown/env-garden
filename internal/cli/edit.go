package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/config"
	"github.com/stjbrown/env-garden/internal/profile"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "edit <profile>",
		Short:             "Open a profile in $EDITOR",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeProfile,
		RunE: func(cmd *cobra.Command, args []string) error {
			exists, err := profile.Exists(args[0])
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("no profile %q (create one with: eg add <tool> <provider>)", args[0])
			}
			path, err := config.ProfilePath(args[0])
			if err != nil {
				return err
			}
			editor := firstEnv("EG_EDITOR", "VISUAL", "EDITOR")
			if editor == "" {
				editor = "vi"
			}
			ed := exec.Command("sh", "-c", editor+` "$1"`, "sh", path)
			ed.Stdin, ed.Stdout, ed.Stderr = os.Stdin, os.Stdout, os.Stderr
			return ed.Run()
		},
	}
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
