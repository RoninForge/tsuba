package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newListCmd(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List supported marketplaces (v0.1) or planned surface",
	}
	cmd.AddCommand(newListMarketplacesCmd(stdout))
	return cmd
}

// marketplaceEntry captures the human-readable summary of a marketplace
// target for the `list marketplaces` command. The full submission
// logic will live behind `tsuba publish` in v0.2.
type marketplaceEntry struct {
	Slug  string
	Name  string
	Notes string
	URL   string
}

// Marketplaces reflects what we researched in phase-1-spec.md section
// 5. Kept here (not in a separate package) because the data is small
// and editorial - every row was written by hand. When v0.2 adds
// `tsuba publish`, we'll promote this to its own package.
var marketplaces = []marketplaceEntry{
	{
		Slug:  "anthropic",
		Name:  "Anthropic official plugin marketplace",
		Notes: "Submit via clau.de/plugin-directory-submission. Strict reserved-name list, author object strongly required.",
		URL:   "https://github.com/anthropics/claude-plugins-official",
	},
	{
		Slug:  "buildwithclaude",
		Name:  "buildwithclaude (davepoon)",
		Notes: "PR against the repo. Directory naming convention: plugins/<type>-<category>/.claude-plugin/plugin.json.",
		URL:   "https://github.com/davepoon/buildwithclaude",
	},
	{
		Slug:  "cc-marketplace",
		Name:  "cc-marketplace (ananddtyagi)",
		Notes: "PR against the repo. Requires name, version, and description as non-optional fields.",
		URL:   "https://github.com/ananddtyagi/cc-marketplace",
	},
	{
		Slug:  "claudemarketplaces",
		Name:  "claudemarketplaces.com",
		Notes: "Auto-discovery. Host marketplace.json at your own repo; the site crawls GitHub daily.",
		URL:   "https://claudemarketplaces.com/",
	},
}

func newListMarketplacesCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "marketplaces",
		Short: "List known Claude Code plugin marketplaces and submission notes",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(stdout, "Supported marketplaces for `tsuba publish` (v0.2):")
			fmt.Fprintln(stdout)
			for _, m := range marketplaces {
				fmt.Fprintf(stdout, "  %-20s  %s\n", m.Slug, m.Name)
				fmt.Fprintf(stdout, "  %-20s  %s\n", "", m.URL)
				fmt.Fprintf(stdout, "  %-20s  %s\n\n", "", m.Notes)
			}
			return nil
		},
	}
}
