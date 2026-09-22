package web

import (
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

// formEncode renders fields back into the form values a browser would post.
func formEncode(fields []webui.Field) url.Values {
	v := url.Values{}
	for _, f := range fields {
		switch f.Kind {
		case webui.KindRow:
		case webui.KindChips:
			for _, s := range f.Values {
				v.Add(f.Path, s)
			}
		default:
			v.Set(f.Path, f.Value)
		}
	}
	return v
}

// formFull returns a configuration with every key set to a non-default value.
func formFull() *config.Config {
	return &config.Config{
		Version:         1,
		LinkName:        ".snapshot",
		Timestamps:      "local",
		StateDir:        "/var/lib/snapback",
		HistoryMount:    "/var/lib/snapback/mounts/history",
		BackendMountDir: "/var/lib/snapback/mounts/repositories",
		Web:             config.Web{Enabled: true, Listen: "127.0.0.1:8080", OpenBrowser: true, AssetsDir: "/srv/assets"},
		Catalog: config.Catalog{
			RefreshInterval:      90 * time.Second,
			PrewarmSnapshots:     3,
			PrewarmConcurrency:   4,
			ProbeConcurrency:     5,
			PresenceCacheEntries: 1234,
			PresenceCacheTTL:     2 * time.Minute,
			ReaderPolicy:         config.ReaderPolicy{DenyProcesses: []string{"rg", "fd"}, BurstLimit: 7},
		},
		Views: config.Views{
			Rsnapshot:     true,
			RsnapshotKeep: config.RsnapshotKeep{Hourly: 1, Daily: 2, Weekly: 3, Monthly: 4},
		},
		Discovery: config.Discovery{
			Mode:     "seed",
			Shell:    true,
			Seed:     config.SeedSettings{InodeThreshold: 0.9, MaxLinksPerPath: 500000},
			OnAccess: config.OnAccess{AllowProcesses: []string{"bash", "zsh"}, HandlerTimeout: 20 * time.Millisecond},
		},
		Repositories: []config.Repository{{
			ID:           "main",
			Repository:   "/srv/restic",
			ResticBinary: "/usr/bin/restic",
			RcloneBinary: "/usr/bin/rclone",
			PasswordFile: "/etc/snapback/main.pass",
			CacheDir:     "/var/cache/snapback",
			NoCache:      true,
			LockMode:     "none",
			Environment:  map[string]string{"AWS_PROFILE": "backup", "RESTIC_PACK_SIZE": "32"},
		}},
		Roots: []config.Root{{
			ID:                   "home",
			LocalPath:            "/home/a",
			RepositoryID:         "main",
			PrefixMap:            []config.PrefixMapping{{Hostname: "h1", SourcePath: "/home/a", TreePrefix: "home/a"}},
			Snapshots:            config.SnapshotFilter{Hostname: "h1", TagsAll: []string{"daily"}, SourcePathsExact: []string{"/home/a"}},
			SeedPaths:            []config.SeedPath{{Path: "/home/a", MaxDepth: 3}},
			ExcludeRelativePaths: []string{"cache", "tmp"},
			Snap:                 config.SnapSettings{Tags: []string{"manual"}},
		}},
		Service: config.Service{Manager: "systemd", Scope: "system", RunAsUser: "adeel"},
	}
}

func TestDecodeFormValues(t *testing.T) {
	v := url.Values{}
	for k, s := range map[string]string{
		"version":                            "1",
		"repositories[0].id":                 "main",
		"repositories[0].repository":         "/srv/one",
		"repositories[0].lock_mode":          "none",
		"repositories[0].no_cache":           "on",
		"repositories[1].id":                 "cold",
		"repositories[1].repository":         "/srv/two",
		"repositories[1].lock_mode":          "normal",
		"roots[0].id":                        "home",
		"roots[0].local_path":                "/home/a",
		"roots[0].repository_id":             "main",
		"roots[0].prefix_map[0].hostname":    "h1",
		"roots[0].prefix_map[0].source_path": "/home/a",
		"roots[0].prefix_map[0].tree_prefix": "home/a",
		"roots[0].seed_paths[0].path":        "/home/a",
		"roots[0].seed_paths[0].max_depth":   "3",
		"catalog.refresh_interval":           "90s",
		"catalog.reader_policy.burst_limit":  "7",
	} {
		v.Set(k, s)
	}
	v.Add("roots[0].exclude_relative_paths", "cache")
	v.Add("roots[0].exclude_relative_paths", "tmp")
	v.Add("catalog.reader_policy.deny_processes", "rg")
	v.Add("catalog.reader_policy.deny_processes", "fd")

	want := &config.Config{
		Version: 1,
		Catalog: config.Catalog{
			RefreshInterval: 90 * time.Second,
			ReaderPolicy:    config.ReaderPolicy{DenyProcesses: []string{"rg", "fd"}, BurstLimit: 7},
		},
		Repositories: []config.Repository{
			{ID: "main", Repository: "/srv/one", NoCache: true, LockMode: "none"},
			{ID: "cold", Repository: "/srv/two", LockMode: "normal"},
		},
		Roots: []config.Root{{
			ID:                   "home",
			LocalPath:            "/home/a",
			RepositoryID:         "main",
			PrefixMap:            []config.PrefixMapping{{Hostname: "h1", SourcePath: "/home/a", TreePrefix: "home/a"}},
			SeedPaths:            []config.SeedPath{{Path: "/home/a", MaxDepth: 3}},
			ExcludeRelativePaths: []string{"cache", "tmp"},
		}},
	}

	got, errs := Decode(v)
	if len(errs) != 0 {
		t.Fatalf("Decode(fixture) errors = %v, want none", errs)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Decode(fixture) = %+v, want %+v", got, want)
	}
}

func TestFieldsDecodeRoundTrip(t *testing.T) {
	want := formFull()
	got, errs := Decode(formEncode(Fields(want)))
	if len(errs) != 0 {
		t.Fatalf("Decode(Fields(full)) errors = %v, want none", errs)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Decode(Fields(full)) = %+v, want %+v", got, want)
	}
}

func TestDecodeMalformedValues(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"duration", "catalog.refresh_interval", "sixty"},
		{"int", "catalog.reader_policy.burst_limit", "x"},
		{"bool", "web.enabled", "maybe"},
		{"unknown", "catalog.nope", "1"},
		{"environment", "repositories[0].environment", "novalue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := Decode(url.Values{tt.key: []string{tt.val}})
			if len(errs) != 1 || errs[0].Path != tt.key || errs[0].Msg == "" {
				t.Fatalf("Decode(%s=%s) errors = %v, want one error on %s", tt.key, tt.val, errs, tt.key)
			}
		})
	}
}

func TestFieldsCoverEveryConfigKey(t *testing.T) {
	got := map[string]bool{}
	for _, f := range Fields(config.Default()) {
		if f.Kind == webui.KindRow {
			continue
		}
		if f.Label == "" {
			t.Errorf("Fields()[%s].Label = %q, want a label", f.Path, f.Label)
		}
		got[f.Path] = true
	}
	var missing []string
	for _, p := range yamlLeafPaths(reflect.TypeFor[config.Config](), "") {
		if !got[p] {
			missing = append(missing, p)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		t.Errorf("Fields(Default()) misses %d paths: %s", len(missing), strings.Join(missing, ", "))
	}
}

// yamlLeafPaths lists every settable yaml path reachable in t, using index 0
// for slices.
func yamlLeafPaths(t reflect.Type, prefix string) []string {
	var out []string
	for i := range t.NumField() {
		f := t.Field(i)
		tag := yamlName(f.Tag.Get("yaml"))
		if tag == "" {
			continue
		}
		path := tag
		if prefix != "" {
			path = prefix + "." + tag
		}
		ft := f.Type
		if ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Struct {
			out = append(out, yamlLeafPaths(ft.Elem(), path+"[0]")...)
			continue
		}
		if ft.Kind() == reflect.Struct {
			out = append(out, yamlLeafPaths(ft, path)...)
			continue
		}
		out = append(out, path)
	}
	return out
}

func TestFieldKinds(t *testing.T) {
	kinds := map[string]webui.Kind{}
	for _, f := range Fields(config.Default()) {
		kinds[f.Path] = f.Kind
	}
	tests := []struct {
		path string
		want webui.Kind
	}{
		{"link_name", webui.KindText},
		{"version", webui.KindNumber},
		{"web.enabled", webui.KindToggle},
		{"catalog.refresh_interval", webui.KindDuration},
		{"catalog.reader_policy.deny_processes", webui.KindChips},
		{"repositories[0].lock_mode", webui.KindSelect},
		{"discovery.mode", webui.KindSelect},
		{"service.scope", webui.KindSelect},
		{"repositories[0]", webui.KindRow},
		{"roots[0].prefix_map[0]", webui.KindRow},
		{"roots[0].seed_paths[0]", webui.KindRow},
	}
	for _, tt := range tests {
		if got := kinds[tt.path]; got != tt.want {
			t.Errorf("Fields()[%s].Kind = %q, want %q", tt.path, got, tt.want)
		}
	}
	for _, f := range Fields(config.Default()) {
		if f.Path == "repositories[0].lock_mode" && len(f.Options) != 2 {
			t.Errorf("Fields()[repositories[0].lock_mode].Options = %v, want 2 options", f.Options)
		}
	}
}

func TestSectionsGroupFieldsByTopLevelKey(t *testing.T) {
	secs := webui.Sections(Fields(config.Default()))
	var titles []string
	for _, s := range secs {
		titles = append(titles, s.Title)
		if len(s.Fields) == 0 {
			t.Errorf("Sections() section %q has no fields", s.Title)
		}
	}
	joined := strings.Join(titles, ",")
	for _, want := range []string{"Catalog", "Discovery", "Repositories", "Roots", "Service", "Views", "Web"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Sections() titles = %s, want one named %s", joined, want)
		}
	}
	if n := len(secs); n < 7 {
		t.Errorf("len(Sections()) = %d, want at least 7", n)
	}
}
