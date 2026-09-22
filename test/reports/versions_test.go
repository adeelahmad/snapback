package reports_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	pinnedRestic  = "0.19.0"
	pinnedHeading = "## Pinned versions"
)

var (
	goFuseRequire = regexp.MustCompile(`(?m)^\s*(?:require\s+)?github\.com/hanwen/go-fuse/v2\s+(\S+)`)
	toolchainLine = regexp.MustCompile(`(?m)^toolchain\s+(\S+)\s*$`)
	resticVersion = regexp.MustCompile(`\d+\.\d+\.\d+`)
	rcloneVersion = regexp.MustCompile(`v?\d+\.\d+\.\d+`)
)

// readGoMod returns the repository's go.mod text.
func readGoMod(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	return string(data)
}

// pinnedRows returns name -> value for the Pinned versions table, failing if it has no rows.
func pinnedRows(t *testing.T) map[string]string {
	t.Helper()
	rows := tableRows(section(t, readReport(t), pinnedHeading))
	if len(rows) == 0 {
		t.Fatalf("%s: %q has no table rows", reportPath, pinnedHeading)
	}
	got := map[string]string{}
	for _, cells := range rows {
		if len(cells) < 2 {
			t.Errorf("Pinned versions row %q has %d cells, want 2", cells, len(cells))
			continue
		}
		if _, dup := got[cells[0]]; dup {
			t.Errorf("Pinned versions row %q appears more than once", cells[0])
		}
		got[cells[0]] = cells[1]
	}
	return got
}

// presentWithPrefix returns the present evidence files whose name starts with prefix.
func presentWithPrefix(t *testing.T, prefix string) []string {
	t.Helper()
	var names []string
	for _, name := range presentEvidence(t) {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	return names
}

// stringField returns obj[key] as a string, failing with the file and key named.
func stringField(t *testing.T, file string, obj map[string]any, key string) string {
	t.Helper()
	v, ok := obj[key]
	if !ok {
		t.Fatalf("evidence %s: key %q missing", file, key)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("evidence %s: %q = %v (%T), want string", file, key, v, v)
	}
	return s
}

func TestGoFuseVersionMatchesGoMod(t *testing.T) {
	m := goFuseRequire.FindStringSubmatch(readGoMod(t))
	if m == nil {
		t.Fatalf("go.mod: no github.com/hanwen/go-fuse/v2 require line")
	}
	want := m[1]

	rows := pinnedRows(t)
	got, ok := rows["go-fuse"]
	if !ok {
		t.Fatalf("Pinned versions: no go-fuse row")
	}
	if got != want {
		t.Errorf("Pinned versions go-fuse = %q, want %q (go.mod)", got, want)
	}

	for _, name := range presentWithPrefix(t, "catalog-") {
		if v := stringField(t, name, loadEvidence(t, name), "go_fuse_version"); v != want {
			t.Errorf("evidence %s go_fuse_version = %q, want %q (go.mod)", name, v, want)
		}
	}
}

func TestPinnedVersionsTable(t *testing.T) {
	rows := pinnedRows(t)
	for _, name := range []string{"restic", "rclone", "Go toolchain"} {
		if _, ok := rows[name]; !ok {
			t.Errorf("Pinned versions: no %q row", name)
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	if got := rows["restic"]; got != pinnedRestic {
		t.Errorf("Pinned versions restic = %q, want %q", got, pinnedRestic)
	}
	for _, name := range presentWithPrefix(t, "pathtemplate-") {
		raw := stringField(t, name, loadEvidence(t, name), "restic_version")
		if v := resticVersion.FindString(raw); v != rows["restic"] {
			t.Errorf("evidence %s restic_version %q has version %q, want %q (report)", name, raw, v, rows["restic"])
		}
	}

	latency := loadEvidence(t, "latency.json")
	rawRestic := stringField(t, "latency.json", latency, "restic_version")
	if v := resticVersion.FindString(rawRestic); v != rows["restic"] {
		t.Errorf("latency.json restic_version %q has version %q, want %q (report)", rawRestic, v, rows["restic"])
	}
	rawRclone := stringField(t, "latency.json", latency, "rclone_version")
	wantRclone := rcloneVersion.FindString(rawRclone)
	if wantRclone == "" {
		t.Errorf("latency.json rclone_version %q has no version number", rawRclone)
	}
	if got := rows["rclone"]; got != wantRclone {
		t.Errorf("Pinned versions rclone = %q, want %q (latency.json %q)", got, wantRclone, rawRclone)
	}

	m := toolchainLine.FindStringSubmatch(readGoMod(t))
	if m == nil {
		t.Fatalf("go.mod: no toolchain directive")
	}
	if got := rows["Go toolchain"]; got != m[1] {
		t.Errorf("Pinned versions Go toolchain = %q, want %q (go.mod)", got, m[1])
	}
}
