package doctor

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func readBundle(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open(%q) = _, %v, want nil error", path, err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip.NewReader(%q) = _, %v, want nil error", path, err)
	}
	defer gz.Close()
	members := make(map[string]string)
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("tar.Next() = _, %v, want nil error", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("io.ReadAll(%q) = _, %v, want nil error", hdr.Name, err)
		}
		members[hdr.Name] = string(body)
	}
	return members
}

func memberNames(members map[string]string) []string {
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestWriteBundleContents(t *testing.T) {
	dir := t.TempDir()
	in := BundleInput{
		DoctorJSON:     []byte(`{"checks":[{"name":"repo","status":"ok"}]}`),
		Version:        "1.4.1",
		Commit:         "abc441b",
		GOOS:           "linux",
		GOARCH:         "amd64",
		DaemonLog:      []byte("started\nstopped\n"),
		ConfigRedacted: []byte("repo: <redacted>\n"),
	}
	now := time.Date(2026, 9, 23, 7, 8, 9, 0, time.UTC)

	path, err := WriteBundle(dir, in, now)
	if err != nil {
		t.Fatalf("WriteBundle(dir, in, now) = _, %v, want nil error", err)
	}

	wantPath := filepath.Join(dir, "snapback-bundle-20260923-070809.tar.gz")
	if path != wantPath {
		t.Errorf("WriteBundle(dir, in, now) = %q, want %q", path, wantPath)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) = _, %v, want nil error", path, err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("os.Stat(path).Mode().Perm() = %v, want %v", got, want)
	}

	members := readBundle(t, path)
	gotNames := memberNames(members)
	wantNames := []string{"README.txt", "config.yaml", "daemon.log", "doctor.json", "versions.txt"}
	if strings.Join(gotNames, ",") != strings.Join(wantNames, ",") {
		t.Errorf("bundle members = %v, want %v", gotNames, wantNames)
	}

	if got, want := members["doctor.json"], string(in.DoctorJSON); got != want {
		t.Errorf("bundle doctor.json = %q, want %q", got, want)
	}
	if got, want := members["daemon.log"], string(in.DaemonLog); got != want {
		t.Errorf("bundle daemon.log = %q, want %q", got, want)
	}
	if got, want := members["config.yaml"], string(in.ConfigRedacted); got != want {
		t.Errorf("bundle config.yaml = %q, want %q", got, want)
	}

	versions := members["versions.txt"]
	for _, want := range []string{"1.4.1", "abc441b", "linux/amd64", runtime.Version()} {
		if !strings.Contains(versions, want) {
			t.Errorf("bundle versions.txt = %q, want it to contain %q", versions, want)
		}
	}

	readme := members["README.txt"]
	for _, want := range []string{"redacted", "https://github.com/adeelahmad/snapback/issues"} {
		if !strings.Contains(readme, want) {
			t.Errorf("bundle README.txt = %q, want it to contain %q", readme, want)
		}
	}
}

func TestWriteBundleOmitsEmptyMembers(t *testing.T) {
	dir := t.TempDir()
	in := BundleInput{
		DoctorJSON: []byte(`{}`),
		Version:    "dev",
		Commit:     "none",
		GOOS:       "darwin",
		GOARCH:     "arm64",
	}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	path, err := WriteBundle(dir, in, now)
	if err != nil {
		t.Fatalf("WriteBundle(dir, in, now) = _, %v, want nil error", err)
	}
	if got, want := filepath.Base(path), "snapback-bundle-20260102-030405.tar.gz"; got != want {
		t.Errorf("filepath.Base(WriteBundle(dir, in, now)) = %q, want %q", got, want)
	}

	gotNames := memberNames(readBundle(t, path))
	wantNames := []string{"README.txt", "doctor.json", "versions.txt"}
	if strings.Join(gotNames, ",") != strings.Join(wantNames, ",") {
		t.Errorf("bundle members = %v, want %v", gotNames, wantNames)
	}
}
