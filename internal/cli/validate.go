package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/RoninForge/tsuba/internal/hanko"
	"github.com/spf13/cobra"
)

func newValidateCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate a plugin directory by delegating to hanko",
		Long: `tsuba validate runs "hanko check" against the target directory. Requires
the hanko binary on PATH; install it with:

    curl -fsSL https://roninforge.org/hanko/install.sh | sh

Or see https://github.com/RoninForge/hanko for other install paths.

Pass a directory argument to validate that path; otherwise the current
directory is used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) == 1 {
				target = args[0]
			}
			exitCode, out, err := hanko.Validate(target)
			if err != nil {
				if errors.Is(err, hanko.ErrNotInstalled) {
					fmt.Fprintln(stderr, "hanko is required for validation but was not found on PATH.")
					fmt.Fprintln(stderr, "Install it with:")
					fmt.Fprintln(stderr, "")
					fmt.Fprintln(stderr, "    curl -fsSL https://roninforge.org/hanko/install.sh | sh")
					fmt.Fprintln(stderr, "")
					fmt.Fprintln(stderr, "See https://github.com/RoninForge/hanko for other install paths.")
					return ErrFailed
				}
				return err
			}
			fmt.Fprint(stdout, out)
			if exitCode != 0 {
				return ErrFailed
			}
			return nil
		},
	}
	return cmd
}
