package scaffold

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestTitleCase guards the small helper that turns kebab-case plugin
// names into the README heading and skill title.
func TestTitleCase(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "Hello"},
		{"hello-world", "Hello World"},
		{"my-cool-plugin", "My Cool Plugin"},
		{"a", "A"},
		{"abc-123", "Abc 123"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := titleCase(tt.in); got != tt.want {
				t.Errorf("titleCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestScaffoldRejectsInvalidName ensures kebab-case validation fails
// fast before the scaffold tries to write anything.
func TestScaffoldRejectsInvalidName(t *testing.T) {
	for _, bad := range []string{"", "PascalCase", "under_score", "has space", "UPPER", "-leading"} {
		_, err := Scaffold(Options{Kind: KindPlugin, Name: bad, TargetDir: t.TempDir()})
		if err == nil {
			t.Errorf("expected rejection for %q, got nil", bad)
			continue
		}
		if !errors.Is(err, ErrNameInvalid) {
			t.Errorf("expected ErrNameInvalid for %q, got %v", bad, err)
		}
	}
}

// TestScaffoldPluginHappyPath checks the core scaffolder output against
// the expected file set, and verifies the generated plugin.json is
// valid JSON with the right shape.
func TestScaffoldPluginHappyPath(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:        KindPlugin,
		Name:        "my-plugin",
		TargetDir:   tmp,
		Description: "a test plugin",
		Author:      Author{Name: "Test User", Email: "test@example.com"},
		Attribution: true,
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}

	wantFiles := []string{
		".claude-plugin/plugin.json",
		"README.md",
		"LICENSE",
		"skills/hello/SKILL.md",
	}
	if len(r.Files) != len(wantFiles) {
		t.Errorf("got %d files, want %d: %v", len(r.Files), len(wantFiles), r.Files)
	}
	for _, f := range wantFiles {
		full := filepath.Join(r.Root, f)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected file %s missing: %v", f, err)
		}
	}

	// Parse plugin.json and assert key fields.
	raw, err := os.ReadFile(filepath.Join(r.Root, ".claude-plugin/plugin.json"))
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v\n%s", err, raw)
	}
	if m["name"] != "my-plugin" {
		t.Errorf("plugin.json name = %v, want my-plugin", m["name"])
	}
	if m["description"] != "a test plugin" {
		t.Errorf("plugin.json description = %v, want \"a test plugin\"", m["description"])
	}
	if m["version"] != "0.1.0" {
		t.Errorf("plugin.json version = %v, want 0.1.0", m["version"])
	}
	author, ok := m["author"].(map[string]any)
	if !ok {
		t.Fatalf("plugin.json author is not an object: %v", m["author"])
	}
	if author["name"] != "Test User" {
		t.Errorf("author.name = %v, want \"Test User\"", author["name"])
	}

	// README should include the title case and attribution footer.
	readme, _ := os.ReadFile(filepath.Join(r.Root, "README.md"))
	if !strings.Contains(string(readme), "# My Plugin") {
		t.Errorf("README missing title-case heading:\n%s", readme)
	}
	if !strings.Contains(string(readme), "Scaffolded with [tsuba]") {
		t.Errorf("README missing attribution footer:\n%s", readme)
	}

	// LICENSE should include the author name.
	license, _ := os.ReadFile(filepath.Join(r.Root, "LICENSE"))
	if !strings.Contains(string(license), "Test User") {
		t.Errorf("LICENSE missing author name:\n%s", license)
	}
}

// TestScaffoldPluginNoAttribution verifies the --no-attribution flag
// path drops the footer cleanly (no orphan horizontal rule).
func TestScaffoldPluginNoAttribution(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:        KindPlugin,
		Name:        "no-attr",
		TargetDir:   tmp,
		Author:      Author{Name: "X", Email: "x@x.x"},
		Attribution: false,
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	readme, _ := os.ReadFile(filepath.Join(r.Root, "README.md"))
	if strings.Contains(string(readme), "Scaffolded with") {
		t.Errorf("README still has attribution with Attribution=false:\n%s", readme)
	}
}

// TestScaffoldRefusesExistingDir is the safety check that prevents
// blindly overwriting someone's work.
func TestScaffoldRefusesExistingDir(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "taken")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Scaffold(Options{Kind: KindPlugin, Name: "taken", TargetDir: tmp})
	if !errors.Is(err, ErrTargetExists) {
		t.Errorf("expected ErrTargetExists, got %v", err)
	}
}

// TestScaffoldForceOverwrites confirms --force bypasses the existing
// directory check (and the subsequent writes land).
func TestScaffoldForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "forceme")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "forceme",
		TargetDir: tmp,
		Force:     true,
		Author:    Author{Name: "X", Email: "x@x.x"},
	}); err != nil {
		t.Fatalf("Scaffold with Force=true: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude-plugin/plugin.json")); err != nil {
		t.Errorf("plugin.json should have been written, got: %v", err)
	}
}

// TestScaffoldSkill exercises the standalone skill path.
func TestScaffoldSkill(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:        KindSkill,
		Name:        "code-reviewer",
		TargetDir:   tmp,
		Description: "Review code for quality",
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	// Root is <target>/skills/<name>/.
	wantRoot := filepath.Join(tmp, "skills", "code-reviewer")
	if r.Root != wantRoot {
		t.Errorf("r.Root = %q, want %q", r.Root, wantRoot)
	}
	want := filepath.Join(r.Root, "SKILL.md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected SKILL.md at %s, got: %v", want, err)
	}
	content, _ := os.ReadFile(want)
	s := string(content)
	for _, want := range []string{
		"name: code-reviewer",
		`description: "Review code for quality"`, // YAML double-quoted scalar per T2-2 fix
		"# Code Reviewer",
		"## When to use",
		"## Instructions",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("SKILL.md missing %q. Full content:\n%s", want, s)
		}
	}
}

// TestEnsureTargetWithBrokenSymlink covers the Round 2 os.Lstat fix.
// A dangling symlink at the target path must be detected as "exists"
// (not as "nothing there"), so --force can clean it up via RemoveAll
// instead of crashing later in MkdirAll.
//
// Only runs on non-Windows because os.Symlink on Windows requires
// admin or developer-mode. CI covers Linux + macOS which is where the
// bug lives anyway.
func TestEnsureTargetWithBrokenSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Symlink on Windows needs admin; the Lstat fix is irrelevant there")
	}
	tmp := t.TempDir()
	root := filepath.Join(tmp, "link-target")
	if err := os.Symlink("/does/not/exist-broken-symlink", root); err != nil {
		t.Fatalf("create dangling symlink: %v", err)
	}

	// Without --force, dangling link must be detected as existing.
	_, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "link-target",
		TargetDir: tmp,
	})
	if !errors.Is(err, ErrTargetExists) {
		t.Errorf("without --force, dangling symlink should return ErrTargetExists, got: %v", err)
	}

	// With --force, scaffold must clean the link and proceed.
	if _, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "link-target",
		TargetDir: tmp,
		Force:     true,
		Author:    Author{Name: "X"},
	}); err != nil {
		t.Errorf("with --force, dangling symlink should be replaced with a real dir, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude-plugin/plugin.json")); err != nil {
		t.Errorf("plugin.json should exist after --force over a broken link, got: %v", err)
	}
}

// TestSampleSkillMustBeKebabCase is defensive: SampleSkill is not a
// CLI flag today, but if it ever becomes one, a non-kebab-case value
// would flow into YAML frontmatter in sample-skill.md.tmpl and reopen
// the same injection class the T2-2 fix closed for description.
func TestSampleSkillMustBeKebabCase(t *testing.T) {
	_, err := Scaffold(Options{
		Kind:        KindPlugin,
		Name:        "valid-name",
		TargetDir:   t.TempDir(),
		SampleSkill: "Invalid Skill Name",
		Author:      Author{Name: "X"},
	})
	if !errors.Is(err, ErrSampleSkillInvalid) {
		t.Errorf("expected ErrSampleSkillInvalid for non-kebab SampleSkill, got: %v", err)
	}
}

// TestScaffoldUnknownKind guards the switch default.
func TestScaffoldUnknownKind(t *testing.T) {
	_, err := Scaffold(Options{Kind: "bogus", Name: "x", TargetDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for unknown kind, got nil")
	}
	if !strings.Contains(err.Error(), "unknown kind") {
		t.Errorf("error should mention unknown kind, got: %v", err)
	}
}

// TestPluginJSONInjection is the regression guard for the Round 1
// review finding: before this test, a user whose name or description
// contained a JSON-sensitive character (`"`, `\`, `\n`, `\t`, Unicode
// line separators) would produce a broken plugin.json that failed
// `hanko check` on the first try. The fix routes plugin.json through
// encoding/json.Marshal; every round-trip below must re-parse into the
// original string.
func TestPluginJSONInjection(t *testing.T) {
	cases := []struct {
		name        string
		description string
		authorName  string
	}{
		{"quote in author", "plain", `Jose "Pepe" Lopez`},
		{"backslash in description", `C:\Users\home`, "plain"},
		{"newline in description", "line one\nline two", "plain"},
		{"tab in description", "with\ttab", "plain"},
		{"unicode line separator", "text\u2028more", "plain"},
		{"injection attempt", `","version":"666","license":"pwn`, "plain"},
		{"quote in email", "plain", "plain"}, // email tested separately below
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			r, err := Scaffold(Options{
				Kind:        KindPlugin,
				Name:        "injection-test",
				TargetDir:   tmp,
				Description: tt.description,
				Author:      Author{Name: tt.authorName, Email: "test@example.com"},
			})
			if err != nil {
				t.Fatalf("Scaffold: %v", err)
			}
			raw, err := os.ReadFile(filepath.Join(r.Root, ".claude-plugin/plugin.json"))
			if err != nil {
				t.Fatalf("read plugin.json: %v", err)
			}
			var got pluginManifest
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("plugin.json is not valid JSON: %v\n%s", err, raw)
			}
			if got.Description != tt.description {
				t.Errorf("description mismatch\n  want: %q\n  got:  %q\n  raw:  %s", tt.description, got.Description, raw)
			}
			if got.Author.Name != tt.authorName {
				t.Errorf("author.name mismatch\n  want: %q\n  got:  %q", tt.authorName, got.Author.Name)
			}
		})
	}
}

// TestForceClearsStaleFiles verifies that --force removes existing
// files from a previous scaffold before writing the new set. Without
// this, orphaned files from earlier tsuba versions linger after an
// upgrade-and-rescaffold cycle.
func TestForceClearsStaleFiles(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "clearme")
	if err := os.MkdirAll(root, 0o750); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(root, "stale-from-earlier-scaffold.md")
	if err := os.WriteFile(stale, []byte("leftover"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "clearme",
		TargetDir: tmp,
		Force:     true,
		Author:    Author{Name: "X", Email: "x@x.x"},
	}); err != nil {
		t.Fatalf("Scaffold with Force=true: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("expected stale file to be removed, stat err = %v", err)
	}
}

// TestPluginJSONOmitsAuthorWhenEmpty is the Round 2 T2-1 guard:
// a fresh-machine user with no git config used to produce a
// plugin.json with `"author": {"name":"","email":""}` that hanko
// rejected with a HANKO-SCHEMA minLength error. The fix makes Author
// a pointer and omits it entirely when both fields are empty, which
// downgrades the hanko signal to HANKO003 (warning, non-blocking).
func TestPluginJSONOmitsAuthorWhenEmpty(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "no-author-plugin",
		TargetDir: tmp,
		Author:    Author{}, // both fields empty
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(r.Root, ".claude-plugin/plugin.json"))
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	if strings.Contains(string(raw), `"author"`) {
		t.Errorf("author field should be omitted when both Name and Email are empty, got:\n%s", raw)
	}

	// Also assert the manifest is still a valid JSON object.
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("plugin.json still needs to be valid JSON: %v", err)
	}
	if _, hasAuthor := got["author"]; hasAuthor {
		t.Error("author key present in decoded map; should be absent")
	}
}

// TestPluginJSONEmailOnlyOmitsAuthor is the Round 3 guard. An
// email-only Author (name empty) used to still emit the author
// object, which breaks hanko's minLength:1 check on author.name.
// The schema treats email as optional; the whole object must be
// absent when name is empty.
func TestPluginJSONEmailOnlyOmitsAuthor(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "email-only",
		TargetDir: tmp,
		Author:    Author{Email: "anon@example.com"},
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(r.Root, ".claude-plugin/plugin.json"))
	if strings.Contains(string(raw), `"author"`) {
		t.Errorf("author object must be omitted when Name is empty (email-only). raw:\n%s", raw)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, hasAuthor := got["author"]; hasAuthor {
		t.Error("author key should be absent from decoded map")
	}
}

// TestPluginJSONNameOnlyKeepsAuthorOmitsEmail is the Round 3 guard
// for the opposite partial case. Name is required per schema; email
// is optional. A name-only author should serialize as
// {"name":"X"} without an empty email field (thanks to omitempty).
func TestPluginJSONNameOnlyKeepsAuthorOmitsEmail(t *testing.T) {
	tmp := t.TempDir()
	r, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "name-only",
		TargetDir: tmp,
		Author:    Author{Name: "Anonymous"},
	})
	if err != nil {
		t.Fatalf("Scaffold: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(r.Root, ".claude-plugin/plugin.json"))
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	author, ok := got["author"].(map[string]any)
	if !ok {
		t.Fatalf("author should be emitted when Name is set, got %v", got["author"])
	}
	if author["name"] != "Anonymous" {
		t.Errorf("author.name = %v, want Anonymous", author["name"])
	}
	if _, hasEmail := author["email"]; hasEmail {
		t.Error("author.email should be omitted (omitempty) when the input email is empty")
	}
}

// TestSkillYAMLInjection is the Round 2 T2-2 guard: the standalone
// skill template rendered `description: {{.Description}}` into YAML
// frontmatter. A description containing newlines, YAML control chars
// (`:` followed by space, `|`, `>`, `#`, `[`), or the literal text of
// a YAML alias produced broken frontmatter that Claude Code could not
// parse. The fix routes description through a JSON-as-YAML-double-
// quoted-scalar helper; every input below must round-trip.
func TestSkillYAMLInjection(t *testing.T) {
	cases := []struct {
		name        string
		description string
	}{
		{"plain", "Review code for quality issues"},
		{"yaml block scalar marker", "| yaml block scalar marker"},
		{"yaml comment marker", "# looks like a comment"},
		{"yaml flow sequence", "[hello, world]"},
		{"colon space pair", "key: value"},
		{"newline", "line one\nline two"},
		{"double quote", `has a "quote" in it`},
		{"backslash", `C:\path\to\thing`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			r, err := Scaffold(Options{
				Kind:        KindSkill,
				Name:        "yaml-inject-test",
				TargetDir:   tmp,
				Description: tt.description,
			})
			if err != nil {
				t.Fatalf("Scaffold: %v", err)
			}
			raw, err := os.ReadFile(filepath.Join(r.Root, "SKILL.md"))
			if err != nil {
				t.Fatalf("read SKILL.md: %v", err)
			}
			// Frontmatter must be three lines: opening ---,
			// name: yaml-inject-test, description: "...", closing ---.
			// We don't parse YAML here (no import), but we assert the
			// description line starts with a double-quoted scalar and
			// contains a JSON-encoded form of the input.
			expected := `description: ` + yamlQuoteString(tt.description)
			if !strings.Contains(string(raw), expected) {
				t.Errorf("SKILL.md does not contain expected double-quoted description line.\nwant line: %q\nraw:\n%s", expected, raw)
			}
			// Also assert there is exactly one description: line (a
			// broken YAML injection could produce multiple).
			if count := strings.Count(string(raw), "\ndescription:"); count > 1 {
				t.Errorf("SKILL.md should contain exactly 1 `description:` line in frontmatter, got %d. raw:\n%s", count, raw)
			}
		})
	}
}

// TestForceWorksWhenTargetIsPlainFile ensures --force removes a
// non-directory at the target path instead of crashing with the
// confusing "mkdir: not a directory" message Round 1 surfaced.
func TestForceWorksWhenTargetIsPlainFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "was-a-file")
	if err := os.WriteFile(target, []byte("collision"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scaffold(Options{
		Kind:      KindPlugin,
		Name:      "was-a-file",
		TargetDir: tmp,
		Force:     true,
		Author:    Author{Name: "X", Email: "x@x.x"},
	}); err != nil {
		t.Fatalf("Scaffold should clear plain-file collision under --force, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".claude-plugin/plugin.json")); err != nil {
		t.Errorf("expected plugin.json written after --force over a file, got: %v", err)
	}
}
