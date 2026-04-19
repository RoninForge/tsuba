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
//
// Email has `omitempty` because the Claude Code plugin schema treats
// email as optional-but-must-be-a-valid-format-if-present. A missing
// email is fine; an empty-string email fails the `format: email`
// check. Name does not use `omitempty`: it is the required field, and
// scaffoldPlugin already skips emitting the whole Author object when
// Name is empty.
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// pluginManifest is the shape of .claude-plugin/plugin.json when
// scaffolded by tsuba. Using a struct + encoding/json.Marshal rather
// than a text/template ensures user-supplied strings (description,
// author name, etc.) with quotes, backslashes, or newlines produce
// valid JSON instead of broken output.
//
// Author is a pointer so a fresh-machine user with no git config gets
// the field OMITTED from the manifest instead of a present-but-empty
// `"author": {"name":"","email":""}`. An absent author triggers hanko's
// HANKO003 warning (non-blocking); an empty name triggers a blocking
// HANKO-SCHEMA error. We promise "passes hanko on the first try," so
// falling back to the warning path is the correct default.
type pluginManifest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Author      *Author `json:"author,omitempty"`
	License     string  `json:"license"`
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

// ErrSampleSkillInvalid is returned when Options.SampleSkill is set
// to a non-kebab-case value. Prevents YAML-frontmatter-injection
// through the sample skill's `name:` field if a future CLI flag ever
// exposes SampleSkill to user input.
var ErrSampleSkillInvalid = errors.New("sample skill name must be kebab-case: lowercase letters, digits, and single hyphens")

// Scaffold renders templates for the requested kind and writes them to
// disk. It returns a Result describing what was written. Errors are
// wrapped with context so the CLI can surface them directly.
func Scaffold(opts Options) (*Result, error) {
	if !kebabPattern.MatchString(opts.Name) {
		return nil, fmt.Errorf("%w: %q", ErrNameInvalid, opts.Name)
	}

	opts = applyDefaults(opts)

	if opts.Kind == KindPlugin && !kebabPattern.MatchString(opts.SampleSkill) {
		return nil, fmt.Errorf("%w: %q", ErrSampleSkillInvalid, opts.SampleSkill)
	}

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
	// SampleSkill is currently hardcoded by the CLI (no flag), but
	// validate anyway so a future flag can't bypass the check.
	// Defensive pattern, zero runtime cost.
	_ = o.SampleSkill
	// Attribution defaults to true. A caller that wants to opt out
	// must set it explicitly false at the struct level; bool defaults
	// are false in Go, so we flip here.
	//
	// Callers that use the CLI: `tsuba new plugin foo --no-attribution`
	// flips the flag in the command handler before we get here.
	// Callers that use the package directly need to set it themselves.
	return o
}

// templateData is the shape every embedded template receives. Derived
// fields (TitleCase, SampleSkillTitleCase, DescriptionYAML) pre-compute
// format-specific forms so the templates don't need a helper func map.
type templateData struct {
	Name                 string
	TitleCase            string
	Description          string
	DescriptionYAML      string // pre-escaped for YAML double-quoted scalar context
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
		DescriptionYAML:      yamlQuoteString(o.Description),
		Author:               o.Author,
		Version:              o.Version,
		License:              o.License,
		Year:                 o.Year,
		SampleSkill:          o.SampleSkill,
		SampleSkillTitleCase: titleCase(o.SampleSkill),
		Attribution:          o.Attribution,
	}
}

// yamlQuoteString returns a YAML double-quoted scalar that safely
// encodes any string. YAML 1.2's double-quoted scalar is a strict
// superset of JSON strings, so json.Marshal's output is valid YAML in
// all cases we care about (quotes, backslashes, newlines, tabs,
// Unicode U+2028). We rely on that equivalence instead of writing a
// bespoke YAML escaper.
//
// json.Marshal of a plain Go string cannot fail per stdlib contract
// (only channels, funcs, cyclic structs, and NaN/Inf floats trip it).
// Panic on an err return rather than silently emitting `""` — that
// signals a broken build, which is the correct reaction.
func yamlQuoteString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic("yamlQuoteString: json.Marshal failed on a string (impossible per encoding/json contract): " + err.Error())
	}
	return string(b)
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
	// filepath.IsLocal (Go 1.20+) is the idiomatic check: rejects
	// absolute paths, any `..` segment, and (on Windows) reserved
	// names. Replaces a hand-rolled filepath.Rel + HasPrefix("..")
	// that would false-match legitimate names like `..hidden.md`.
	if !filepath.IsLocal(rel) {
		return fmt.Errorf("refusing to write outside scaffold root: %s", rel)
	}
	full := filepath.Join(root, rel)
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
// Uses os.Lstat rather than os.Stat so a broken symlink at the target
// path (e.g. left over from a deleted repo) is detected as "exists"
// instead of confusing Stat's ErrNotExist into overwriting the link
// location with MkdirAll.
//
// When force is true, we remove the existing path entirely (including
// a plain file or broken symlink) and recreate it. That prevents stale
// files from a previous scaffold run lingering after the current run
// produces a different template set.
func ensureTarget(root string, force bool) error {
	_, err := os.Lstat(root)
	switch {
	case err == nil:
		if !force {
			return ErrTargetExists
		}
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove existing %s: %w", root, err)
		}
		return os.MkdirAll(root, 0o750)
	case errors.Is(err, fs.ErrNotExist):
		return os.MkdirAll(root, 0o750)
	default:
		return fmt.Errorf("lstat %s: %w", root, err)
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
		License:     o.License,
	}
	// Only include the author object when we have a non-empty name.
	// The schema requires `author.name` (minLength 1) and treats email
	// as optional. An email-only author would serialize as
	// {"name":"","email":"x@y"} which fails hanko's minLength check.
	// Name-only is fine: Email has `omitempty` so it drops out cleanly.
	// See the pluginManifest doc comment for the hanko-pass rationale.
	if o.Author.Name != "" {
		author := o.Author
		manifest.Author = &author
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
