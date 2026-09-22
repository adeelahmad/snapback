package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const versionPkg = "github.com/adeelahmad/snapback/internal/version"

func buildBinary(t *testing.T, ldflags string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "snapback")
	args := []string{"build", "-o", bin}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, ".")
	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return bin
}

func execBinary(bin string, args ...string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func TestBinaryVersionLdflags(t *testing.T) {
	ldflags := "-X " + versionPkg + ".Version=v1.2.3" +
		" -X " + versionPkg + ".Commit=abc1234" +
		" -X " + versionPkg + ".Target=linux/arm64"
	bin := buildBinary(t, ldflags)

	stdout, stderr, err := execBinary(bin, "version")

	if err != nil {
		t.Fatalf("snapback version: err = %v, want exit 0", err)
	}
	if want := "snapback v1.2.3 (commit abc1234, target linux/arm64)\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestBinaryDefaultsWithoutLdflags(t *testing.T) {
	bin := buildBinary(t, "")

	stdout, _, err := execBinary(bin, "version")

	if err != nil {
		t.Fatalf("snapback version: err = %v, want exit 0", err)
	}
	want := "snapback dev (commit none, target " + runtime.GOOS + "/" + runtime.GOARCH + ")"
	if !strings.Contains(stdout, want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout, want)
	}
}

func TestBinaryUsageExitCode(t *testing.T) {
	bin := buildBinary(t, "")

	cases := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"unknown subcommand", []string{"bogus"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := execBinary(bin, tc.args...)

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("err = %v, want *exec.ExitError with exit code 2", err)
			}
			if code := exitErr.ExitCode(); code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(stderr, usageLine) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, usageLine)
			}
		})
	}
}
