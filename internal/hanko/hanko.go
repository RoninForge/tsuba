// Package hanko adapts the Hanko CLI for tsuba. We shell out via
// os/exec rather than importing Hanko as a Go package because Hanko's
// validator lives in an internal/ path and we do not want to force
// Hanko to commit to a stable API surface just so Tsuba can call it.
// See docs/research/phase-1-spec.md section 8 for the reasoning.
package hanko

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

// ErrNotInstalled is returned when the hanko binary is not found on PATH.
// Callers (CLI, doctor) turn this into a clear user-facing message.
var ErrNotInstalled = errors.New("hanko is not installed or not on PATH")

// Available reports whether the hanko binary is on PATH.
func Available() bool {
	_, err := exec.LookPath("hanko")
	return err == nil
}

// Validate runs `hanko check <targetDir>` and returns exit code + stdout.
// When exit is non-zero, stdout contains the pretty error report; we do
// not parse JSON here (the CLI can ask Hanko for --json itself if it
// wants structured output). Returns ErrNotInstalled when hanko is
// missing, so the doctor / validate commands can print a friendly note.
func Validate(targetDir string) (exitCode int, stdout string, err error) {
	if !Available() {
		return 0, "", ErrNotInstalled
	}
	var out bytes.Buffer
	// G204: `targetDir` is a user-supplied file path, not an arbitrary
	// shell command. We pass it as an exec.Command arg (not through a
	// shell) so metacharacters cannot be reinterpreted.
	cmd := exec.Command("hanko", "check", "--color=false", targetDir) //nolint:gosec
	cmd.Stdout = &out
	cmd.Stderr = &out
	runErr := cmd.Run()
	if runErr == nil {
		return 0, out.String(), nil
	}
	var exit *exec.ExitError
	if errors.As(runErr, &exit) {
		return exit.ExitCode(), out.String(), nil
	}
	return -1, out.String(), fmt.Errorf("run hanko: %w", runErr)
}

// Version returns the `hanko version` string or empty if hanko is not
// installed. Used by `tsuba doctor`.
func Version() string {
	if !Available() {
		return ""
	}
	cmd := exec.Command("hanko", "version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(out))
}
