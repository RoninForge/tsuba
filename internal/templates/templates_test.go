package templates

import (
	"io/fs"
	"testing"
)

func TestFSHasTemplates(t *testing.T) {
	var count int
	err := fs.WalkDir(FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk FS: %v", err)
	}
	if count < 4 {
		t.Errorf("expected at least 4 templates (skill, plugin.json, plugin README, plugin LICENSE, sample-skill), got %d", count)
	}
}

func TestReadKnownTemplates(t *testing.T) {
	for _, path := range []string{
		"skill/SKILL.md.tmpl",
		"plugin/plugin.json.tmpl",
		"plugin/README.md.tmpl",
		"plugin/LICENSE.tmpl",
		"plugin/sample-skill.md.tmpl",
	} {
		data, err := Read(path)
		if err != nil {
			t.Errorf("Read(%q): %v", path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("Read(%q) returned empty bytes", path)
		}
	}
}

func TestReadMissing(t *testing.T) {
	if _, err := Read("does/not/exist.tmpl"); err == nil {
		t.Error("Read of missing path should return an error")
	}
}
