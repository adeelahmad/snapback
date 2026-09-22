//go:build integration

// Package acceptance drives the built snapback binary through the SPEC §20
// acceptance items. Every repo, root, HOME and mountpoint is disposable.
package acceptance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/adeelahmad/snapback/internal/compat/resticfx"
)

var (
	// snapbackBin is the binary under test, built once by TestMain or taken
	// from SNAPBACK_BIN.
	snapbackBin string
	// realHome is the invoking user's home, captured so tests can prove the
	// child process never sees it.
	realHome string
)

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "acceptance: home dir: %v\n", err)
		return 1
	}
	realHome = home

	if bin := os.Getenv("SNAPBACK_BIN"); bin != "" {
		snapbackBin = bin
		return m.Run()
	}
	dir, err := os.MkdirTemp("", "snapback-acceptance-bin-")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "acceptance: temp dir: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()
	snapbackBin = filepath.Join(dir, "snapback")
	build := exec.Command("go", "build", "-o", snapbackBin, "github.com/adeelahmad/snapback/cmd/snapback")
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "acceptance: build snapback: %v\n", err)
		return 1
	}
	return m.Run()
}

// requireFUSE skips a gated test unless every real-mount prerequisite is
// present, naming the first missing one.
func requireFUSE(t *testing.T) {
	t.Helper()
	p := resticfx.Probe{Getenv: os.Getenv, LookPath: exec.LookPath, Stat: os.Stat, GOOS: runtime.GOOS}
	if os.Getenv("SNAPBACK_FUSE_TESTS") == "1" {
		if _, err := exec.LookPath("restic"); err == nil {
			out, err := resticfx.ExecRunner{}.Run(context.Background(), "restic", []string{"version"})
			if err == nil {
				p.ResticVersionOut = string(out)
			}
		}
	}
	if msg := resticfx.MissingPrerequisite(p); msg != "" {
		skip(t, "missing prerequisite: "+msg)
	}
}
