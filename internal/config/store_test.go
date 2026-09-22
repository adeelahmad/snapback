package config

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

var revPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// storePath returns <tmp>/cfg/config.yaml under a fresh t.TempDir(); the
// cfg directory does not exist yet.
func storePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cfg", "config.yaml")
}

// sha256Hex returns the lowercase hex SHA-256 of b as a Revision.
func sha256Hex(b []byte) Revision {
	sum := sha256.Sum256(b)
	return Revision(hex.EncodeToString(sum[:]))
}

// mustRead returns the bytes of path or fails the test.
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v", path, err)
	}
	return b
}

// mustSave saves c at path expecting rev and fails the test on error.
func mustSave(t *testing.T, path string, c *Config, expected Revision) Revision {
	t.Helper()
	rev, err := Save(path, c, expected)
	if err != nil {
		t.Fatalf("Save(%q, c, %q) = %v, want nil error", path, expected, err)
	}
	return rev
}

func TestSaveNewFileAndLoad(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)

	rev, err := Save(path, c, "")
	if err != nil {
		t.Fatalf("Save(%q, c, \"\") = %v, want nil error", path, err)
	}
	if !revPattern.MatchString(string(rev)) {
		t.Errorf("Save revision = %q, want 64 lowercase hex characters", rev)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v, want the saved file", path, err)
	}
	if want := sha256Hex(data); rev != want {
		t.Errorf("Save revision = %q, want sha256 of file %q", rev, want)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) = %v", path, err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}
	di, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat(parent) = %v", err)
	}
	if got := di.Mode().Perm(); got != 0o700 {
		t.Errorf("parent mode = %o, want 700", got)
	}

	got, gotRev, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) = %v, want nil error", path, err)
	}
	if gotRev != rev {
		t.Errorf("Load revision = %q, want %q", gotRev, rev)
	}
	// Parse normalizes nil slices and maps to empty ones, so compare against
	// the saved config as it reads back from its own YAML.
	b, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v", err)
	}
	want, err := Parse(b)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load(%q) = %+v, want %+v", path, got, want)
	}
}

func TestSaveWithCurrentRevisionSucceeds(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	rev1 := mustSave(t, path, c, "")

	c.LinkName = ".history"
	rev2, err := Save(path, c, rev1)
	if err != nil {
		t.Fatalf("Save(path, c, rev1) = %v, want nil error", err)
	}
	if rev2 == rev1 {
		t.Errorf("Save(path, c, rev1) revision = %q, want different from rev1", rev2)
	}
	got, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) = %v, want nil error", path, err)
	}
	if got.LinkName != ".history" {
		t.Errorf("Load(%q).LinkName = %q, want %q", path, got.LinkName, ".history")
	}
}

func TestSaveStaleRevisionConflicts(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	rev1 := mustSave(t, path, c, "")
	c.LinkName = ".two"
	mustSave(t, path, c, rev1)
	before := mustRead(t, path)

	c.LinkName = ".three"
	_, err := Save(path, c, rev1)
	if !errors.Is(err, ErrRevisionConflict) {
		t.Errorf("Save(path, c3, rev1) = %v, want ErrRevisionConflict", err)
	}
	if got := errcode.Of(err); got != errcode.StaleState {
		t.Errorf("errcode.Of(Save(path, c3, rev1)) = %q, want %q", got, errcode.StaleState)
	}
	if after := mustRead(t, path); string(after) != string(before) {
		t.Errorf("file after conflict = %q, want unchanged %q", after, before)
	}
}

func TestSaveExpectingAbsentButFileExists(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	mustSave(t, path, c, "")
	before := mustRead(t, path)

	c.LinkName = ".other"
	_, err := Save(path, c, "")
	if !errors.Is(err, ErrRevisionConflict) {
		t.Errorf("Save(path, c, \"\") on existing file = %v, want ErrRevisionConflict", err)
	}
	if after := mustRead(t, path); string(after) != string(before) {
		t.Errorf("file after conflict = %q, want unchanged %q", after, before)
	}
}

func TestSaveRenameFailureLeavesOldFile(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	rev1 := mustSave(t, path, c, "")
	before := mustRead(t, path)

	orig := rename
	t.Cleanup(func() { rename = orig })
	rename = func(string, string) error { return errors.New("injected") }

	c.LinkName = ".crash"
	_, err := Save(path, c, rev1)
	if err == nil || !strings.Contains(err.Error(), "injected") {
		t.Errorf("Save with failing rename = %v, want error containing %q", err, "injected")
	}
	if after := mustRead(t, path); string(after) != string(before) {
		t.Errorf("file after failed rename = %q, want unchanged %q", after, before)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir = %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if want := []string{"config.yaml", "config.yaml.lock"}; !slices.Equal(names, want) {
		t.Errorf("ReadDir after failed rename = %q, want %q (no temp file)", names, want)
	}
}

func TestSaveInvalidConfigWritesNothing(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	rev1 := mustSave(t, path, c, "")
	before := mustRead(t, path)

	invalid := validConfig(t)
	invalid.LinkName = ""
	_, err := Save(path, invalid, rev1)
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(Save(path, invalid, rev1)) = %q, want %q (err %v)", got, errcode.InvalidConfig, err)
	}
	if after := mustRead(t, path); string(after) != string(before) {
		t.Errorf("file after invalid save = %q, want unchanged %q", after, before)
	}

	absent := storePath(t)
	_, err = Save(absent, invalid, "")
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(Save(absent, invalid, \"\")) = %q, want %q", got, errcode.InvalidConfig)
	}
	if _, err := os.Stat(absent); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(%q) after invalid save = %v, want not exist", absent, err)
	}
}

func TestConcurrentSavesOneWins(t *testing.T) {
	const n = 8
	path := storePath(t)
	rev1 := mustSave(t, path, validConfig(t), "")

	cfgs := make([]*Config, n)
	for i := range cfgs {
		cfgs[i] = validConfig(t)
		cfgs[i].LinkName = fmt.Sprintf(".snap%d", i)
	}
	errs := make([]error, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			_, errs[i] = Save(path, cfgs[i], rev1)
		})
	}
	close(start)
	wg.Wait()

	winner, conflicts := -1, 0
	for i, err := range errs {
		switch {
		case err == nil:
			if winner >= 0 {
				t.Errorf("Save #%d and #%d both succeeded, want exactly one", winner, i)
			}
			winner = i
		case errors.Is(err, ErrRevisionConflict):
			conflicts++
		default:
			t.Errorf("Save #%d = %v, want nil or ErrRevisionConflict", i, err)
		}
	}
	if winner < 0 || conflicts != n-1 {
		t.Fatalf("concurrent saves: winner %d, conflicts %d, want one winner and %d conflicts", winner, conflicts, n-1)
	}
	got, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) = %v, want nil error", path, err)
	}
	if want := cfgs[winner].LinkName; got.LinkName != want {
		t.Errorf("Load(%q).LinkName = %q, want winner's %q", path, got.LinkName, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "none.yaml")
	_, rev, err := Load(path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load(%q) error = %v, want fs.ErrNotExist", path, err)
	}
	if rev != "" {
		t.Errorf("Load(%q) revision = %q, want \"\"", path, rev)
	}
}

func TestLoadInvalidFileReturnsRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("version: 1\nbogus: x\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v", path, err)
	}

	c, rev, err := Load(path)
	if c != nil {
		t.Errorf("Load(invalid) config = %+v, want nil", c)
	}
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(Load(invalid)) = %q, want %q (err %v)", got, errcode.InvalidConfig, err)
	}
	if want := sha256Hex(data); rev != want {
		t.Errorf("Load(invalid) revision = %q, want %q", rev, want)
	}
}

// pathFieldKey reports whether a YAML key names a filesystem path.
func pathFieldKey(key string) bool {
	if key == "local_path" {
		return true
	}
	for _, s := range []string{"_dir", "_mount", "_file", "_binary"} {
		if strings.HasSuffix(key, s) {
			return true
		}
	}
	return false
}

// collectPathFields records in out every non-empty string field in v whose
// YAML key is a path key, recursing into structs and slices.
func collectPathFields(v reflect.Value, out map[string]string, prefix string) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			collectPathFields(v.Elem(), out, prefix)
		}
	case reflect.Struct:
		for i := range v.NumField() {
			key, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("yaml"), ",")
			f := v.Field(i)
			if f.Kind() == reflect.String {
				if pathFieldKey(key) && f.String() != "" {
					out[prefix+key] = f.String()
				}
				continue
			}
			collectPathFields(f, out, prefix+key+".")
		}
	case reflect.Slice:
		for i := range v.Len() {
			collectPathFields(v.Index(i), out, fmt.Sprintf("%s%d.", prefix, i))
		}
	}
}

func TestSavedFileHasNoRelativeOrTildePaths(t *testing.T) {
	path := storePath(t)
	c := validConfig(t)
	mustSave(t, path, c, "")

	got, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q) = %v, want nil error", path, err)
	}
	fields := map[string]string{}
	collectPathFields(reflect.ValueOf(got), fields, "")
	if len(fields) < 5 {
		t.Errorf("path fields found = %d (%v), want at least 5", len(fields), fields)
	}
	for k, p := range fields {
		if !filepath.IsAbs(p) || strings.HasPrefix(p, "~") {
			t.Errorf("%s = %q, want an absolute path without ~", k, p)
		}
	}
}
