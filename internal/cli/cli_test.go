package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(args ...string) (stdout, stderr string, code int) {
	var out, errb bytes.Buffer
	code = Execute(&out, &errb, args)
	return out.String(), errb.String(), code
}

func TestVersionCommand(t *testing.T) {
	stdout, _, code := runCLI("version")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if stdout == "" {
		t.Error("version should write something to stdout")
	}
}

func TestHelpCommand(t *testing.T) {
	stdout, _, code := runCLI("--help")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	for _, want := range []string{"tsuba", "new", "validate", "doctor", "list"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestNewPluginHappyPath(t *testing.T) {
	tmp := t.TempDir()
	stdout, _, code := runCLI("new", "plugin", "my-plugin",
		"--into", tmp,
		"--description", "a test plugin",
		"--author", "Test User",
		"--email", "test@example.com")
	if code != 0 {
		t.Errorf("exit = %d, want 0. stdout: %s", code, stdout)
	}
	if !strings.Contains(stdout, "Scaffolded plugin") {
		t.Errorf("stdout should announce scaffold, got: %s", stdout)
	}
	// Spot-check a generated file.
	pluginJSON, err := os.ReadFile(filepath.Join(tmp, "my-plugin", ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatalf("plugin.json not written: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(pluginJSON, &m); err != nil {
		t.Fatalf("plugin.json invalid JSON: %v\n%s", err, pluginJSON)
	}
	if m["name"] != "my-plugin" {
		t.Errorf("plugin.json name = %v, want my-plugin", m["name"])
	}
}

func TestNewSkillHappyPath(t *testing.T) {
	tmp := t.TempDir()
	stdout, _, code := runCLI("new", "skill", "code-reviewer",
		"--into", tmp,
		"--description", "Review code for quality")
	if code != 0 {
		t.Errorf("exit = %d, want 0. stdout: %s", code, stdout)
	}
	skill, err := os.ReadFile(filepath.Join(tmp, "skills", "code-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatalf("SKILL.md not written: %v", err)
	}
	for _, want := range []string{
		"name: code-reviewer",
		"description: Review code for quality",
		"# Code Reviewer",
		"## When to use",
	} {
		if !strings.Contains(string(skill), want) {
			t.Errorf("SKILL.md missing %q", want)
		}
	}
}

func TestNewRejectsBadName(t *testing.T) {
	_, _, code := runCLI("new", "plugin", "BadName", "--into", t.TempDir())
	if code == 0 {
		t.Error("PascalCase name should fail")
	}
}

func TestNewWithoutArgsFails(t *testing.T) {
	_, _, code := runCLI("new", "plugin")
	if code == 0 {
		t.Error("missing name arg should fail")
	}
}

func TestNewRefusesExistingDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "taken"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, code := runCLI("new", "plugin", "taken", "--into", tmp)
	if code == 0 {
		t.Error("existing dir without --force should fail")
	}
}

func TestNewForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "forceme"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, code := runCLI("new", "plugin", "forceme",
		"--into", tmp,
		"--force",
		"--author", "X",
		"--email", "x@x.x")
	if code != 0 {
		t.Errorf("--force should allow overwrite, got %d", code)
	}
}

func TestNoAttributionFlag(t *testing.T) {
	tmp := t.TempDir()
	_, _, code := runCLI("new", "plugin", "noattr",
		"--into", tmp,
		"--author", "X",
		"--email", "x@x.x",
		"--no-attribution")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	readme, _ := os.ReadFile(filepath.Join(tmp, "noattr", "README.md"))
	if strings.Contains(string(readme), "Scaffolded with") {
		t.Errorf("--no-attribution should drop the footer, got:\n%s", readme)
	}
}

func TestDoctor(t *testing.T) {
	stdout, _, code := runCLI("doctor")
	if code != 0 {
		t.Errorf("doctor exit = %d", code)
	}
	// Output should mention hanko and git whether they're installed or not.
	for _, want := range []string{"hanko", "git"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("doctor output missing %q mention", want)
		}
	}
}

func TestListMarketplaces(t *testing.T) {
	stdout, _, code := runCLI("list", "marketplaces")
	if code != 0 {
		t.Errorf("exit = %d", code)
	}
	for _, want := range []string{"anthropic", "buildwithclaude", "cc-marketplace", "claudemarketplaces"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("marketplaces list missing %q", want)
		}
	}
}

func TestValidateWithoutHanko(t *testing.T) {
	// Path the test-runner's PATH to a tmpdir with no hanko binary.
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", oldPath)

	_, stderr, code := runCLI("validate", t.TempDir())
	if code == 0 {
		t.Error("validate without hanko should fail with non-zero exit")
	}
	if !strings.Contains(stderr, "hanko is required") && !strings.Contains(stderr, "not found on PATH") {
		t.Errorf("stderr should guide user to install hanko, got: %s", stderr)
	}
}

func TestUnknownSubcommand(t *testing.T) {
	_, _, code := runCLI("no-such-command")
	if code == 0 {
		t.Error("unknown subcommand should fail")
	}
}
