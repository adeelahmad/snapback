package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// validationFields returns the fields of the *ValidationError in err, or
// fails the test when err is not one.
func validationFields(t *testing.T, err error) []FieldError {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("Validate() = %v, want a *ValidationError", err)
	}
	return ve.Fields
}

// findField returns the first field whose Path equals path, or nil.
func findField(fields []FieldError, path string) *FieldError {
	for i := range fields {
		if fields[i].Path == path {
			return &fields[i]
		}
	}
	return nil
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	if err := Validate(validConfig(t)); err != nil {
		t.Errorf("Validate(valid) = %v, want nil", err)
	}
}

func TestValidateFieldErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *Config)
		path   string
	}{
		{"version 2", func(c *Config) { c.Version = 2 }, "version"},
		{"link_name empty", func(c *Config) { c.LinkName = "" }, "link_name"},
		{"link_name dot", func(c *Config) { c.LinkName = "." }, "link_name"},
		{"link_name slash", func(c *Config) { c.LinkName = "a/b" }, "link_name"},
		{"timestamps gmt", func(c *Config) { c.Timestamps = "gmt" }, "timestamps"},
		{"state_dir tilde", func(c *Config) { c.StateDir = "~/s" }, "state_dir"},
		{"state_dir relative", func(c *Config) { c.StateDir = "rel/s" }, "state_dir"},
		{"history_mount relative", func(c *Config) { c.HistoryMount = "rel" }, "history_mount"},
		{"web.listen all interfaces", func(c *Config) { c.Web.Listen = "0.0.0.0:8080" }, "web.listen"},
		{"web.listen nonsense", func(c *Config) { c.Web.Listen = "nonsense" }, "web.listen"},
		{"refresh_interval zero", func(c *Config) { c.Catalog.RefreshInterval = 0 }, "catalog.refresh_interval"},
		{"prewarm_snapshots negative", func(c *Config) { c.Catalog.PrewarmSnapshots = -1 }, "catalog.prewarm_snapshots"},
		{"prewarm_concurrency zero", func(c *Config) { c.Catalog.PrewarmConcurrency = 0 }, "catalog.prewarm_concurrency"},
		{"probe_concurrency zero", func(c *Config) { c.Catalog.ProbeConcurrency = 0 }, "catalog.probe_concurrency"},
		{"presence_cache_entries negative", func(c *Config) { c.Catalog.PresenceCacheEntries = -1 }, "catalog.presence_cache_entries"},
		{"presence_cache_ttl negative", func(c *Config) { c.Catalog.PresenceCacheTTL = -1 }, "catalog.presence_cache_ttl"},
		{"burst_limit zero", func(c *Config) { c.Catalog.ReaderPolicy.BurstLimit = 0 }, "catalog.reader_policy.burst_limit"},
		{"rsnapshot daily negative", func(c *Config) { c.Views.RsnapshotKeep.Daily = -1 }, "views.rsnapshot_keep.daily"},
		{"discovery.mode lazy", func(c *Config) { c.Discovery.Mode = "lazy" }, "discovery.mode"},
		{"inode_threshold zero", func(c *Config) { c.Discovery.Seed.InodeThreshold = 0 }, "discovery.seed.inode_threshold"},
		{"inode_threshold above one", func(c *Config) { c.Discovery.Seed.InodeThreshold = 1.5 }, "discovery.seed.inode_threshold"},
		{"max_links_per_path zero", func(c *Config) { c.Discovery.Seed.MaxLinksPerPath = 0 }, "discovery.seed.max_links_per_path"},
		{"service.scope global", func(c *Config) { c.Service.Scope = "global" }, "service.scope"},
		{"service.manager upstart", func(c *Config) { c.Service.Manager = "upstart" }, "service.manager"},
		{"lock_mode off", func(c *Config) { c.Repositories[0].LockMode = "off" }, "repositories[0].lock_mode"},
		{"restic_binary relative", func(c *Config) { c.Repositories[0].ResticBinary = "restic" }, "repositories[0].restic_binary"},
		{"rclone_binary missing for rclone repo", func(c *Config) {
			c.Repositories[0].Repository = "rclone:remote:repo"
			c.Repositories[0].RcloneBinary = ""
		}, "repositories[0].rclone_binary"},
		{"cache_dir tilde", func(c *Config) { c.Repositories[0].CacheDir = "~/c" }, "repositories[0].cache_dir"},
		{"repository empty", func(c *Config) { c.Repositories[0].Repository = "" }, "repositories[0].repository"},
		{"local_path tilde", func(c *Config) { c.Roots[0].LocalPath = "~/work" }, "roots[0].local_path"},
		{"repository_id unknown", func(c *Config) { c.Roots[0].RepositoryID = "nope" }, "roots[0].repository_id"},
		{"tree_prefix relative", func(c *Config) {
			c.Roots[0].PrefixMap = []PrefixMapping{{SourcePath: "/home/x", TreePrefix: "home/x"}}
		}, "roots[0].prefix_map[0].tree_prefix"},
		{"source_paths_exact relative", func(c *Config) {
			c.Roots[0].Snapshots.SourcePathsExact = []string{"rel"}
		}, "roots[0].snapshots.source_paths_exact[0]"},
		{"seed path absolute", func(c *Config) { c.Roots[0].SeedPaths[0].Path = "/abs" }, "roots[0].seed_paths[0].path"},
		{"seed path parent", func(c *Config) { c.Roots[0].SeedPaths[0].Path = "../up" }, "roots[0].seed_paths[0].path"},
		{"seed max_depth zero", func(c *Config) { c.Roots[0].SeedPaths[0].MaxDepth = 0 }, "roots[0].seed_paths[0].max_depth"},
		{"exclude parent", func(c *Config) {
			c.Roots[0].ExcludeRelativePaths = []string{"../x"}
		}, "roots[0].exclude_relative_paths[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			got := validationFields(t, Validate(c))
			if findField(got, tt.path) == nil {
				t.Errorf("Validate(%s) fields = %+v, want one at %s", tt.name, got, tt.path)
			}
		})
	}
}

func TestValidateIDGrammar(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"a", true},
		{"work", true},
		{"w0_-x", true},
		{strings.Repeat("a", 32), true},
		{"", false},
		{"Work", false},
		{"0a", false},
		{"-a", false},
		{"a b", false},
		{"a.b", false},
		{strings.Repeat("a", 33), false},
	}
	for _, tt := range tests {
		t.Run("repository "+tt.id, func(t *testing.T) {
			c := validConfig(t)
			c.Repositories[0].ID = tt.id
			c.Roots[0].RepositoryID = tt.id
			checkID(t, c, tt.id, tt.valid, "repositories[0].id")
		})
		t.Run("root "+tt.id, func(t *testing.T) {
			c := validConfig(t)
			c.Roots[0].ID = tt.id
			checkID(t, c, tt.id, tt.valid, "roots[0].id")
		})
	}
}

func checkID(t *testing.T, c *Config, id string, valid bool, path string) {
	t.Helper()
	err := Validate(c)
	if valid {
		if err != nil {
			t.Errorf("Validate(%s=%q) = %v, want nil", path, id, err)
		}
		return
	}
	if got := validationFields(t, err); findField(got, path) == nil {
		t.Errorf("Validate(%s=%q) fields = %+v, want one at %s", path, id, got, path)
	}
}

func TestValidateIDsUnique(t *testing.T) {
	t.Run("duplicate repositories", func(t *testing.T) {
		c := validConfig(t)
		c.Repositories[0].ID = "personal"
		c.Roots[0].RepositoryID = "personal"
		second := c.Repositories[0]
		second.Repository = filepath.Join(tmpOf(c), "repo2")
		c.Repositories = append(c.Repositories, second)
		got := validationFields(t, Validate(c))
		if f := findField(got, "repositories[1].id"); f == nil || !strings.Contains(f.Msg, "duplicate") {
			t.Errorf("Validate(two personal repositories) fields = %+v, want repositories[1].id containing %q", got, "duplicate")
		}
	})
	t.Run("duplicate roots", func(t *testing.T) {
		c := validConfig(t)
		second := c.Roots[0]
		second.LocalPath = filepath.Join(tmpOf(c), "work2")
		c.Roots = append(c.Roots, second)
		got := validationFields(t, Validate(c))
		if findField(got, "roots[1].id") == nil {
			t.Errorf("Validate(two work roots) fields = %+v, want one at roots[1].id", got)
		}
	})
	t.Run("root and repository share an id", func(t *testing.T) {
		c := validConfig(t)
		c.Repositories[0].ID = "x"
		c.Roots[0].ID = "x"
		c.Roots[0].RepositoryID = "x"
		if err := Validate(c); err != nil {
			t.Errorf("Validate(root and repository both x) = %v, want nil", err)
		}
	})
}

func TestValidateRequiresRepositoryAndRoot(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *Config)
		path   string
	}{
		{"no repositories", func(c *Config) { c.Repositories = nil }, "repositories"},
		{"no roots", func(c *Config) { c.Roots = nil }, "roots"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			got := validationFields(t, Validate(c))
			if findField(got, tt.path) == nil {
				t.Errorf("Validate(%s) fields = %+v, want one at %s", tt.name, got, tt.path)
			}
		})
	}
}

func TestValidateReportsEveryError(t *testing.T) {
	c := validConfig(t)
	c.Timestamps = "gmt"
	c.Roots[0].ID = "Bad"
	c.Catalog.ProbeConcurrency = 0

	got := validationFields(t, Validate(c))
	if len(got) != 3 {
		t.Errorf("Validate(three bad fields) = %d fields %+v, want 3", len(got), got)
	}
	for _, path := range []string{"timestamps", "roots[0].id", "catalog.probe_concurrency"} {
		if findField(got, path) == nil {
			t.Errorf("Validate(three bad fields) fields = %+v, want one at %s", got, path)
		}
	}
}

func TestValidateV01RejectionsCarryNamedCodes(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(c *Config)
		path     string
		wantCode errcode.Code
		valid    bool
	}{
		{"on-access", func(c *Config) { c.Discovery.Mode = "on-access" }, "discovery.mode", errcode.OnAccessUnavailable, false},
		{"launchd", func(c *Config) { c.Service.Manager = "launchd" }, "service.manager", errcode.UnsupportedServiceManager, false},
		{"openrc", func(c *Config) { c.Service.Manager = "openrc" }, "service.manager", errcode.UnsupportedServiceManager, false},
		{"systemd", func(c *Config) { c.Service.Manager = "systemd" }, "", "", true},
		{"auto", func(c *Config) { c.Service.Manager = "auto" }, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			tt.mutate(c)
			err := Validate(c)
			if tt.valid {
				if err != nil {
					t.Errorf("Validate(%s) = %v, want nil", tt.name, err)
				}
				return
			}
			got := validationFields(t, err)
			f := findField(got, tt.path)
			if f == nil {
				t.Fatalf("Validate(%s) fields = %+v, want one at %s", tt.name, got, tt.path)
			}
			if f.Code != tt.wantCode {
				t.Errorf("Validate(%s) %s Code = %q, want %q", tt.name, tt.path, f.Code, tt.wantCode)
			}
		})
	}
}

func TestValidateEnvironmentKeys(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		valid   bool
		prefix  string
		wantMsg string
	}{
		{"rclone config", map[string]string{"RCLONE_CONFIG": "/x"}, true, "", ""},
		{"restic password", map[string]string{"RESTIC_PASSWORD": "p"}, false, "repositories[0].environment.RESTIC_PASSWORD", "password_file"},
		{"restic password command", map[string]string{"RESTIC_PASSWORD_COMMAND": "cat p"}, false, "repositories[0].environment.RESTIC_PASSWORD", "password_file"},
		{"lowercase key", map[string]string{"lower": "x"}, false, "repositories[0].environment", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig(t)
			c.Repositories[0].Environment = tt.env
			err := Validate(c)
			if tt.valid {
				if err != nil {
					t.Errorf("Validate(environment %v) = %v, want nil", tt.env, err)
				}
				return
			}
			got := validationFields(t, err)
			for _, f := range got {
				if strings.HasPrefix(f.Path, tt.prefix) && strings.Contains(f.Msg, tt.wantMsg) {
					return
				}
			}
			t.Errorf("Validate(environment %v) fields = %+v, want one under %s with message containing %q", tt.env, got, tt.prefix, tt.wantMsg)
		})
	}
}

func TestFieldErrorShape(t *testing.T) {
	fe := reflect.TypeFor[FieldError]()
	wantFields := []struct {
		name string
		typ  reflect.Type
	}{
		{"Path", reflect.TypeFor[string]()},
		{"Msg", reflect.TypeFor[string]()},
		{"Code", reflect.TypeFor[errcode.Code]()},
	}
	for _, w := range wantFields {
		f, ok := fe.FieldByName(w.name)
		if !ok || f.Type != w.typ {
			t.Errorf("FieldError.%s = %v (present %v), want type %v", w.name, f.Type, ok, w.typ)
		}
	}
	f, ok := reflect.TypeFor[ValidationError]().FieldByName("Fields")
	if want := reflect.TypeFor[[]FieldError](); !ok || f.Type != want {
		t.Errorf("ValidationError.Fields = %v (present %v), want type %v", f.Type, ok, want)
	}
}

func TestValidationErrorCodeAndText(t *testing.T) {
	c := validConfig(t)
	c.Timestamps = "gmt"
	c.Catalog.ProbeConcurrency = 0
	err := Validate(c)

	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(Validate(invalid)) = %q, want %q", got, errcode.InvalidConfig)
	}
	fields := validationFields(t, err)
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		parts = append(parts, f.Path+": "+f.Msg)
	}
	msg := err.Error()
	for _, p := range parts {
		if !strings.Contains(msg, p) {
			t.Errorf("Validate(invalid).Error() = %q, want it to contain %q", msg, p)
		}
	}
	if want := strings.Join(parts, "; "); !strings.Contains(msg, want) {
		t.Errorf("Validate(invalid).Error() = %q, want it to contain %q", msg, want)
	}
	if findField(fields, "timestamps") == nil || findField(fields, "catalog.probe_concurrency") == nil {
		t.Errorf("Validate(invalid) fields = %+v, want timestamps and catalog.probe_concurrency", fields)
	}
}

func TestValidateIncludesTopologyAndCredentials(t *testing.T) {
	c := validConfig(t)
	c.HistoryMount = c.Roots[0].LocalPath
	if err := os.Chmod(c.Repositories[0].PasswordFile, 0o644); err != nil {
		t.Fatalf("chmod password file: %v", err)
	}

	got := validationFields(t, Validate(c))
	for _, path := range []string{"history_mount", "repositories[0].password_file"} {
		if findField(got, path) == nil {
			t.Errorf("Validate(topology and credential errors) fields = %+v, want one at %s", got, path)
		}
	}
}
