package web

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

func TestStoreCredentialWritesA0600File(t *testing.T) {
	stateDir := t.TempDir()
	const secret = "correct horse battery staple"

	got, err := storeCredential(stateDir, "main", secret)
	if err != nil {
		t.Fatalf("storeCredential(stateDir, main, secret) error = %v, want nil", err)
	}
	want := filepath.Join(stateDir, "credentials", "main.pass")
	if got != want {
		t.Errorf("storeCredential(stateDir, main, secret) = %q, want %q", got, want)
	}

	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("os.ReadFile(credential) error = %v, want nil", err)
	}
	if string(data) != secret {
		t.Errorf("credential file = %q, want %q", data, secret)
	}

	fi, err := os.Stat(got)
	if err != nil {
		t.Fatalf("os.Stat(credential) error = %v, want nil", err)
	}
	if got, want := fi.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("credential file mode = %v, want %v", got, want)
	}
	di, err := os.Stat(filepath.Dir(want))
	if err != nil {
		t.Fatalf("os.Stat(credentials dir) error = %v, want nil", err)
	}
	if got, want := di.Mode().Perm(), os.FileMode(0o700); got != want {
		t.Errorf("credentials dir mode = %v, want %v", got, want)
	}
}

func TestStoreCredentialEmptySecretKeepsTheFile(t *testing.T) {
	stateDir := t.TempDir()
	const secret = "kept as typed"

	path, err := storeCredential(stateDir, "main", secret)
	if err != nil {
		t.Fatalf("storeCredential(stateDir, main, secret) error = %v, want nil", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(credential) error = %v, want nil", err)
	}

	got, err := storeCredential(stateDir, "main", "")
	if err != nil {
		t.Fatalf("storeCredential(stateDir, main, \"\") error = %v, want nil", err)
	}
	if got != path {
		t.Errorf("storeCredential(stateDir, main, \"\") = %q, want %q", got, path)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(credential) error = %v, want nil", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("credential mtime = %v, want %v", after.ModTime(), before.ModTime())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(credential) error = %v, want nil", err)
	}
	if string(data) != secret {
		t.Errorf("credential file = %q, want %q", data, secret)
	}
}

func TestStoreCredentialEmptySecretWithoutFile(t *testing.T) {
	stateDir := t.TempDir()

	got, err := storeCredential(stateDir, "main", "")
	if err != nil {
		t.Fatalf("storeCredential(stateDir, main, \"\") error = %v, want nil", err)
	}
	if got != "" {
		t.Errorf("storeCredential(stateDir, main, \"\") = %q, want %q", got, "")
	}
}

func TestStoreCredentialRejectsTraversingIDs(t *testing.T) {
	stateDir := t.TempDir()
	for _, id := range []string{"../x", "a/b", "", "main$"} {
		got, err := storeCredential(stateDir, id, "secret")
		if err == nil {
			t.Errorf("storeCredential(stateDir, %q, secret) = %q, nil, want an error", id, got)
		}
		if got != "" {
			t.Errorf("storeCredential(stateDir, %q, secret) = %q, want %q", id, got, "")
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "credentials")); !os.IsNotExist(err) {
		t.Errorf("os.Stat(credentials dir) error = %v, want a not-exist error", err)
	}
}

func TestMarshalOfAStoredCredentialHoldsOnlyThePath(t *testing.T) {
	stateDir := t.TempDir()
	const secret = "never-in-yaml"

	path, err := storeCredential(stateDir, "main", secret)
	if err != nil {
		t.Fatalf("storeCredential(stateDir, main, secret) error = %v, want nil", err)
	}
	cfg := &config.Config{Repositories: []config.Repository{{
		ID:           "main",
		Repository:   "local:/srv/restic",
		PasswordFile: path,
	}}}

	data, err := config.Marshal(cfg)
	if err != nil {
		t.Fatalf("config.Marshal(cfg) error = %v, want nil", err)
	}
	yaml := string(data)
	if strings.Contains(yaml, secret) {
		t.Errorf("config.Marshal(cfg) contains the secret %q, want it absent", secret)
	}
	want := "password_file: " + filepath.Join(stateDir, "credentials", "main.pass")
	if !strings.Contains(yaml, want) {
		t.Errorf("config.Marshal(cfg) = %q, want it to contain %q", yaml, want)
	}
	for _, line := range strings.Split(yaml, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "password:") {
			t.Errorf("config.Marshal(cfg) line %q, want no password key", line)
		}
	}
}

func TestCredentialMode(t *testing.T) {
	tests := []struct {
		name string
		v    url.Values
		i    int
		want string
	}{
		{"typed", url.Values{"repositories[0].password_mode": {"typed"}}, 0, "typed"},
		{"file", url.Values{"repositories[0].password_mode": {"file"}}, 0, "file"},
		{"second row", url.Values{"repositories[1].password_mode": {"typed"}}, 1, "typed"},
		{"missing", url.Values{}, 0, "file"},
		{"unknown", url.Values{"repositories[0].password_mode": {"keychain"}}, 0, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credentialMode(tt.v, tt.i); got != tt.want {
				t.Errorf("credentialMode(v, %d) = %q, want %q", tt.i, got, tt.want)
			}
		})
	}
}
