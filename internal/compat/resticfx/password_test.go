package resticfx

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var passwordHexRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

func TestNewPasswordFile(t *testing.T) {
	dir := t.TempDir()

	first := newCheckedPasswordFile(t, dir)
	second := newCheckedPasswordFile(t, dir)

	if first == second {
		t.Errorf("two NewPasswordFile calls produced identical contents %q", first)
	}
}

func newCheckedPasswordFile(t *testing.T, dir string) string {
	t.Helper()
	path, err := NewPasswordFile(dir)
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	if path == "" {
		t.Fatal("NewPasswordFile returned an empty path")
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		t.Fatalf("password file %q is not inside %q", path, dir)
	}
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat %s: %v", path, err)
	}
	if !fi.Mode().IsRegular() {
		t.Fatalf("password file %s is not a regular file: %v", path, fi.Mode())
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("password file mode = %v, want 0600", fi.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !passwordHexRe.Match(data) {
		t.Errorf("password file content is not 64 lowercase hex chars (len %d)", len(data))
	}
	return string(data)
}
