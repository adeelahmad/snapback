package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// writePasswordFile creates a 0600 placeholder password file at
// <tmp>/password and returns its path.
func writePasswordFile(t *testing.T, tmp string) string {
	t.Helper()
	p := filepath.Join(tmp, "password")
	if err := os.WriteFile(p, []byte("placeholder\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v", p, err)
	}
	if err := os.Chmod(p, 0o600); err != nil {
		t.Fatalf("Chmod(%q) = %v", p, err)
	}
	return p
}

// exampleYAML returns testdata/example.yaml with @TMP@ replaced by a fresh
// temp dir that also holds the fixture's password file.
func exampleYAML(t *testing.T) (tmp string, data []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "example.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(testdata/example.yaml) = %v", err)
	}
	tmp = t.TempDir()
	writePasswordFile(t, tmp)
	return tmp, bytes.ReplaceAll(raw, []byte("@TMP@"), []byte(tmp))
}

// minimalYAML returns a config with only version, one repository and one
// root. top, repo and root are extra lines spliced into the top level, the
// repository and the root.
func minimalYAML(tmp, top, repo, root string) []byte {
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString(top)
	b.WriteString("repositories:\n")
	b.WriteString("  - id: personal\n")
	b.WriteString("    repository: rclone:gdrive:Backups/restic\n")
	b.WriteString("    restic_binary: /usr/bin/restic\n")
	b.WriteString("    rclone_binary: /usr/bin/rclone\n")
	b.WriteString("    password_file: " + filepath.Join(tmp, "password") + "\n")
	b.WriteString(repo)
	b.WriteString("roots:\n")
	b.WriteString("  - id: work\n")
	b.WriteString("    local_path: " + filepath.Join(tmp, "work") + "\n")
	b.WriteString("    repository_id: personal\n")
	b.WriteString(root)
	return []byte(b.String())
}

func TestParseExampleGolden(t *testing.T) {
	tmp, data := exampleYAML(t)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(example) = %v, want nil error", err)
	}

	want := &Config{
		Version:         1,
		LinkName:        ".snapshot",
		Timestamps:      "utc",
		StateDir:        tmp + "/state",
		HistoryMount:    tmp + "/state/mounts/history",
		BackendMountDir: tmp + "/state/mounts/repositories",
		Web:             Web{Enabled: true, Listen: "127.0.0.1:0", Bind: "127.0.0.1:0", OpenBrowser: false, AssetsDir: ""},
		Catalog: Catalog{
			RefreshInterval:      60 * time.Second,
			PrewarmSnapshots:     2,
			PrewarmConcurrency:   2,
			ProbeConcurrency:     4,
			PresenceCacheEntries: 10000,
			PresenceCacheTTL:     5 * time.Minute,
			ReaderPolicy: ReaderPolicy{
				DenyProcesses: []string{"rg", "fd", "find", "rsync", "mdworker", "mds"},
				BurstLimit:    50,
			},
		},
		Views: Views{
			Rsnapshot:     false,
			RsnapshotKeep: RsnapshotKeep{Hourly: 0, Daily: 7, Weekly: 4, Monthly: 6},
		},
		Discovery: Discovery{
			Mode:  "seed",
			Shell: true,
			Seed:  SeedSettings{InodeThreshold: 0.90, MaxLinksPerPath: 500000},
			OnAccess: OnAccess{
				AllowProcesses: []string{"zsh", "bash", "fish", "nautilus", "dolphin", "code"},
				HandlerTimeout: 20 * time.Millisecond,
			},
		},
		Repositories: []Repository{{
			ID:           "personal",
			Repository:   "rclone:gdrive:Backups/restic",
			ResticBinary: "/usr/bin/restic",
			RcloneBinary: "/usr/bin/rclone",
			PasswordFile: tmp + "/password",
			CacheDir:     tmp + "/cache/restic",
			NoCache:      false,
			LockMode:     "normal",
			Environment:  map[string]string{"RCLONE_CONFIG": tmp + "/rclone.conf"},
		}},
		Roots: []Root{{
			ID:           "work",
			LocalPath:    tmp + "/work",
			RepositoryID: "personal",
			PrefixMap: []PrefixMapping{
				{Hostname: "workstation", SourcePath: "/home/alex/work", TreePrefix: "/home/alex/work"},
				{Hostname: "alex-laptop", SourcePath: "/home/alex/src/work", TreePrefix: "/home/alex/work"},
			},
			Snapshots: SnapshotFilter{Hostname: "", TagsAll: []string{}, SourcePathsExact: []string{}},
			SeedPaths: []SeedPath{
				{Path: "project", MaxDepth: 6},
				{Path: "notes", MaxDepth: 3},
			},
			ExcludeRelativePaths: []string{".git", "node_modules"},
			Snap:                 SnapSettings{Tags: []string{"snapback:adhoc"}},
		}},
		Service: Service{Manager: "systemd", Scope: "user", RunAsUser: "alex"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(example) = %+v, want %+v", got, want)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	_, data := exampleYAML(t)
	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(example) = %v, want nil error", err)
	}

	b, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	c2, err := Parse(b)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error\n%s", err, b)
	}
	if !reflect.DeepEqual(c, c2) {
		t.Errorf("Parse(Marshal(c)) = %+v, want %+v", c2, c)
	}
	b2, err := Marshal(c2)
	if err != nil {
		t.Fatalf("Marshal(c2) = %v, want nil error", err)
	}
	if !bytes.Equal(b, b2) {
		t.Errorf("Marshal(c2) = %q, want byte-identical %q", b2, b)
	}
}

func TestParseRejectsDiscoveryFinder(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	data := minimalYAML(tmp, "discovery:\n  finder: true\n", "", "")

	_, err := Parse(data)
	if err == nil {
		t.Fatalf("Parse(discovery.finder) = nil error, want unknown-field error")
	}
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(Parse(discovery.finder)) = %q, want %q", got, errcode.InvalidConfig)
	}
	msg := err.Error()
	low := strings.ToLower(msg)
	if !strings.Contains(low, "finder") || !strings.Contains(low, "discovery") || !strings.Contains(msg, "line") {
		t.Errorf("Parse(discovery.finder) error = %q, want it to name discovery.finder and a line", msg)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	tests := []struct {
		name string
		data []byte
		key  string
	}{
		{"root seed_path", minimalYAML(tmp, "", "", "    seed_path: project\n"), "seed_path"},
		{"repository passwrd_file", minimalYAML(tmp, "", "    passwrd_file: /x\n", ""), "passwrd_file"},
		{"top-level extra", minimalYAML(tmp, "extra: 1\n", "", ""), "extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.data)
			if err == nil {
				t.Fatalf("Parse(%s) = nil error, want unknown-field error", tt.name)
			}
			if got := errcode.Of(err); got != errcode.InvalidConfig {
				t.Errorf("errcode.Of(Parse(%s)) = %q, want %q", tt.name, got, errcode.InvalidConfig)
			}
			if msg := err.Error(); !strings.Contains(msg, tt.key) || !strings.Contains(msg, "line") {
				t.Errorf("Parse(%s) error = %q, want it to name %q and a line", tt.name, msg, tt.key)
			}
		})
	}
}

func TestParseRejectsPlaintextPassword(t *testing.T) {
	_, data := exampleYAML(t)
	data = bytes.Replace(data, []byte("    lock_mode: normal\n"), []byte("    lock_mode: normal\n    password: secret\n"), 1)
	if !bytes.Contains(data, []byte("password: secret")) {
		t.Fatal("fixture splice failed: no password: secret line")
	}

	_, err := Parse(data)
	if err == nil {
		t.Fatal("Parse(example with password) = nil error, want error")
	}
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(err) = %q, want %q", got, errcode.InvalidConfig)
	}
	msg := err.Error()
	if !strings.Contains(msg, "password_file") {
		t.Errorf("Parse error = %q, want it to mention password_file", msg)
	}
	if strings.Contains(msg, "secret") {
		t.Errorf("Parse error = %q, want no plaintext password value", msg)
	}
}

func TestParseRejectsEmptyAndWrongType(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"version one", []byte("version: one\n")},
		{"prewarm many", minimalYAML(tmp, "catalog: {prewarm_snapshots: many}\n", "", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.data)
			if err == nil {
				t.Fatalf("Parse(%s) = nil error, want error", tt.name)
			}
			if got := errcode.Of(err); got != errcode.InvalidConfig {
				t.Errorf("errcode.Of(Parse(%s)) = %q, want %q", tt.name, got, errcode.InvalidConfig)
			}
		})
	}
}

func TestParseAppliesDefaults(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	xs := filepath.Join(tmp, "xs")
	t.Setenv("XDG_STATE_HOME", xs)

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(minimal) = %v, want nil error", err)
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"LinkName", c.LinkName, ".snapshot"},
		{"Timestamps", c.Timestamps, "utc"},
		{"Web.Enabled", c.Web.Enabled, true},
		{"Web.Listen", c.Web.Listen, "127.0.0.1:0"},
		{"Catalog.RefreshInterval", c.Catalog.RefreshInterval, 60 * time.Second},
		{"Catalog.PrewarmSnapshots", c.Catalog.PrewarmSnapshots, 2},
		{"Catalog.PrewarmConcurrency", c.Catalog.PrewarmConcurrency, 2},
		{"Catalog.ProbeConcurrency", c.Catalog.ProbeConcurrency, 4},
		{"Catalog.PresenceCacheEntries", c.Catalog.PresenceCacheEntries, 10000},
		{"Catalog.PresenceCacheTTL", c.Catalog.PresenceCacheTTL, 5 * time.Minute},
		{"ReaderPolicy.DenyProcesses", c.Catalog.ReaderPolicy.DenyProcesses, []string{"rg", "fd", "find", "rsync", "mdworker", "mds"}},
		{"ReaderPolicy.BurstLimit", c.Catalog.ReaderPolicy.BurstLimit, 50},
		{"Discovery.Mode", c.Discovery.Mode, "seed"},
		{"Discovery.Shell", c.Discovery.Shell, true},
		{"Seed.InodeThreshold", c.Discovery.Seed.InodeThreshold, 0.90},
		{"Seed.MaxLinksPerPath", c.Discovery.Seed.MaxLinksPerPath, 500000},
		{"Service", c.Service, Service{Manager: "auto", Scope: "user"}},
		{"StateDir", c.StateDir, xs + "/snapback"},
		{"HistoryMount", c.HistoryMount, xs + "/snapback/mounts/history"},
		{"BackendMountDir", c.BackendMountDir, xs + "/snapback/mounts/repositories"},
	}
	if len(c.Repositories) != 1 {
		t.Fatalf("Parse(minimal) repositories = %d, want 1", len(c.Repositories))
	}
	checks = append(checks, struct {
		name string
		got  any
		want any
	}{"Repositories[0].LockMode", c.Repositories[0].LockMode, "normal"})
	for _, ck := range checks {
		if !reflect.DeepEqual(ck.got, ck.want) {
			t.Errorf("Parse(minimal).%s = %v, want %v", ck.name, ck.got, ck.want)
		}
	}
}

func TestParseExplicitValuesOverrideDefaults(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))
	top := "link_name: \"@snapshot\"\nweb: {enabled: false}\ndiscovery: {shell: false}\n"

	c, err := Parse(minimalYAML(tmp, top, "", ""))
	if err != nil {
		t.Fatalf("Parse(minimal+overrides) = %v, want nil error", err)
	}
	if c.Web.Enabled {
		t.Errorf("Parse(web.enabled: false).Web.Enabled = true, want false")
	}
	if c.Discovery.Shell {
		t.Errorf("Parse(discovery.shell: false).Discovery.Shell = true, want false")
	}
	if c.LinkName != "@snapshot" {
		t.Errorf("Parse(link_name: @snapshot).LinkName = %q, want %q", c.LinkName, "@snapshot")
	}
}

func TestParseStateDirFallsBackToHome(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", tmp)

	c, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(minimal) = %v, want nil error", err)
	}
	if want := tmp + "/.local/state/snapback"; c.StateDir != want {
		t.Errorf("Parse(minimal).StateDir = %q, want %q", c.StateDir, want)
	}
}

func TestParseRejectsRelativeXDGStateHome(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", "relative")
	t.Setenv("HOME", tmp)

	_, err := Parse(minimalYAML(tmp, "", "", ""))
	if err == nil {
		t.Fatalf("Parse(XDG_STATE_HOME=relative) = nil error, want an error")
	}
	if !strings.Contains(err.Error(), "XDG_STATE_HOME must be an absolute path") {
		t.Errorf("Parse(XDG_STATE_HOME=relative) = %q, want XDG_STATE_HOME must be an absolute path", err.Error())
	}
}

func TestParseRunsValidate(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))
	data := bytes.Replace(minimalYAML(tmp, "", "", ""), []byte("  - id: work\n"), []byte("  - id: Bad\n"), 1)

	_, err := Parse(data)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("Parse(roots[0].id: Bad) = %v, want *ValidationError", err)
	}
	if findField(ve.Fields, "roots[0].id") == nil {
		t.Errorf("Parse(roots[0].id: Bad) fields = %+v, want one at roots[0].id", ve.Fields)
	}
}

func TestDefaultPath(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		name    string
		xdg     string
		home    string
		want    string
		wantErr bool
	}{
		{"xdg set", tmp + "/xc", tmp, tmp + "/xc/snapback/config.yaml", false},
		{"xdg empty", "", tmp, tmp + "/.config/snapback/config.yaml", false},
		{"xdg relative", "rel", tmp, tmp + "/.config/snapback/config.yaml", false},
		{"neither", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.xdg)
			t.Setenv("HOME", tt.home)

			got, err := DefaultPath()
			if tt.wantErr {
				if err == nil {
					t.Errorf("DefaultPath() = %q, nil, want error", got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("DefaultPath() = %q, %v, want %q, nil", got, err, tt.want)
			}
		})
	}
}
