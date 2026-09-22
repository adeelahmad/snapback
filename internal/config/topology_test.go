package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// validConfig returns a fully valid *Config whose paths all live under a
// fresh t.TempDir(), with a 0600 password file at <tmp>/password.
func validConfig(t *testing.T) *Config {
	t.Helper()
	tmp := t.TempDir()
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	if err := os.Chmod(pw, 0o600); err != nil {
		t.Fatalf("chmod password file: %v", err)
	}
	state := filepath.Join(tmp, "state")
	return &Config{
		Version:         1,
		LinkName:        ".snapshot",
		Timestamps:      "utc",
		StateDir:        state,
		HistoryMount:    filepath.Join(state, "mounts", "history"),
		BackendMountDir: filepath.Join(state, "mounts", "repositories"),
		Web:             Web{Enabled: true, Listen: "127.0.0.1:0"},
		Catalog: Catalog{
			RefreshInterval:      60 * time.Second,
			PrewarmSnapshots:     2,
			PrewarmConcurrency:   2,
			ProbeConcurrency:     4,
			PresenceCacheEntries: 10000,
			PresenceCacheTTL:     5 * time.Minute,
			ReaderPolicy:         ReaderPolicy{DenyProcesses: []string{"rg"}, BurstLimit: 50},
		},
		Views: Views{RsnapshotKeep: RsnapshotKeep{Daily: 7, Weekly: 4, Monthly: 6}},
		Discovery: Discovery{
			Mode:     "seed",
			Shell:    true,
			Seed:     SeedSettings{InodeThreshold: 0.90, MaxLinksPerPath: 500000},
			OnAccess: OnAccess{HandlerTimeout: 20 * time.Millisecond},
		},
		Repositories: []Repository{{
			ID:           "main",
			Repository:   filepath.Join(tmp, "repo"),
			ResticBinary: "/usr/bin/restic",
			PasswordFile: pw,
			LockMode:     "normal",
		}},
		Roots: []Root{{
			ID:           "work",
			LocalPath:    filepath.Join(tmp, "work"),
			RepositoryID: "main",
			SeedPaths:    []SeedPath{{Path: "src", MaxDepth: 2}},
		}},
		Service: Service{Manager: "auto", Scope: "user"},
	}
}

// tmpOf returns the t.TempDir() that validConfig built c under.
func tmpOf(c *Config) string {
	return filepath.Dir(c.StateDir)
}

// countFieldErrors returns how many errs have the given path and a message
// containing substr.
func countFieldErrors(errs []FieldError, path, substr string) int {
	n := 0
	for _, e := range errs {
		if e.Path == path && strings.Contains(e.Msg, substr) {
			n++
		}
	}
	return n
}

func TestWithinIsComponentWise(t *testing.T) {
	tests := []struct {
		parent, child string
		want          bool
	}{
		{"/a/b", "/a/b", true},
		{"/a/b", "/a/b/c", true},
		{"/a/b", "/a/bc", false},
		{"/a/b/", "/a/b/c", true},
		{"/", "/x", true},
		{"/a/b/c", "/a/b", false},
	}
	for _, tt := range tests {
		if got := within(tt.parent, tt.child); got != tt.want {
			t.Errorf("within(%q, %q) = %v, want %v", tt.parent, tt.child, got, tt.want)
		}
	}
}

func TestTopologyValidConfigHasNoErrors(t *testing.T) {
	c := validConfig(t)
	tmp := tmpOf(c)
	c.HistoryMount = filepath.Join(tmp, "state", "mounts", "history")
	c.BackendMountDir = filepath.Join(tmp, "state", "mounts", "repositories")
	c.Roots[0].LocalPath = filepath.Join(tmp, "work")

	if got := checkTopology(c); len(got) != 0 {
		t.Errorf("checkTopology(valid) = %+v, want none", got)
	}
}

func TestTopologyMountAtOrAboveRoot(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *Config)
		path   string
	}{
		{"history equals root", func(c *Config) { c.HistoryMount = c.Roots[0].LocalPath }, "history_mount"},
		{"history is parent of root", func(c *Config) { c.HistoryMount = filepath.Dir(c.Roots[0].LocalPath) }, "history_mount"},
		{"backend is filesystem root", func(c *Config) { c.BackendMountDir = "/" }, "backend_mount_dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			got := checkTopology(c)
			if n := countFieldErrors(got, tt.path, "roots[0].local_path"); n != 1 {
				t.Errorf("checkTopology() = %+v, want exactly one %s error naming roots[0].local_path (got %d)", got, tt.path, n)
			}
		})
	}
}

func TestTopologyMountInsideRootAllowed(t *testing.T) {
	c := validConfig(t)
	c.HistoryMount = filepath.Join(c.Roots[0].LocalPath, ".snapback", "history")

	for _, e := range checkTopology(c) {
		if e.Path == "history_mount" {
			t.Errorf("checkTopology(history inside root) reported %+v, want no history_mount error", e)
		}
	}
}

func TestTopologyMountsDisjoint(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *Config)
	}{
		{"backend equals history", func(c *Config) { c.BackendMountDir = c.HistoryMount }},
		{"backend inside history", func(c *Config) { c.BackendMountDir = filepath.Join(c.HistoryMount, "repos") }},
		{"history inside backend", func(c *Config) { c.HistoryMount = filepath.Join(c.BackendMountDir, "history") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			got := checkTopology(c)
			if countFieldErrors(got, "backend_mount_dir", "history_mount") == 0 {
				t.Errorf("checkTopology() = %+v, want a backend_mount_dir error naming history_mount", got)
			}
		})
	}
}

func TestTopologyRepositoryStorageOverlap(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		path    string
		wantErr bool
	}{
		{
			name: "history inside local repository",
			mutate: func(c *Config) {
				repo := filepath.Join(tmpOf(c), "repo")
				c.Repositories[0].Repository = repo
				c.HistoryMount = filepath.Join(repo, "h")
			},
			path:    "history_mount",
			wantErr: true,
		},
		{
			name: "backend contains local: repository",
			mutate: func(c *Config) {
				c.Repositories[0].Repository = "local:" + filepath.Join(tmpOf(c), "repo")
				c.BackendMountDir = tmpOf(c)
			},
			path:    "backend_mount_dir",
			wantErr: true,
		},
		{
			name: "rclone repository has no local storage",
			mutate: func(c *Config) {
				c.Repositories[0].Repository = "rclone:gdrive:x"
				c.Repositories[0].RcloneBinary = "/usr/bin/rclone"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			got := checkTopology(c)
			if !tt.wantErr {
				if len(got) != 0 {
					t.Errorf("checkTopology() = %+v, want none", got)
				}
				return
			}
			if countFieldErrors(got, tt.path, "repositories[0].repository") == 0 {
				t.Errorf("checkTopology() = %+v, want a %s error naming repositories[0].repository", got, tt.path)
			}
		})
	}
}

func TestTopologyStateDirNotInsideMount(t *testing.T) {
	c := validConfig(t)
	c.StateDir = filepath.Join(c.HistoryMount, "s")

	got := checkTopology(c)
	if countFieldErrors(got, "state_dir", "") == 0 {
		t.Errorf("checkTopology() = %+v, want a state_dir error", got)
	}
}

func TestCredentialsPasswordFile(t *testing.T) {
	const path = "repositories[0].password_file"
	tests := []struct {
		name    string
		setup   func(t *testing.T, p string)
		wantMsg string // "" means no error
	}{
		{"mode 0600", writeMode(0o600), ""},
		{"mode 0400", writeMode(0o400), ""},
		{"mode 0640", writeMode(0o640), "0600"},
		{"mode 0604", writeMode(0o604), "0600"},
		{"missing", func(t *testing.T, p string) {}, "does not exist"},
		{"directory", func(t *testing.T, p string) {
			if err := os.Mkdir(p, 0o700); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
		}, "regular file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			p := filepath.Join(t.TempDir(), "pw")
			tt.setup(t, p)
			c.Repositories[0].PasswordFile = p

			got := checkCredentials(c)
			if tt.wantMsg == "" {
				if len(got) != 0 {
					t.Errorf("checkCredentials(%s) = %+v, want none", tt.name, got)
				}
				return
			}
			if countFieldErrors(got, path, tt.wantMsg) == 0 {
				t.Errorf("checkCredentials(%s) = %+v, want a %s error containing %q", tt.name, got, path, tt.wantMsg)
			}
		})
	}
}

func TestCredentialsMessageHasNoContents(t *testing.T) {
	const contents = "hunter2-secret"
	c := validConfig(t)
	p := filepath.Join(t.TempDir(), "pw")
	writeMode(0o644)(t, p)
	if err := os.WriteFile(p, []byte(contents), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	c.Repositories[0].PasswordFile = p

	got := checkCredentials(c)
	if len(got) == 0 {
		t.Fatalf("checkCredentials(mode 0644) = none, want an error")
	}
	for _, e := range got {
		if strings.Contains(e.Msg, contents) {
			t.Errorf("checkCredentials() message %q contains the file contents", e.Msg)
		}
	}
}

// writeMode returns a setup func that writes a small fixture file with the
// given permission bits, set explicitly so the umask cannot interfere.
func writeMode(mode os.FileMode) func(t *testing.T, p string) {
	return func(t *testing.T, p string) {
		t.Helper()
		if err := os.WriteFile(p, []byte("fixture"), 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := os.Chmod(p, mode); err != nil {
			t.Fatalf("chmod: %v", err)
		}
	}
}
