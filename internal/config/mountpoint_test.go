package config

import (
	"bytes"
	"path/filepath"
	"strconv"
	"testing"
)

// mountPointErrors returns the field errors Validate reports on
// repositories[i].mount_point.
func mountPointErrors(t *testing.T, err error, i int) []FieldError {
	t.Helper()
	if err == nil {
		return nil
	}
	path := "repositories[" + strconv.Itoa(i) + "].mount_point"
	var out []FieldError
	for _, f := range validationFields(t, err) {
		if f.Path == path {
			out = append(out, f)
		}
	}
	return out
}

// TestRepositoryMountPointRoundTrip pins that an absolute mount_point survives
// Parse -> Marshal -> Parse and that an absent one stays absent (omitempty), so
// the example golden and TestMarshalRoundTrip keep passing.
func TestRepositoryMountPointRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	writePasswordFile(t, tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	c, err := Parse(minimalYAML(tmp, "", "    mount_point: /mnt/home-nas\n", ""))
	if err != nil {
		t.Fatalf("Parse(mount_point) = %v, want nil error", err)
	}
	if got, want := c.Repositories[0].MountPoint, "/mnt/home-nas"; got != want {
		t.Errorf("Repositories[0].MountPoint = %q, want %q", got, want)
	}

	out, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c) = %v, want nil error", err)
	}
	if !bytes.Contains(out, []byte("mount_point: /mnt/home-nas")) {
		t.Errorf("Marshal(c) = %s, want it to carry mount_point: /mnt/home-nas", out)
	}
	back, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Marshal(c)) = %v, want nil error", err)
	}
	if got, want := back.Repositories[0].MountPoint, "/mnt/home-nas"; got != want {
		t.Errorf("Parse(Marshal(c)).Repositories[0].MountPoint = %q, want %q", got, want)
	}

	absent, err := Parse(minimalYAML(tmp, "", "", ""))
	if err != nil {
		t.Fatalf("Parse(no mount_point) = %v, want nil error", err)
	}
	if got := absent.Repositories[0].MountPoint; got != "" {
		t.Errorf("Repositories[0].MountPoint = %q, want %q when the key is absent", got, "")
	}
	outAbsent, err := Marshal(absent)
	if err != nil {
		t.Fatalf("Marshal(absent) = %v, want nil error", err)
	}
	if bytes.Contains(outAbsent, []byte("mount_point")) {
		t.Errorf("Marshal(absent) = %s, want no mount_point key", outAbsent)
	}
}

// TestRepositoryMountPointPathRules pins that a mount point, when set, must be
// an absolute clean path, reported on repositories[0].mount_point.
func TestRepositoryMountPointPathRules(t *testing.T) {
	tests := []struct {
		name  string
		mount string
		valid bool
	}{
		{"absent", "", true},
		{"absolute", "/mnt/home-nas", true},
		{"relative", "mnt/x", false},
		{"dot relative", "./mnt/x", false},
		{"tilde", "~/mnt/x", false},
		{"trailing slash", "/mnt/home-nas/", false},
		{"dot dot", "/mnt/../mnt/home-nas", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Repositories[0].MountPoint = tt.mount

			got := mountPointErrors(t, Validate(c), 0)
			if tt.valid && len(got) != 0 {
				t.Errorf("Validate(mount_point=%q) = %+v, want no mount_point error", tt.mount, got)
			}
			if !tt.valid && len(got) == 0 {
				t.Errorf("Validate(mount_point=%q) = nil, want a field error on repositories[0].mount_point", tt.mount)
			}
		})
	}
}

// TestRepositoryMountPointOverlap pins that a mount point must not overlap the
// state dir, either mount dir or any root, in either containment direction, and
// that the message names the field it collides with.
func TestRepositoryMountPointOverlap(t *testing.T) {
	tests := []struct {
		name  string
		mount func(c *Config) string
		other string // the colliding field path, or "" when the config stays valid
	}{
		{"disjoint", func(*Config) string { return "/mnt/home-nas" }, ""},
		{"equals state_dir", func(c *Config) string { return c.StateDir }, "state_dir"},
		{"inside state_dir", func(c *Config) string { return filepath.Join(c.StateDir, "nas") }, "state_dir"},
		{"equals history_mount", func(c *Config) string { return c.HistoryMount }, "history_mount"},
		{"inside history_mount", func(c *Config) string { return filepath.Join(c.HistoryMount, "nas") }, "history_mount"},
		{"equals backend_mount_dir", func(c *Config) string { return c.BackendMountDir }, "backend_mount_dir"},
		{"inside backend_mount_dir", func(c *Config) string { return filepath.Join(c.BackendMountDir, "nas") }, "backend_mount_dir"},
		{"equals root local_path", func(c *Config) string { return c.Roots[0].LocalPath }, "roots[0].local_path"},
		{"inside root local_path", func(c *Config) string { return filepath.Join(c.Roots[0].LocalPath, "nas") }, "roots[0].local_path"},
		{"contains root local_path", func(c *Config) string { return filepath.Dir(c.Roots[0].LocalPath) }, "roots[0].local_path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			// A deeper root leaves room for the reverse-containment case
			// without also swallowing the state dir.
			c.Roots[0].LocalPath = filepath.Join(tmpOf(c), "work", "deep")
			c.Repositories[0].MountPoint = tt.mount(c)

			got := mountPointErrors(t, Validate(c), 0)
			if tt.other == "" {
				if len(got) != 0 {
					t.Errorf("Validate(mount_point=%q) = %+v, want no mount_point error", c.Repositories[0].MountPoint, got)
				}
				return
			}
			if countFieldErrors(got, "repositories[0].mount_point", tt.other) == 0 {
				t.Errorf("Validate(mount_point=%q) = %+v, want an error on repositories[0].mount_point naming %s",
					c.Repositories[0].MountPoint, got, tt.other)
			}
		})
	}
}

// TestRepositoryMountPointDuplicate pins that two repositories cannot share one
// mount point; the error lands on the second repository.
func TestRepositoryMountPointDuplicate(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		want   bool
	}{
		{"distinct", "/mnt/home-nas", "/mnt/off-site", false},
		{"both absent", "", "", false},
		{"one absent", "/mnt/home-nas", "", false},
		{"shared", "/mnt/home-nas", "/mnt/home-nas", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			second := c.Repositories[0]
			second.ID = "offsite"
			second.Repository = filepath.Join(tmpOf(c), "repo2")
			c.Repositories = append(c.Repositories, second)
			c.Repositories[0].MountPoint = tt.first
			c.Repositories[1].MountPoint = tt.second

			got := mountPointErrors(t, Validate(c), 1)
			if tt.want && len(got) == 0 {
				t.Errorf("Validate(%q, %q) = nil, want a field error on repositories[1].mount_point", tt.first, tt.second)
			}
			if !tt.want && len(got) != 0 {
				t.Errorf("Validate(%q, %q) = %+v, want no mount_point error", tt.first, tt.second, got)
			}
		})
	}
}
