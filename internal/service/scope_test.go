package service

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// assertSystemScopeRefused checks that a command run with --scope system was
// refused: exit 1, an unsupported_service_manager error naming user scope,
// no systemctl call and no unit file anywhere under the fixture root.
func assertSystemScopeRefused(t *testing.T, f *commandFixture, args []string, code int) {
	t.Helper()
	if code != 1 {
		t.Errorf("%q exit = %d, want 1", args, code)
	}
	stderr := f.stderr.String()
	if !strings.Contains(stderr, "unsupported_service_manager") {
		t.Errorf("%q stderr = %q, want it to contain %q", args, stderr, "unsupported_service_manager")
	}
	if !strings.Contains(strings.ToLower(stderr), "user scope") {
		t.Errorf("%q stderr = %q, want it to name %q", args, stderr, "user scope")
	}
	if calls := f.run.argv(); len(calls) != 0 {
		t.Errorf("%q runner calls = %q, want none", args, calls)
	}
	root := filepath.Dir(f.unitDir)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".service") {
			t.Errorf("%q wrote unit file %s, want none", args, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk fixture root: %v", err)
	}
}

func TestInstallCommandRefusesSystemScope(t *testing.T) {
	args := []string{"service", "--scope", "system"}
	f := newCommandFixture(t, fakeProbe("systemd"), "ready")

	code := installCommand(f.deps).Run(t.Context(), f.env, args)

	assertSystemScopeRefused(t, f, args, code)
}

func TestServiceCommandRefusesSystemScope(t *testing.T) {
	for _, sub := range []string{"start", "stop", "restart", "status", "uninstall"} {
		t.Run(sub, func(t *testing.T) {
			args := []string{sub, "--scope", "system"}
			f := newCommandFixture(t, fakeProbe("systemd"))

			code := serviceCommand(f.deps).Run(t.Context(), f.env, args)

			assertSystemScopeRefused(t, f, args, code)
		})
	}
}
