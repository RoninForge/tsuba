package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/RoninForge/tsuba/internal/gitctx"
	"github.com/RoninForge/tsuba/internal/hanko"
	"github.com/spf13/cobra"
)

func newDoctorCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Report the status of the local environment tsuba depends on",
		Long: `tsuba doctor reports which optional dependencies are set up. All checks
are advisory - tsuba's scaffolding commands work without hanko or a
configured git identity, but several fields will be placeholders.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runDoctor(stdout, stderr)
			return nil
		},
	}
}

func runDoctor(stdout, stderr io.Writer) {
	fmt.Fprintln(stdout, "tsuba doctor - local environment check")
	fmt.Fprintln(stdout)

	// hanko - needed by `tsuba validate`.
	if hanko.Available() {
		fmt.Fprintf(stdout, "  ok    hanko on PATH: %s\n", hanko.Version())
	} else {
		fmt.Fprintln(stdout, "  warn  hanko not on PATH. `tsuba validate` will fail.")
		fmt.Fprintln(stdout, "        curl -fsSL https://roninforge.org/hanko/install.sh | sh")
	}

	// git - needed to auto-populate author + repository fields.
	if gitctx.Available() {
		name := gitctx.UserName()
		email := gitctx.UserEmail()
		if name != "" && email != "" {
			fmt.Fprintf(stdout, "  ok    git identity: %s <%s>\n", name, email)
		} else {
			fmt.Fprintln(stdout, "  warn  git installed but user.name / user.email not configured.")
			fmt.Fprintln(stdout, "        `tsuba new plugin` author fields will be placeholders.")
			fmt.Fprintln(stdout, "        Fix:")
			fmt.Fprintln(stdout, "          git config --global user.name \"Your Name\"")
			fmt.Fprintln(stdout, "          git config --global user.email \"you@example.com\"")
		}
	} else {
		fmt.Fprintln(stdout, "  warn  git not found. Author fields will need explicit --author and --email flags.")
	}

	// cwd - where the next scaffold will land.
	cwd, err := os.Getwd()
	if err == nil {
		fmt.Fprintf(stdout, "  info  cwd: %s\n", cwd)
	}
}
