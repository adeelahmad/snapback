package setup

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// existingFixture writes a configuration built from a detection result to a
// file and returns the path and the configuration a second setup would want.
func existingFixture(t *testing.T) (string, *config.Config) {
	t.Helper()
	res, state := detectedResult(t)
	cfg, err := ToConfig(res, Options{StateDir: state})
	if err != nil {
		t.Fatalf("ToConfig(detected) error = %v, want nil", err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save(cfg, path) error = %v, want nil", err)
	}
	return path, cfg
}

func TestExisting(t *testing.T) {
	path, cfg := existingFixture(t)

	changedRepo := *cfg
	changedRepo.Repositories = append([]config.Repository(nil), cfg.Repositories...)
	changedRepo.Repositories[0].Repository = "sftp:backup@other:/srv/restic"

	changedRoot := *cfg
	changedRoot.Roots = append([]config.Root(nil), cfg.Roots...)
	changedRoot.Roots[0].LocalPath = "/elsewhere/projects"

	tests := []struct {
		name       string
		path       string
		want       *config.Config
		wantExists bool
		wantSame   bool
		wantLines  []string
	}{
		{
			name:       "missing",
			path:       filepath.Join(t.TempDir(), "absent.yaml"),
			want:       cfg,
			wantExists: false,
			wantSame:   false,
		},
		{
			name:       "same",
			path:       path,
			want:       cfg,
			wantExists: true,
			wantSame:   true,
			wantLines: []string{
				"config: " + path + " (unchanged)",
				"repository: " + cfg.Repositories[0].Repository,
				"root: " + cfg.Roots[0].LocalPath,
			},
		},
		{
			name:       "repository differs",
			path:       path,
			want:       &changedRepo,
			wantExists: true,
			wantSame:   false,
			wantLines: []string{
				"config: " + path + " (differs)",
				"repository: " + cfg.Repositories[0].Repository + " → " + changedRepo.Repositories[0].Repository,
			},
		},
		{
			name:       "root differs",
			path:       path,
			want:       &changedRoot,
			wantExists: true,
			wantSame:   false,
			wantLines: []string{
				"config: " + path + " (differs)",
				"root: " + cfg.Roots[0].LocalPath + " → " + changedRoot.Roots[0].LocalPath,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Existing(tc.path, tc.want)
			if err != nil {
				t.Fatalf("Existing(%q, want) error = %v, want nil", tc.path, err)
			}
			if got.Exists != tc.wantExists {
				t.Errorf("Existing(%q, want).Exists = %v, want %v", tc.path, got.Exists, tc.wantExists)
			}
			if got.Same != tc.wantSame {
				t.Errorf("Existing(%q, want).Same = %v, want %v", tc.path, got.Same, tc.wantSame)
			}
			if tc.wantLines == nil {
				if len(got.Lines) != 0 {
					t.Errorf("Existing(%q, want).Lines = %v, want none", tc.path, got.Lines)
				}
				return
			}
			if !reflect.DeepEqual(got.Lines, tc.wantLines) {
				t.Errorf("Existing(%q, want).Lines = %v, want %v", tc.path, got.Lines, tc.wantLines)
			}
		})
	}
}

func TestExistingDoesNotWrite(t *testing.T) {
	path, cfg := existingFixture(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}

	got, err := Existing(path, cfg)
	if err != nil {
		t.Fatalf("Existing(path, cfg) error = %v, want nil", err)
	}
	if !got.Same {
		t.Fatalf("Existing(path, cfg).Same = false, want true")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture after: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("Existing(path, cfg) rewrote the file: got %q, want %q", after, before)
	}
	info2, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fixture after: %v", err)
	}
	if !info2.ModTime().Equal(info.ModTime()) {
		t.Errorf("Existing(path, cfg) ModTime = %v, want %v", info2.ModTime(), info.ModTime())
	}
}
