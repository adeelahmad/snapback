package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const (
	linuxAmd64Asset = "snapback_linux_amd64.tar.gz"
	wrongDigest     = "0000000000000000000000000000000000000000000000000000000000000000"
)

// releaseServer serves a fake release directory and records every path hit.
type releaseServer struct {
	*httptest.Server
	mu   sync.Mutex
	hits map[string]int
}

func (s *releaseServer) hitCount(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits[path]
}

// newReleaseServer serves files by name; a name absent from files returns 404.
func newReleaseServer(t *testing.T, files map[string][]byte) *releaseServer {
	t.Helper()
	rs := &releaseServer{hits: map[string]int{}}
	rs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		rs.mu.Lock()
		rs.hits[name]++
		rs.mu.Unlock()
		body, ok := files[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(rs.Close)
	return rs
}

// fakeArchive builds a tar.gz holding a `snapback` script at its root that echoes ok.
func fakeArchive(t *testing.T) []byte {
	t.Helper()
	script := []byte("#!/bin/sh\necho ok\n")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: "snapback", Mode: 0o755, Size: int64(len(script)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(script); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func installEnv(baseURL, installDir string) map[string]string {
	return map[string]string{
		"SNAPBACK_OS":          "Linux",
		"SNAPBACK_ARCH":        "x86_64",
		"SNAPBACK_BASE_URL":    baseURL,
		"SNAPBACK_INSTALL_DIR": installDir,
	}
}

func assertDirEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read install dir: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("install dir %s is not empty: %v", dir, names)
	}
}

func TestInstallGoodChecksum(t *testing.T) {
	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		linuxAmd64Asset: archive,
		"checksums.txt": []byte(sha256Hex(archive) + "  " + linuxAmd64Asset + "\n"),
	})
	installDir := t.TempDir()

	_, stderr, code := runInstaller(t, installEnv(srv.URL, installDir))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	bin := filepath.Join(installDir, "snapback")
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("installed binary mode = %v, want executable bit set", info.Mode().Perm())
	}
	out, err := exec.Command(bin).Output()
	if err != nil {
		t.Fatalf("run installed binary: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "ok" {
		t.Errorf("installed binary printed %q, want %q", got, "ok")
	}
}

func TestInstallBadChecksumRefuses(t *testing.T) {
	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		linuxAmd64Asset: archive,
		"checksums.txt": []byte(wrongDigest + "  " + linuxAmd64Asset + "\n"),
	})
	installDir := t.TempDir()

	_, stderr, code := runInstaller(t, installEnv(srv.URL, installDir))

	if code == 0 {
		t.Errorf("exit code = 0, want non-zero on checksum mismatch")
	}
	if !strings.Contains(stderr, "checksum") {
		t.Errorf("stderr does not contain %q; stderr=%q", "checksum", stderr)
	}
	if srv.hitCount(linuxAmd64Asset) == 0 {
		t.Errorf("server never received a request for %s; the script did not attempt the download", linuxAmd64Asset)
	}
	assertDirEmpty(t, installDir)
}

func TestInstallAssetMissingFromChecksums(t *testing.T) {
	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		linuxAmd64Asset: archive,
		"checksums.txt": []byte(sha256Hex(archive) + "  snapback_darwin_arm64.tar.gz\n" +
			wrongDigest + "  snapback_linux_arm64.tar.gz\n"),
	})
	installDir := t.TempDir()

	_, stderr, code := runInstaller(t, installEnv(srv.URL, installDir))

	if code == 0 {
		t.Errorf("exit code = 0, want non-zero when the asset is absent from checksums.txt")
	}
	if !strings.Contains(stderr, "checksum") {
		t.Errorf("stderr does not contain %q; stderr=%q", "checksum", stderr)
	}
	if srv.hitCount("checksums.txt") == 0 {
		t.Errorf("server never received a request for checksums.txt; the script did not attempt verification")
	}
	assertDirEmpty(t, installDir)
}

func TestInstallDownloadFailure(t *testing.T) {
	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		"checksums.txt": []byte(sha256Hex(archive) + "  " + linuxAmd64Asset + "\n"),
	})
	installDir := t.TempDir()

	_, stderr, code := runInstaller(t, installEnv(srv.URL, installDir))

	if code == 0 {
		t.Errorf("exit code = 0, want non-zero when the asset download 404s; stderr=%q", stderr)
	}
	if srv.hitCount(linuxAmd64Asset) == 0 {
		t.Errorf("server never received a request for %s; the script did not attempt the download", linuxAmd64Asset)
	}
	assertDirEmpty(t, installDir)
}
