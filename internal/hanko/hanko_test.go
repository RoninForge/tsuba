package hanko

import (
	"os"
	"testing"
)

// TestAvailableAndValidateWithoutHanko isolates PATH so the hanko
// binary is NOT found, then verifies Available/Validate/Version all
// behave correctly on a system without hanko installed.
func TestAvailableAndValidateWithoutHanko(t *testing.T) {
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", oldPath)

	if Available() {
		t.Error("Available should be false when hanko is not on PATH")
	}
	if Version() != "" {
		t.Error("Version should be empty when hanko is not on PATH")
	}
	exitCode, out, err := Validate("/tmp")
	if err == nil || err != ErrNotInstalled {
		t.Errorf("Validate should return ErrNotInstalled, got: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitCode on missing hanko = %d, want 0", exitCode)
	}
	if out != "" {
		t.Errorf("out on missing hanko should be empty, got %q", out)
	}
}
