package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/profile"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newExecCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "exec <profile> [--] <command> [args...]",
		Short: "Run a command with a profile's environment injected",
		Long: "Run a command in a subprocess with the profile's variables set,\n" +
			"resolving op:// references in-memory. The current shell is untouched.\n\n" +
			"  eg exec myproxy -- python agent.py",
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     completeProfiles,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, rest, err := splitExecArgs(args)
			if err != nil {
				return err
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

			child := exec.Command(rest[0], rest[1:]...)
			child.Env = mergeEnv(os.Environ(), pairs, name)
			child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := child.Run(); err != nil {
				var ee *exec.ExitError
				if errors.As(err, &ee) {
					os.Exit(ee.ExitCode())
				}
				return fmt.Errorf("running %s: %w", rest[0], err)
			}
			return nil
		},
	}
	// Keep everything after the profile name verbatim for the child command.
	c.Flags().SetInterspersed(false)
	return c
}

// splitExecArgs separates the profile name from the command to run. With
// SetInterspersed(false), pflag stops flag parsing after the profile name and
// passes a literal "--" through, so we strip an optional leading "--" here.
func splitExecArgs(args []string) (name string, rest []string, err error) {
	name, rest = args[0], args[1:]
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return "", nil, errors.New("no command given (usage: eg exec <profile> -- <command>)")
	}
	return name, rest, nil
}

// mergeEnv overlays the resolved pairs (and EG_ACTIVE) onto base, with later
// entries winning.
func mergeEnv(base []string, pairs []shell.KV, active string) []string {
	idx := make(map[string]int, len(base))
	out := make([]string, len(base))
	copy(out, base)
	for i, kv := range out {
		if eq := strings.IndexByte(kv, '='); eq > 0 {
			idx[kv[:eq]] = i
		}
	}
	set := func(k, v string) {
		entry := k + "=" + v
		if i, ok := idx[k]; ok {
			out[i] = entry
		} else {
			idx[k] = len(out)
			out = append(out, entry)
		}
	}
	for _, p := range pairs {
		set(p.Key, p.Value)
	}
	set(shell.ActiveVar, active)
	return out
}
