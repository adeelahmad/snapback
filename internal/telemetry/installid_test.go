package telemetry_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// isLowerHex reports whether s is exactly n lowercase hex digits.
func isLowerHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

// TestInstallIDCreatesFile pins decision D6: the id is 32 lowercase hex chars
// (128 bits of crypto/rand) written to <dir>/install_id with mode 0600, in a
// directory the call creates with 0700 when it is missing.
func TestInstallIDCreatesFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "telemetry")

	got, err := telemetry.InstallID(dir)
	if err != nil {
		t.Fatalf("InstallID(%q) returned error: %v", dir, err)
	}
	if !isLowerHex(got, 32) {
		t.Errorf("InstallID(%q) = %q, want 32 lowercase hex chars", dir, got)
	}

	path := filepath.Join(dir, "install_id")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if strings.TrimSpace(string(raw)) != got {
		t.Errorf("%s holds %q, want the returned id %q", path, string(raw), got)
	}

	if runtime.GOOS == "windows" {
		t.Skip("file and directory modes are not meaningful on windows")
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if perm := fi.Mode().Perm(); perm != fs.FileMode(0o600) {
		t.Errorf("mode of %s = %#o, want 0600", path, perm)
	}
	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat %s: %v", dir, err)
	}
	if perm := di.Mode().Perm(); perm != fs.FileMode(0o700) {
		t.Errorf("mode of %s = %#o, want 0700", dir, perm)
	}
}

// TestInstallIDIsStable pins that the second call reads the stored id back
// rather than minting a new one.
func TestInstallIDIsStable(t *testing.T) {
	dir := t.TempDir()

	first, err := telemetry.InstallID(dir)
	if err != nil {
		t.Fatalf("first InstallID(%q) returned error: %v", dir, err)
	}
	second, err := telemetry.InstallID(dir)
	if err != nil {
		t.Fatalf("second InstallID(%q) returned error: %v", dir, err)
	}
	if !isLowerHex(first, 32) {
		t.Fatalf("first InstallID(%q) = %q, want 32 lowercase hex chars", dir, first)
	}
	if first != second {
		t.Errorf("InstallID(%q) = %q then %q, want the same value twice", dir, first, second)
	}
}

// TestInstallIDDiffersPerDirectory pins that the id is random, not derived from
// the hostname, user or repository: two installs on this one host differ.
func TestInstallIDDiffersPerDirectory(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()

	idA, err := telemetry.InstallID(dirA)
	if err != nil {
		t.Fatalf("InstallID(%q) returned error: %v", dirA, err)
	}
	idB, err := telemetry.InstallID(dirB)
	if err != nil {
		t.Fatalf("InstallID(%q) returned error: %v", dirB, err)
	}
	if idA == idB {
		t.Errorf("two install dirs on one host both gave %q, want different random ids", idA)
	}
}

// TestInstallIDRejectsCorruptFile pins that a file that is not 32 hex chars is
// an error naming the path, never a silently regenerated id.
func TestInstallIDRejectsCorruptFile(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"too short", "abc123"},
		{"too long", strings.Repeat("a", 33)},
		{"not hex", strings.Repeat("z", 32)},
		{"uppercase", strings.Repeat("A", 32)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "install_id")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("writing %s: %v", path, err)
			}

			got, err := telemetry.InstallID(dir)
			if err == nil {
				t.Fatalf("InstallID(%q) = %q, nil; want an error for a corrupt file", dir, got)
			}
			if !strings.Contains(err.Error(), path) {
				t.Errorf("InstallID(%q) error = %q, want it to name %s", dir, err, path)
			}
		})
	}
}

// TestForgetInstallIDRemovesFile pins that disable deletes the id and that a
// later call mints a fresh one.
func TestForgetInstallIDRemovesFile(t *testing.T) {
	dir := t.TempDir()

	first, err := telemetry.InstallID(dir)
	if err != nil {
		t.Fatalf("InstallID(%q) returned error: %v", dir, err)
	}
	if err := telemetry.ForgetInstallID(dir); err != nil {
		t.Fatalf("ForgetInstallID(%q) returned error: %v", dir, err)
	}
	path := filepath.Join(dir, "install_id")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("stat %s after ForgetInstallID = %v, want a not-exist error", path, err)
	}

	second, err := telemetry.InstallID(dir)
	if err != nil {
		t.Fatalf("InstallID(%q) after forget returned error: %v", dir, err)
	}
	if second == first {
		t.Errorf("InstallID(%q) after forget = %q, want a freshly minted id", dir, first)
	}
}

// TestForgetInstallIDIsIdempotent pins that a missing file, and a missing
// directory, are not errors.
func TestForgetInstallIDIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := telemetry.ForgetInstallID(dir); err != nil {
		t.Errorf("ForgetInstallID(%q) on a missing file returned error: %v", dir, err)
	}
	missing := filepath.Join(dir, "never-created")
	if err := telemetry.ForgetInstallID(missing); err != nil {
		t.Errorf("ForgetInstallID(%q) on a missing dir returned error: %v", missing, err)
	}
}
