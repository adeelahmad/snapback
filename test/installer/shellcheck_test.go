package installer

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestShellcheck(t *testing.T) {
	bin, err := exec.LookPath("shellcheck")
	if err != nil {
		t.Skip("shellcheck not installed; CI runs it via S1-02")
	}

	out, runErr := exec.Command(bin, "-s", "sh", installScript).CombinedOutput()

	exitCode := 0
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
	case errors.As(runErr, &exitErr):
		exitCode = exitErr.ExitCode()
	default:
		t.Fatalf("run shellcheck: %v", runErr)
	}
	if exitCode != 0 {
		t.Errorf("shellcheck -s sh install.sh exit = %d, want 0", exitCode)
	}
	if got := strings.TrimSpace(string(out)); got != "" {
		t.Errorf("shellcheck output not empty:\n%s", got)
	}
}
