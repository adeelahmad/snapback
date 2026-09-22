package service

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallCreatesMissingUnitDir(t *testing.T) {
	xdg := filepath.Join(t.TempDir(), "xdg")
	dir := filepath.Join(xdg, "systemd", "user")
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"ready"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 5 * time.Second}

	if err := s.Install(t.Context(), userOpts); err != nil {
		t.Fatalf("Install(userOpts) with missing unit dir = %v, want nil", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat unit dir: %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o755); !info.IsDir() || got != want {
		t.Errorf("unit dir mode = %o (dir %t), want %o dir", got, info.IsDir(), want)
	}
	for _, p := range []string{xdg, filepath.Join(xdg, "systemd")} {
		pi, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat parent %s: %v", p, err)
		}
		if got := pi.Mode().Perm(); !pi.IsDir() || (got != 0o700 && got != 0o755) {
			t.Errorf("parent %s mode = %o (dir %t), want 700 or 755 dir", p, got, pi.IsDir())
		}
	}
	unit := filepath.Join(dir, "snapback.service")
	ui, err := os.Stat(unit)
	if err != nil {
		t.Fatalf("stat installed unit: %v", err)
	}
	if got, want := ui.Mode().Perm(), fs.FileMode(0o644); got != want {
		t.Errorf("unit mode = %o, want %o", got, want)
	}

	if err := s.Uninstall(t.Context()); err != nil {
		t.Fatalf("Uninstall() = %v, want nil", err)
	}
	if _, err := os.Stat(unit); !os.IsNotExist(err) {
		t.Errorf("stat unit after Uninstall() = %v, want not exist", err)
	}
	if di, err := os.Stat(dir); err != nil || !di.IsDir() {
		t.Errorf("stat unit dir after Uninstall() = %v, want dir kept", err)
	}
}

func TestInstallKeepsExistingUnitDirMode(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "user")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("create unit dir: %v", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatalf("chmod unit dir: %v", err)
	}
	run := &fakeRunner{}
	ready := &scriptedReady{states: []string{"ready"}}
	s := &Systemd{UnitDir: dir, Run: run.run, Ready: ready.ready, ReadyTimeout: 5 * time.Second}

	if err := s.Install(t.Context(), userOpts); err != nil {
		t.Fatalf("Install(userOpts) with existing unit dir = %v, want nil", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat unit dir: %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o700); got != want {
		t.Errorf("existing unit dir mode after Install() = %o, want %o", got, want)
	}
}
