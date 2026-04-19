// Package gitctx reads git configuration values so the scaffold can
// auto-populate author fields when the user does not pass explicit flags.
// Failures are soft: a missing git config just returns empty strings and
// lets the caller fall back to placeholders or prompts.
package gitctx

import (
	"errors"
	"os/exec"
	"strings"
)

// UserName returns `git config user.name`, or empty string if unset or
// git is not on PATH.
func UserName() string { return configValue("user.name") }

// UserEmail returns `git config user.email`, or empty string if unset or
// git is not on PATH.
func UserEmail() string { return configValue("user.email") }

// Available reports whether `git` exists on PATH. Useful for doctor
// checks that want to distinguish "git not installed" from "git config
// unset."
func Available() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func configValue(key string) string {
	// G204: `key` is a compile-time constant passed by sibling functions
	// (UserName, UserEmail); never user input.
	cmd := exec.Command("git", "config", "--get", key) //nolint:gosec
	out, err := cmd.Output()
	if err != nil {
		// exit status 1 = key not set. Any other error means git is
		// missing or something weirder; the caller doesn't care which
		// in the scaffold path - empty string is a safe fallback.
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return ""
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}
