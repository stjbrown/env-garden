package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/stjbrown/env-garden/internal/doctor"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/shell"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var (
		model    string
		insecure bool
	)
	c := &cobra.Command{
		Use:   "doctor [profile]",
		Short: "Smoke-test a profile against its provider",
		Long: "Send a tiny real request to verify a profile actually works.\n" +
			"With no argument, tests the profile active in this shell.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeProfiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := os.Getenv(shell.ActiveVar)
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				return errors.New("no profile given and none active (usage: eg doctor <profile>)")
			}
			p, err := profile.Load(name)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("no profile %q (try: eg list)", name)
				}
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}
			env := make(map[string]string, len(pairs))
			for _, kv := range pairs {
				env[kv.Key] = kv.Value
			}

			fmt.Fprintf(os.Stderr, "eg: testing %s…\n", name)
			res := doctor.Run(env, model, insecure)
			out := cmd.OutOrStdout()
			if res.OK {
				fmt.Fprintf(out, "✅ %s (%s): %s\n", name, res.Provider, res.Reply)
				return nil
			}
			fmt.Fprintf(out, "❌ %s (%s): %s\n", name, res.Provider, res.Detail)
			return errSilent
		},
	}
	c.Flags().StringVar(&model, "model", "", "model to test (default: provider's default / discovered)")
	c.Flags().BoolVar(&insecure, "insecure", false, "skip TLS verification (for self-signed corporate proxies)")
	return c
}
