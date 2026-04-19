package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/RoninForge/tsuba/internal/gitctx"
	"github.com/RoninForge/tsuba/internal/scaffold"
	"github.com/spf13/cobra"
)

// newFlags captures the flags common to every `tsuba new <kind>` variant.
type newFlags struct {
	description   string
	authorName    string
	authorEmail   string
	version       string
	license       string
	noAttribution bool
	force         bool
	into          string // target directory override
}

func (f *newFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.description, "description", "", "short description (defaults to a placeholder)")
	cmd.Flags().StringVar(&f.authorName, "author", "", "author name (defaults to `git config user.name`)")
	cmd.Flags().StringVar(&f.authorEmail, "email", "", "author email (defaults to `git config user.email`)")
	cmd.Flags().StringVar(&f.version, "version", "", "initial version (defaults to 0.1.0)")
	cmd.Flags().StringVar(&f.license, "license", "", "SPDX license id for plugin.json (defaults to MIT)")
	cmd.Flags().BoolVar(&f.noAttribution, "no-attribution", false, "skip the \"scaffolded with tsuba\" README footer")
	cmd.Flags().BoolVar(&f.force, "force", false, "overwrite the target directory if it already exists")
	cmd.Flags().StringVar(&f.into, "into", "", "write the scaffold under this directory instead of the current one")
}

// resolveAuthor picks the final author name + email given the flag
// values and git config fallbacks. Empty strings propagate. When both
// are empty, scaffold omits the author object from plugin.json entirely
// so hanko emits a HANKO003 warning (non-blocking) instead of a
// HANKO-SCHEMA error on minLength. Users can add --author / --email or
// set git config to silence the warning.
func (f *newFlags) resolveAuthor() scaffold.Author {
	name := f.authorName
	if name == "" {
		name = gitctx.UserName()
	}
	email := f.authorEmail
	if email == "" {
		email = gitctx.UserEmail()
	}
	return scaffold.Author{Name: name, Email: email}
}

func newNewCmd(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Scaffold a new skill or plugin",
		Long: `tsuba new <kind> <name> creates a directory tree for a Claude Code
plugin or skill. Names must be kebab-case (lowercase letters, digits,
and single hyphens).

Supported kinds:

  plugin    Full plugin directory with .claude-plugin/plugin.json,
            a sample skill, LICENSE, and README.
  skill     Standalone SKILL.md under skills/<name>/.

Hook and agent scaffolding ships in v0.2.`,
	}
	cmd.AddCommand(newNewPluginCmd(stdout))
	cmd.AddCommand(newNewSkillCmd(stdout))
	return cmd
}

func newNewPluginCmd(stdout io.Writer) *cobra.Command {
	var flags newFlags
	cmd := &cobra.Command{
		Use:   "plugin <name>",
		Short: "Scaffold a Claude Code plugin directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(scaffold.KindPlugin, args[0], flags, stdout)
		},
	}
	flags.register(cmd)
	return cmd
}

func newNewSkillCmd(stdout io.Writer) *cobra.Command {
	var flags newFlags
	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Scaffold a standalone SKILL.md",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(scaffold.KindSkill, args[0], flags, stdout)
		},
	}
	flags.register(cmd)
	return cmd
}

func runNew(kind scaffold.Kind, name string, f newFlags, stdout io.Writer) error {
	opts := scaffold.Options{
		Kind:        kind,
		Name:        name,
		TargetDir:   f.into, // empty → scaffold.applyDefaults picks cwd
		Description: f.description,
		Author:      f.resolveAuthor(),
		Version:     f.version,
		License:     f.license,
		Attribution: !f.noAttribution,
		Force:       f.force,
	}
	r, err := scaffold.Scaffold(opts)
	if err != nil {
		if errors.Is(err, scaffold.ErrTargetExists) {
			return fmt.Errorf("%w: pass --force to overwrite or pick a different name", err)
		}
		return err
	}

	fmt.Fprintf(stdout, "Scaffolded %s at %s\n\n", kind, r.Root)
	for _, f := range r.Files {
		fmt.Fprintf(stdout, "  %s\n", f)
	}

	fmt.Fprintln(stdout)
	switch kind {
	case scaffold.KindPlugin:
		fmt.Fprintln(stdout, "Next steps:")
		fmt.Fprintf(stdout, "  cd %s\n", name)
		fmt.Fprintln(stdout, "  tsuba validate        # catch any schema issues now")
		fmt.Fprintln(stdout, "  <edit SKILL.md and README.md, commit, then submit to a marketplace>")
	case scaffold.KindSkill:
		fmt.Fprintln(stdout, "Next steps:")
		fmt.Fprintf(stdout, "  edit %s/SKILL.md and describe what the skill does\n", r.Root)
	}
	return nil
}
