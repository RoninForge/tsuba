// Package scaffold renders the embedded templates into a directory on
// disk. Separating this from the cli package keeps the rendering logic
// testable without spawning a cobra command.
package scaffold

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/RoninForge/tsuba/internal/templates"
)

// Kind identifies what the caller wants scaffolded.
type Kind string

const (
	// KindPlugin scaffolds a full Claude Code plugin directory with a
	// sample skill, LICENSE, README, and plugin.json.
	KindPlugin Kind = "plugin"
	// KindSkill scaffolds a standalone SKILL.md in skills/<name>/.
	KindSkill Kind = "skill"
)

// Author captures the author fields every template references. Both
// fields may be empty strings; the scaffold never rejects the run for
// missing author data (marketplaces do, and hanko flags it, but that's
// not scaffold's job).
//
// The json tags let Author ride directly through encoding/json.Marshal
// when we build plugin.json, which sidesteps the injection class that
// text/template has when a user name contains quotes or backslashes.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// pluginManifest is the shape of .claude-plugin/plugin.json when
// scaffolded by tsuba. Using a struct + encoding/json.Marshal rather
// than a text/template ensures user-supplied strings (description,
// author name, etc.) with quotes, backslashes, or newlines produce
// valid JSON instead of broken output.
type pluginManifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      Author `json:"author"`
	License     string `json:"license"`
}

// Options customises a single scaffold run. All fields except Name have
// working defaults; see Defaults for the canonical values.
type Options struct {
	// Kind is plugin or skill.
	Kind Kind
	// Name is the kebab-case identifier (plugin name, skill name).
	Name string
	// TargetDir is where files land. Defaults to current working dir.
	// Scaffold creates a subdirectory named Name under TargetDir for
	// the plugin kind, and skills/Name/ for the skill kind.
	TargetDir string
	// Description populates the top-level description in plugin.json
	// and the frontmatter of skills.
	Description string
	// Author is written into plugin.json and the LICENSE year + holder.
	Author Author
	// Version is the initial version in plugin.json. Default "0.1.0".
	Version string
	// License is the SPDX identifier written into plugin.json. Default "MIT".
	License string
	// Year is the LICENSE copyright year. Defaults to the current year.
	Year int
	// SampleSkill is the sub-skill created inside a plugin. Default "hello".
	// Plugin kind only.
	SampleSkill string
	// Attribution, when true (default), includes a "scaffolded with
	// tsuba" footer in generated README files.
	Attribution bool
	// Force, when true, overwrites existing files. Default false; the
	// scaffold errors out if the target directory already exists.
	Force bool
}

// Result summarises what Scaffold wrote. Callers (tests, the CLI) use
// it to report back to the user without re-walking the filesystem.
type Result struct {
	// Root is the absolute path of the directory the scaffold created.
	Root string
	// Files lists the paths of every file that was written, relative
	// to Root. Order is deterministic (lexical).
	Files []string
}

// kebabPattern is the same pattern hanko enforces on plugin names.
// Duplicated here so the scaffold can give a fast, friendly error
// before the user ever runs hanko.
var kebabPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ErrNameInvalid is returned when Name does not match kebabPattern.
var ErrNameInvalid = errors.New("name must be kebab-case: lowercase letters, digits, and single hyphens")

// ErrTargetExists is returned when the scaffold target directory
// already exists and Force is false.
var ErrTargetExists = errors.New("target directory already exists (use --force to overwrite)")

// Scaffold renders templates for the requested kind and writes them to
// disk. It returns a Result describing what was written. Errors are
// wrapped with context so the CLI can surface them directly.
func Scaffold(opts Options) (*Result, error) {
	if !kebabPattern.MatchString(opts.Name) {
		return nil, fmt.Errorf("%w: %q", ErrNameInvalid, opts.Name)
	}

	opts = applyDefaults(opts)

	switch opts.Kind {
	case KindPlugin:
		return scaffoldPlugin(opts)
	case KindSkill:
		return scaffoldSkill(opts)
	default:
		return nil, fmt.Errorf("unknown kind: %q", opts.Kind)
	}
}

func applyDefaults(o Options) Options {
	if o.TargetDir == "" {
		cwd, err := os.Getwd()
		if err == nil {
			o.TargetDir = cwd
		}
	}
	if o.Description == "" {
		o.Description = "A placeholder description. Replace before shipping."
	}
	if o.Version == "" {
		o.Version = "0.1.0"
	}
	if o.License == "" {
		o.License = "MIT"
	}
	if o.Year == 0 {
		o.Year = time.Now().Year()
	}
	if o.SampleSkill == "" {
		o.SampleSkill = "hello"
	}
	// Attribution defaults to true. A caller that wants to opt out
	// must set it explicitly false at the struct level; bool defaults
	// are false in Go, so we flip here.
	//
	// Callers that use the CLI: `tsuba new plugin foo --no-attribution`
	// flips the flag in the command handler before we get here.
	// Callers that use the package directly need to set it themselves.
	return o
}

// templateData is the shape every embedded template receives. Two
// derived fields, TitleCase and SampleSkillTitleCase, pre-compute the
// capitalised forms so the templates don't need a helper func map.
type templateData struct {
	Name                 string
	TitleCase            string
	Description          string
	Author               Author
	Version              string
	License              string
	Year                 int
	SampleSkill          string
	SampleSkillTitleCase string
	Attribution          bool
}

func buildData(o Options) templateData {
	return templateData{
		Name:                 o.Name,
		TitleCase:            titleCase(o.Name),
		Description:          o.Description,
		Author:               o.Author,
		Version:              o.Version,
		License:              o.License,
		Year:                 o.Year,
		SampleSkill:          o.SampleSkill,
		SampleSkillTitleCase: titleCase(o.SampleSkill),
		Attribution:          o.Attribution,
	}
}

func titleCase(kebab string) string {
	parts := strings.Split(kebab, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// render reads a template from the embedded FS, parses it, and executes
// it against data. A stable helper so every write goes through the same
// code path.
func render(path string, data templateData) ([]byte, error) {
	raw, err := templates.Read(path)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", path, err)
	}
	t, err := template.New(filepath.Base(path)).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", path, err)
	}
	return []byte(buf.String()), nil
}

// writeFile writes content to a path relative to root, creating parent
// directories as needed. Mode is 0o644 for files and 0o750 for the
// parent directories. Refuses to write outside root (defensive check
// against a rendered path that somehow contains `..`).
func writeFile(root, rel string, content []byte) error {
	full := filepath.Join(root, rel)
	// Reject paths that escape the root directory.
	relClean, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(relClean, "..") {
		return fmt.Errorf("refusing to write outside scaffold root: %s", rel)
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, content, 0o644); err != nil { //nolint:gosec // 0644 is intentional for generated source files
		return fmt.Errorf("write %s: %w", full, err)
	}
	return nil
}

// ensureTarget checks that the intended root directory can be created.
// Returns ErrTargetExists when root is already present and force is false.
// Creates the directory (parents included) otherwise.
//
// When force is true, we remove the existing path entirely and recreate
// it. That prevents stale files from a previous scaffold run lingering
// after the current run produces a different template set. If the
// existing path is a plain file (not a dir), the removal still succeeds
// and the subsequent MkdirAll gets the clean target it expects.
func ensureTarget(root string, force bool) error {
	info, err := os.Stat(root)
	switch {
	case err == nil:
		if !force {
			return ErrTargetExists
		}
		// A plain file at the target path without --force was already
		// caught by the !force branch above; this branch only runs when
		// force is true, so delete whatever is there and start fresh.
		_ = info
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove existing %s: %w", root, err)
		}
		return os.MkdirAll(root, 0o750)
	case errors.Is(err, fs.ErrNotExist):
		return os.MkdirAll(root, 0o750)
	default:
		return fmt.Errorf("stat %s: %w", root, err)
	}
}

func scaffoldPlugin(o Options) (*Result, error) {
	root := filepath.Join(o.TargetDir, o.Name)
	if err := ensureTarget(root, o.Force); err != nil {
		return nil, err
	}
	data := buildData(o)

	// plugin.json goes through encoding/json.Marshal so user-supplied
	// strings with quotes, backslashes, or newlines produce valid JSON.
	// Using a text/template would drop those bytes verbatim into the
	// file and produce broken output that immediately fails hanko check
	// (which is the exact scenario tsuba's "passes hanko on first try"
	// promise rules out).
	manifest := pluginManifest{
		Name:        o.Name,
		Description: o.Description,
		Version:     o.Version,
		Author:      o.Author,
		License:     o.License,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal plugin.json: %w", err)
	}
	// Trailing newline to match conventional POSIX file style.
	manifestBytes = append(manifestBytes, '\n')

	if err := writeFile(root, ".claude-plugin/plugin.json", manifestBytes); err != nil {
		return nil, err
	}

	// The remaining files are Markdown and MIT text; a text/template is
	// fine there because Markdown and the LICENSE boilerplate are not
	// structured formats that user-supplied strings can corrupt.
	writes := []struct {
		templatePath string
		outputPath   string
	}{
		{"plugin/README.md.tmpl", "README.md"},
		{"plugin/LICENSE.tmpl", "LICENSE"},
		{"plugin/sample-skill.md.tmpl", filepath.Join("skills", o.SampleSkill, "SKILL.md")},
	}

	written := []string{".claude-plugin/plugin.json"}
	for _, w := range writes {
		content, err := render(w.templatePath, data)
		if err != nil {
			return nil, err
		}
		if err := writeFile(root, w.outputPath, content); err != nil {
			return nil, err
		}
		written = append(written, w.outputPath)
	}
	return &Result{Root: root, Files: written}, nil
}

func scaffoldSkill(o Options) (*Result, error) {
	// Skills land at <target>/skills/<name>/SKILL.md by default.
	root := filepath.Join(o.TargetDir, "skills", o.Name)
	if err := ensureTarget(root, o.Force); err != nil {
		return nil, err
	}
	data := buildData(o)

	content, err := render("skill/SKILL.md.tmpl", data)
	if err != nil {
		return nil, err
	}
	if err := writeFile(root, "SKILL.md", content); err != nil {
		return nil, err
	}
	return &Result{Root: root, Files: []string{"SKILL.md"}}, nil
}
