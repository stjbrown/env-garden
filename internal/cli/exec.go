package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stjbrown/env-garden/internal/shell"
)

func newExecCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "exec <profile> [profile...] -- <command> [args...]",
		Short: "Run a command with one or more profiles' environment injected",
		Long: "Run a command in a subprocess with the profile's variables set,\n" +
			"resolving op:// references in-memory. The current shell is untouched.\n\n" +
			"  eg exec myproxy -- python agent.py\n\n" +
			"Multiple profiles are merged in order (later values win); the `--`\n" +
			"separator is required when passing more than one profile:\n\n" +
			"  eg exec dev-vertex zscaler slack -- python agent.py",
		Args:                  cobra.MinimumNArgs(1),
		ValidArgsFunction:     completeProfiles,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			names, rest, err := splitExecArgs(args, profileExists)
			if err != nil {
				return err
			}
			p, err := loadMerged(names)
			if err != nil {
				return err
			}
			pairs, err := resolvePairs(p)
			if err != nil {
				return err
			}

			child := exec.Command(rest[0], rest[1:]...)
			child.Env = mergeEnv(os.Environ(), pairs, p.Name)
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

// splitExecArgs separates the profile name(s) from the command to run. With
// SetInterspersed(false), pflag stops flag parsing at the first positional arg
// and passes a literal "--" through as a regular element, so we split on it
// here:
//
//   - With a "--": everything before it is profile names, everything after is
//     the command. This is the only unambiguous form for multiple profiles.
//   - Without a "--": the first arg is a single profile and the rest is the
//     command (the original single-profile shorthand, kept for compatibility).
//
// isProfile reports whether a name is a known profile; it is used only to give a
// helpful hint when the shorthand is used but the second token also looks like a
// profile (the likely "forgot the --" mistake).
func splitExecArgs(args []string, isProfile func(string) bool) (names []string, rest []string, err error) {
	for i, a := range args {
		if a == "--" {
			names, rest = args[:i], args[i+1:]
			if len(names) == 0 {
				return nil, nil, errors.New("no profile given (usage: eg exec <profile> [profile...] -- <command>)")
			}
			if len(rest) == 0 {
				return nil, nil, errors.New("no command given (usage: eg exec <profile> [profile...] -- <command>)")
			}
			return names, rest, nil
		}
	}
	// No "--": single-profile shorthand.
	names, rest = args[:1], args[1:]
	if len(rest) == 0 {
		return nil, nil, errors.New("no command given (usage: eg exec <profile> [profile...] -- <command>)")
	}
	if isProfile != nil && isProfile(rest[0]) {
		return nil, nil, fmt.Errorf("%q looks like another profile; to combine profiles, separate them from the command with --:\n"+
			"  eg exec %s %s -- <command>", rest[0], names[0], rest[0])
	}
	return names, rest, nil
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
