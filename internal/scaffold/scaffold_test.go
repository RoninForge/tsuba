package scaffold

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
		"description: Review code for quality",
		"# Code Reviewer",
		"## When to use",
		"## Instructions",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("SKILL.md missing %q. Full content:\n%s", want, s)
		}
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
