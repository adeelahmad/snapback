package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// mountStatusConfig returns a config whose backend mount dir is under tmp and
// whose single repository mounts at mp.
func mountStatusConfig(tmp, id, mp string) *config.Config {
	return &config.Config{
		LinkName:        ".snapshot",
		BackendMountDir: filepath.Join(tmp, "backend"),
		Repositories:    []config.Repository{{ID: id, MountPoint: mp}},
	}
}

// TestMountPointStatusStates pins the four states resolved from the
// filesystem: disabled, linked, missing and conflict.
func TestMountPointStatusStates(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		wantRemedy []string
		setup      func(t *testing.T, tmp string) string // returns the mount point
	}{
		{
			name:  "empty mount point is disabled",
			state: "disabled",
			setup: func(t *testing.T, tmp string) string { return "" },
		},
		{
			name:  "symlink to an existing backend dir is linked",
			state: "linked",
			setup: func(t *testing.T, tmp string) string {
				target := filepath.Join(tmp, "backend", "nas")
				if err := os.MkdirAll(target, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) = %v", target, err)
				}
				mp := filepath.Join(tmp, "mnt", "nas")
				if err := os.MkdirAll(mp, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) = %v", mp, err)
				}
				if err := os.Symlink(target, filepath.Join(mp, ".snapshot")); err != nil {
					t.Fatalf("Symlink = %v", err)
				}
				return mp
			},
		},
		{
			name:       "symlink to a missing target is missing",
			state:      "missing",
			wantRemedy: []string{"snapback run"},
			setup: func(t *testing.T, tmp string) string {
				mp := filepath.Join(tmp, "mnt", "nas")
				if err := os.MkdirAll(mp, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) = %v", mp, err)
				}
				target := filepath.Join(tmp, "backend", "nas")
				if err := os.Symlink(target, filepath.Join(mp, ".snapshot")); err != nil {
					t.Fatalf("Symlink = %v", err)
				}
				return mp
			},
		},
		{
			name:       "absent link is missing",
			state:      "missing",
			wantRemedy: []string{"snapback run"},
			setup: func(t *testing.T, tmp string) string {
				mp := filepath.Join(tmp, "mnt", "nas")
				if err := os.MkdirAll(mp, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) = %v", mp, err)
				}
				return mp
			},
		},
		{
			name:  "a regular file under the link name is a conflict",
			state: "conflict",
			setup: func(t *testing.T, tmp string) string {
				mp := filepath.Join(tmp, "mnt", "nas")
				if err := os.MkdirAll(mp, 0o755); err != nil {
					t.Fatalf("MkdirAll(%q) = %v", mp, err)
				}
				if err := os.WriteFile(filepath.Join(mp, ".snapshot"), []byte("x"), 0o644); err != nil {
					t.Fatalf("WriteFile = %v", err)
				}
				return mp
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			mp := tt.setup(t, tmp)
			cfg := mountStatusConfig(tmp, "nas", mp)

			got := MountPointStatuses(cfg)
			if len(got) != 1 {
				t.Fatalf("MountPointStatuses() = %+v, want exactly 1 entry", got)
			}
			if got[0].Repository != "nas" {
				t.Errorf("Repository = %q, want %q", got[0].Repository, "nas")
			}
			if got[0].MountPoint != mp {
				t.Errorf("MountPoint = %q, want %q", got[0].MountPoint, mp)
			}
			if got[0].State != tt.state {
				t.Errorf("State = %q, want %q", got[0].State, tt.state)
			}
			for _, want := range tt.wantRemedy {
				if !strings.Contains(got[0].Remedy, want) {
					t.Errorf("Remedy = %q, want it to contain %q", got[0].Remedy, want)
				}
			}
			if tt.state == "conflict" && !strings.Contains(got[0].Remedy, filepath.Join(mp, ".snapshot")) {
				t.Errorf("Remedy = %q, want it to name %q", got[0].Remedy, filepath.Join(mp, ".snapshot"))
			}
			if tt.state == "linked" && got[0].Remedy != "" {
				t.Errorf("Remedy = %q, want %q for a linked mount point", got[0].Remedy, "")
			}
		})
	}
}

// TestMountPointStatusHumanLines pins one "mount point <mp>: <state>" line per
// repository in the human status output.
func TestMountPointStatusHumanLines(t *testing.T) {
	tmp := t.TempDir()
	cfg := mountStatusConfig(tmp, "nas", filepath.Join(tmp, "mnt", "nas"))
	cfg.Repositories = append(cfg.Repositories, config.Repository{ID: "offsite"})

	got := RenderMountPoints(MountPointStatuses(cfg))

	for _, want := range []string{
		"mount point " + filepath.Join(tmp, "mnt", "nas") + ": missing",
		"mount point : disabled",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("RenderMountPoints() = %q, want it to contain %q", got, want)
		}
	}
	if n := strings.Count(got, "mount point "); n != 2 {
		t.Errorf("RenderMountPoints() = %q, want 2 mount point lines, got %d", got, n)
	}
}

// TestMountPointStatusJSON pins that the JSON payload carries mount_points
// entries with snake_case keys.
func TestMountPointStatusJSON(t *testing.T) {
	tmp := t.TempDir()
	mp := filepath.Join(tmp, "mnt", "nas")
	cfg := mountStatusConfig(tmp, "nas", mp)

	blob, err := json.Marshal(struct {
		MountPoints []MountPointStatus `json:"mount_points"`
	}{MountPoints: MountPointStatuses(cfg)})
	if err != nil {
		t.Fatalf("Marshal = %v", err)
	}

	var back struct {
		MountPoints []struct {
			Repository string `json:"repository"`
			MountPoint string `json:"mount_point"`
			State      string `json:"state"`
			Remedy     string `json:"remedy"`
		} `json:"mount_points"`
	}
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatalf("Unmarshal(%s) = %v", blob, err)
	}
	if len(back.MountPoints) != 1 {
		t.Fatalf("mount_points = %+v, want exactly 1 entry", back.MountPoints)
	}
	e := back.MountPoints[0]
	if e.Repository != "nas" || e.MountPoint != mp || e.State != "missing" {
		t.Errorf("mount_points[0] = %+v, want repository %q, mount_point %q, state %q", e, "nas", mp, "missing")
	}
	if !strings.Contains(e.Remedy, "snapback run") {
		t.Errorf("mount_points[0].remedy = %q, want it to contain %q", e.Remedy, "snapback run")
	}
}
