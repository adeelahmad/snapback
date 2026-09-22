package resticfx

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const t6SnapshotID = "4f3c2a1b0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b"

// t6PasswordFile generates a password file inline so T6 does not depend on
// T4's NewPasswordFile.
func t6PasswordFile(t *testing.T, dir string) (path, value string) {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	value = hex.EncodeToString(buf)
	path = filepath.Join(dir, "password")
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	return path, value
}

func t6Evidence(goos string) Evidence {
	return Evidence{
		GOOS:          goos,
		GOARCH:        "arm64",
		ResticVersion: "0.19.0",
		RcloneVersion: "1.70.0",
		PathTemplate:  "ids/%I",
		SnapshotID:    t6SnapshotID,
		ObservedDir:   t6SnapshotID,
		IDsEntries:    []string{t6SnapshotID},
		Result:        "pass",
		GeneratedAt:   time.Date(2026, 9, 22, 2, 16, 36, 0, time.UTC),
	}
}

func TestEvidenceRoundTrip(t *testing.T) {
	ev := t6Evidence("darwin")
	path, err := WriteEvidence(t.TempDir(), ev)
	if err != nil {
		t.Fatalf("WriteEvidence: %v", err)
	}
	got, err := ReadEvidence(path)
	if err != nil {
		t.Fatalf("ReadEvidence: %v", err)
	}
	if !reflect.DeepEqual(got, ev) {
		t.Errorf("round trip mismatch:\n got %#v\nwant %#v", got, ev)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence file: %v", err)
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Errorf("evidence file does not end with a newline")
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("evidence file is not a JSON object: %v", err)
	}
	for _, key := range []string{"restic_version", "snapshot_id", "observed_dir", "path_template", "result"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("evidence JSON missing key %q", key)
		}
	}
}

func TestWriteEvidenceFileName(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		t.Run(goos, func(t *testing.T) {
			dir := t.TempDir()
			want := filepath.Join(dir, "pathtemplate-"+goos+".json")
			path, err := WriteEvidence(dir, t6Evidence(goos))
			if err != nil {
				t.Fatalf("WriteEvidence: %v", err)
			}
			if path != want {
				t.Errorf("path = %q, want %q", path, want)
			}
			if _, err := os.Stat(want); err != nil {
				t.Errorf("evidence file not created: %v", err)
			}
		})
	}
	t.Run("empty dir", func(t *testing.T) {
		if _, err := WriteEvidence("", t6Evidence("linux")); err == nil {
			t.Errorf("WriteEvidence with empty dir: want error, got nil")
		}
	})
}

func TestEvidenceOmitsPassword(t *testing.T) {
	pwPath, pwValue := t6PasswordFile(t, t.TempDir())
	ev := t6Evidence("linux")
	ev.Reason = "fixture run with fake runner"
	path, err := WriteEvidence(t.TempDir(), ev)
	if err != nil {
		t.Fatalf("WriteEvidence: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence file: %v", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		t.Fatalf("evidence file is empty")
	}
	if strings.Contains(string(data), pwValue) {
		t.Errorf("evidence file contains the password value")
	}
	if strings.Contains(string(data), pwPath) {
		t.Errorf("evidence file contains the password-file path %q", pwPath)
	}
}

func TestEvidenceDirFromEnv(t *testing.T) {
	dir := t.TempDir()
	got := EvidenceDir(func(k string) string {
		if k == "SNAPBACK_EVIDENCE_DIR" {
			return dir
		}
		return ""
	})
	if got != dir {
		t.Errorf("EvidenceDir = %q, want %q", got, dir)
	}
	if got := EvidenceDir(func(string) string { return "" }); got != "" {
		t.Errorf("EvidenceDir with unset env = %q, want empty", got)
	}
}
