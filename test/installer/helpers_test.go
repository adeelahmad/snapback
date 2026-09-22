package installer

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const installScript = "../../install.sh"

// runInstaller runs `sh install.sh` with a minimal environment (PATH, HOME,
// a temp TMPDIR) plus the given overrides, so the host's SNAPBACK_* vars
// never leak into a test.
func runInstaller(t *testing.T, env map[string]string) (stdout, stderr string, exitCode int) {
	t.Helper()

	script, err := filepath.Abs(installScript)
	if err != nil {
		t.Fatalf("resolve install.sh: %v", err)
	}

	cmd := exec.Command("sh", script)
	cmd.Dir = filepath.Dir(script)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"TMPDIR=" + t.TempDir(),
	}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
		exitCode = 0
	case errors.As(runErr, &exitErr):
		exitCode = exitErr.ExitCode()
	default:
		t.Fatalf("run sh install.sh: %v", runErr)
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func dryRunEnv(osName, arch string) map[string]string {
	return map[string]string{
		"SNAPBACK_DRY_RUN": "1",
		"SNAPBACK_OS":      osName,
		"SNAPBACK_ARCH":    arch,
	}
}
