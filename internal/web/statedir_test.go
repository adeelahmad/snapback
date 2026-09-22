package web

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "state", "snapback")

	srv, err := New(Options{Listen: "127.0.0.1:0", StateDir: stateDir, Token: "tok", Stdout: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("New(StateDir: missing %q) error = %v, want nil", stateDir, err)
	}
	t.Cleanup(func() { _ = srv.Close() })

	fi, err := os.Stat(stateDir)
	if err != nil {
		t.Fatalf("os.Stat(state dir) error = %v, want the dir to exist", err)
	}
	if !fi.IsDir() {
		t.Fatalf("state dir %q is not a directory", stateDir)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o700); got != want {
		t.Errorf("state dir mode = %v, want %v", got, want)
	}

	urlFile := filepath.Join(stateDir, "web.url")
	fi, err = os.Stat(urlFile)
	if err != nil {
		t.Fatalf("os.Stat(web.url) error = %v, want the file to exist", err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("web.url mode = %v, want %v", got, want)
	}
	data, err := os.ReadFile(urlFile)
	if err != nil {
		t.Fatalf("os.ReadFile(web.url) error = %v", err)
	}
	if got, want := string(data), srv.URL(); !strings.HasPrefix(got, want) {
		t.Errorf("web.url = %q, want prefix %q", got, want)
	}
}
