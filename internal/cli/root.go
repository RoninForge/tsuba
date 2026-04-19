// Package cli builds the cobra command tree. Keeping the tree in a
// dedicated package (rather than inside cmd/tsuba) makes it testable
// without spawning a binary.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/RoninForge/tsuba/internal/version"
	"github.com/spf13/cobra"
)

// ErrFailed is returned by subcommands when the work they were asked to
// do did not produce the expected result. Execute maps this sentinel to
// exit code 1. Usage errors (bad flag, missing arg) come back via cobra
// and map to exit code 2.
var ErrFailed = errors.New("tsuba failed")

// Execute is the single entry point that cmd/tsuba/main.go calls.
// Returns a process exit code. Testable from unit tests.
//
// Exit codes:
//
//	0 - success
//	1 - command ran but the work failed (validation error, scaffold refused)
//	2 - invocation error (bad flag, missing file)
func Execute(stdout, stderr io.Writer, args []string) int {
	root := newRoot(stdout, stderr)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if errors.Is(err, ErrFailed) {
			return 1
		}
		fmt.Fprintln(stderr, "Error:", err)
		return 2
	}
	return 0
}

func newRoot(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tsuba",
		Short: "Scaffold marketplace-ready Claude Code skills and plugins",
		Long: `tsuba generates valid Claude Code plugin directories and skills in
seconds. Every scaffolded plugin ships with a correct .claude-plugin/plugin.json,
a sample skill, LICENSE, and README, and passes hanko validation on the
first run.

The name means "sword guard" (鍔) in Japanese - the disc between blade
and handle that lets a swordsman hand the weapon off safely. tsuba does
the same thing for your plugin: makes it safe to hand off to the
marketplace.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       formatVersion(),
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.AddCommand(newNewCmd(stdout, stderr))
	cmd.AddCommand(newValidateCmd(stdout, stderr))
	cmd.AddCommand(newDoctorCmd(stdout, stderr))
	cmd.AddCommand(newListCmd(stdout))
	cmd.AddCommand(newVersionCmd(stdout))

	return cmd
}

func formatVersion() string {
	v := version.Get()
	return fmt.Sprintf("%s (commit %s, built %s)", v.Version, v.Commit, v.BuildDate)
}
