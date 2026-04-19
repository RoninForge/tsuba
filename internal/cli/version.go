package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tsuba version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(stdout, formatVersion())
			return nil
		},
	}
}
