package logging_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging"
)

const (
	testDirMode  fs.FileMode = 0o750
	testFileMode fs.FileMode = 0o640
)

func TestFileOpenCreatesMissingParentsAndAppends(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "a", "b", "snapback.log")

	writeLine := func(line string) {
		t.Helper()
		w, closer, err := logging.Open(path, testDirMode, testFileMode, io.Discard)
		if err != nil {
			t.Fatalf("Open(%q) = error %v, want nil", path, err)
		}
		if _, err := w.Write([]byte(line)); err != nil {
			t.Fatalf("Write(%q) = error %v, want nil", line, err)
		}
		if err := closer.Close(); err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
	}

	writeLine("one\n")

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("stat parent of %q = error %v, want the directory to exist", path, err)
	}

	writeLine("two\n")

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) = error %v, want nil", path, err)
	}
	if want := "one\ntwo\n"; string(got) != want {
		t.Errorf("log file contents = %q, want %q (appended, not truncated)", got, want)
	}
}

func TestFileOpenUsesTheConfiguredModes(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "logs")
	path := filepath.Join(dir, "snapback.log")

	_, closer, err := logging.Open(path, testDirMode, testFileMode, io.Discard)
	if err != nil {
		t.Fatalf("Open(%q) = error %v, want nil", path, err)
	}
	if err := closer.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q = error %v, want the log file to exist", path, err)
	}
	if got := fi.Mode().Perm(); got != testFileMode {
		t.Errorf("log file mode = %#o, want %#o", got, testFileMode)
	}

	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat %q = error %v, want the log directory to exist", dir, err)
	}
	if got := di.Mode().Perm(); got != testDirMode {
		t.Errorf("log directory mode = %#o, want %#o", got, testDirMode)
	}
}

func TestFileOpenReportsAnUnwritableDirectory(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root: an unwritable directory is still writable")
	}

	locked := filepath.Join(t.TempDir(), "locked")
	if err := os.Mkdir(locked, 0o500); err != nil {
		t.Fatalf("Mkdir(%q) = error %v, want nil", locked, err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })

	path := filepath.Join(locked, "snapback.log")

	_, _, err := logging.Open(path, testDirMode, testFileMode, io.Discard)
	if err == nil {
		t.Fatalf("Open(%q) = nil error, want an error naming the path", path)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("Open(%q) error = %q, want it to contain %q", path, err.Error(), path)
	}
}

func TestFileOpenWithAnEmptyPathReturnsTheFallback(t *testing.T) {
	t.Parallel()

	var fallback bytes.Buffer

	w, closer, err := logging.Open("", testDirMode, testFileMode, &fallback)
	if err != nil {
		t.Fatalf(`Open("") = error %v, want nil`, err)
	}
	if w != io.Writer(&fallback) {
		t.Errorf(`Open("") writer = %#v, want the fallback writer itself`, w)
	}
	if _, err := w.Write([]byte("fallback\n")); err != nil {
		t.Fatalf("Write = error %v, want nil", err)
	}
	if got, want := fallback.String(), "fallback\n"; got != want {
		t.Errorf("fallback buffer = %q, want %q", got, want)
	}
	if err := closer.Close(); err != nil {
		t.Errorf(`Open("") closer.Close() = %v, want nil (a no-op closer)`, err)
	}
}
