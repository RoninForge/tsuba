// Package templates embeds the text/template sources tsuba renders when
// scaffolding a new skill, plugin, hook, or agent. Keeping them in a
// dedicated package makes the template set a single, testable artifact
// and avoids scattering the scaffold logic across sibling packages.
package templates

import (
	"embed"
	"io/fs"
)

//go:embed tmpl/*
var raw embed.FS

// FS returns a read-only filesystem rooted at the `tmpl/` directory.
// Callers do not need to know the embed prefix.
func FS() fs.FS {
	sub, err := fs.Sub(raw, "tmpl")
	if err != nil {
		// This can only happen if tmpl/ is renamed or removed; the
		// embed directive makes the source tree authoritative. Panic
		// is the honest reaction: a mis-built binary should not ship.
		panic("templates: tmpl/ subdir missing: " + err.Error())
	}
	return sub
}

// Read returns the bytes of a single template file relative to tmpl/.
// Returns the fs.ErrNotExist from io/fs on a missing path.
func Read(path string) ([]byte, error) {
	return fs.ReadFile(FS(), path)
}
