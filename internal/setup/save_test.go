package setup

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/adeelahmad/snapback/internal/config"
)

// saveConfig returns a configuration that passes validation, with the
// repository lock mode left empty so defaults are observable after a save.
func saveConfig(t *testing.T) *config.Config {
	t.Helper()
	tmp := t.TempDir()
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("write password file: %v", err)
	}
	state := filepath.Join(tmp, "state")
	return &config.Config{
		Version:         1,
		LinkName:        ".snapshot",
		Timestamps:      "utc",
		StateDir:        state,
		HistoryMount:    filepath.Join(state, "mounts", "history"),
		BackendMountDir: filepath.Join(state, "mounts", "repositories"),
		Web:             config.Web{Enabled: true, Listen: "127.0.0.1:0"},
		Catalog: config.Catalog{
			RefreshInterval:      60 * time.Second,
			PrewarmSnapshots:     2,
			PrewarmConcurrency:   2,
			ProbeConcurrency:     4,
			PresenceCacheEntries: 10000,
			PresenceCacheTTL:     5 * time.Minute,
			ReaderPolicy:         config.ReaderPolicy{DenyProcesses: []string{"rg"}, BurstLimit: 50},
		},
		Views: config.Views{RsnapshotKeep: config.RsnapshotKeep{Daily: 7, Weekly: 4, Monthly: 6}},
		Discovery: config.Discovery{
			Mode:     "seed",
			Shell:    true,
			Seed:     config.SeedSettings{InodeThreshold: 0.90, MaxLinksPerPath: 500000},
			OnAccess: config.OnAccess{AllowProcesses: []string{}, HandlerTimeout: 20 * time.Millisecond},
		},
		Repositories: []config.Repository{{
			ID:           "main",
			Repository:   filepath.Join(tmp, "repo"),
			ResticBinary: "/usr/bin/restic",
			PasswordFile: pw,
			Environment:  map[string]string{},
		}},
		Roots: []config.Root{{
			ID:                   "work",
			LocalPath:            filepath.Join(tmp, "work"),
			RepositoryID:         "main",
			PrefixMap:            []config.PrefixMapping{},
			Snapshots:            config.SnapshotFilter{TagsAll: []string{}, SourcePathsExact: []string{}},
			SeedPaths:            []config.SeedPath{},
			ExcludeRelativePaths: []string{},
			Snap:                 config.SnapSettings{Tags: []string{}},
		}},
		Service: config.Service{Manager: "auto", Scope: "user"},
	}
}

func TestSaveWritesThroughTheConfigLayer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg", "config.yaml")
	cfg := saveConfig(t)

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save(cfg, %q) = %v, want nil error", path, err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) = %v, want the saved file", path, err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("Stat(%q).Mode() = %04o, want 0600", path, got)
	}
	got, _, err := config.Load(path)
	if err != nil {
		t.Fatalf("config.Load(%q) = %v, want nil error", path, err)
	}
	if want := "normal"; got.Repositories[0].LockMode != want {
		t.Errorf("Load(%q).Repositories[0].LockMode = %q, want %q", path, got.Repositories[0].LockMode, want)
	}
	if !reflect.DeepEqual(got, cfg) {
		t.Errorf("config.Load(%q) = %+v, want %+v", path, got, cfg)
	}
}

func TestSavePreservesUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg", "config.yaml")
	cfg := saveConfig(t)
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save(cfg, %q) = %v, want nil error", path, err)
	}
	existing, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v", path, err)
	}
	if err := os.WriteFile(path, append(existing, []byte("x-custom: 1\n")...), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	cfg.LinkName = ".history"
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save(cfg, %q) = %v, want nil error", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) = %v", path, err)
	}
	var merged map[string]any
	if err := yaml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("yaml.Unmarshal(saved) = %v, want nil error", err)
	}
	if got, want := merged["x-custom"], 1; got != want {
		t.Errorf("saved[\"x-custom\"] = %v, want %v", got, want)
	}
	if got, want := merged["link_name"], ".history"; got != want {
		t.Errorf("saved[\"link_name\"] = %v, want %v", got, want)
	}
}

func TestSaveIsTheOnlySerialiser(t *testing.T) {
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob(*.go) = %v", err)
	}
	for _, name := range entries {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("ReadFile(%q) = %v", name, err)
		}
		body := string(src)
		if name != "save.go" {
			for _, bad := range []string{"yaml.Marshal(", "yaml.NewEncoder("} {
				if strings.Contains(body, bad) {
					t.Errorf("%s contains %q, want YAML written only through the config layer", name, bad)
				}
			}
			continue
		}
		if strings.Contains(body, "yaml.Marshal(") {
			t.Errorf("%s contains \"yaml.Marshal(\", want the struct serialised only by config.Save", name)
		}
		merge := strings.Index(body, "func mergeUnknownKeys(")
		if merge < 0 && strings.Contains(body, "yaml.NewEncoder(") {
			t.Fatalf("%s uses yaml.NewEncoder outside mergeUnknownKeys", name)
		}
		if enc := strings.Index(body, "yaml.NewEncoder("); enc >= 0 && enc < merge {
			t.Errorf("%s uses yaml.NewEncoder at %d, want it only inside mergeUnknownKeys at %d", name, enc, merge)
		}
	}
}
