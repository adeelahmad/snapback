package web

import (
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
)

// formHeader returns the headers of a logged-in HTML form post.
func formHeader(cookie *http.Cookie) http.Header {
	h := http.Header{}
	h.Set("Cookie", cookie.String())
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	return h
}

// configForm is a valid Config form, edited from the formConfig fixture.
func configForm(csrf string) url.Values {
	return url.Values{
		csrfField:          {csrf},
		"revision":         {"r1"},
		"roots":            {"/home/edited\n"},
		"filters":          {"host=edited-host\ntag=weekly\n"},
		"exclusions":       {"cache\n"},
		"seed_paths":       {"/home/edited/projects\n"},
		"discovery_mode":   {"seeded"},
		"cache_dir":        {"/var/cache/edited"},
		"refresh_interval": {"11m"},
	}
}

func TestConfigFormSavesAndRedirects(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := w.Header().Get("Location"), "/config?saved=1"; got != want {
		t.Errorf("POST /config: Location = %q, want %q", got, want)
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /config: saved configs = %d, want %d", got, want)
	}
	if got, want := b.gotRevs[0], b.rev; got != want {
		t.Errorf("POST /config: saved revision = %q, want %q", got, want)
	}
	saved := b.saved[0]
	if got, want := len(saved.Roots), 1; got != want {
		t.Fatalf("POST /config: saved roots = %d, want %d", got, want)
	}
	root := saved.Roots[0]
	if got, want := root.LocalPath, "/home/edited"; got != want {
		t.Errorf("POST /config: root local path = %q, want %q", got, want)
	}
	if got, want := root.ID, "home"; got != want {
		t.Errorf("POST /config: root id = %q, want %q", got, want)
	}
	if got, want := root.Snapshots.Hostname, "edited-host"; got != want {
		t.Errorf("POST /config: root hostname = %q, want %q", got, want)
	}
	if got, want := strings.Join(root.Snapshots.TagsAll, ","), "weekly"; got != want {
		t.Errorf("POST /config: root tags = %q, want %q", got, want)
	}
	if got, want := strings.Join(root.ExcludeRelativePaths, ","), "cache"; got != want {
		t.Errorf("POST /config: root exclusions = %q, want %q", got, want)
	}
	if got, want := len(root.SeedPaths), 1; got != want {
		t.Fatalf("POST /config: root seed paths = %d, want %d", got, want)
	}
	if got, want := root.SeedPaths[0].Path, "/home/edited/projects"; got != want {
		t.Errorf("POST /config: root seed path = %q, want %q", got, want)
	}
	if got, want := saved.Discovery.Mode, "seeded"; got != want {
		t.Errorf("POST /config: discovery mode = %q, want %q", got, want)
	}
	if got, want := saved.Catalog.RefreshInterval.String(), "11m0s"; got != want {
		t.Errorf("POST /config: refresh interval = %q, want %q", got, want)
	}
	if got, want := saved.Repositories[0].CacheDir, "/var/cache/edited"; got != want {
		t.Errorf("POST /config: cache dir = %q, want %q", got, want)
	}
}

func TestConfigFormInvalidFieldRerenders(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("refresh_interval", "eleven minutes")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /config: status = %d, want %d", got, want)
	}
	body := w.Body.String()
	if !strings.Contains(body, "refresh interval") {
		t.Errorf("POST /config: body does not name the bad field: %q", body)
	}
	if !strings.Contains(body, "eleven minutes") {
		t.Errorf("POST /config: body does not keep the submitted value: %q", body)
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /config: saved configs = %d, want %d", got, want)
	}
}

func TestConfigFormWithoutCSRFTokenIsForbidden(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Del(csrfField)
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Fatalf("POST /config without csrf: status = %d, want %d", got, want)
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /config without csrf: saved configs = %d, want %d", got, want)
	}
}

func TestSetupFormSavesAndRedirectsToStatus(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b, Validator: &fakeValidator{}})

	form := url.Values{
		csrfField:         {csrf},
		"restic_path":     {"/usr/local/bin/restic"},
		"rclone_path":     {"/usr/local/bin/rclone"},
		"repo_uri":        {"/srv/edited-repo"},
		"credential_file": {"/etc/snapback/edited"},
		"roots":           {"/home/edited\n"},
	}
	w := do(t, srv, http.MethodPost, "/setup", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /setup: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := w.Header().Get("Location"), "/status"; got != want {
		t.Errorf("POST /setup: Location = %q, want %q", got, want)
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /setup: saved configs = %d, want %d", got, want)
	}
	repo := b.saved[0].Repositories[0]
	if got, want := repo.Repository, "/srv/edited-repo"; got != want {
		t.Errorf("POST /setup: repository = %q, want %q", got, want)
	}
	if got, want := repo.ResticBinary, "/usr/local/bin/restic"; got != want {
		t.Errorf("POST /setup: restic binary = %q, want %q", got, want)
	}
	if got, want := repo.RcloneBinary, "/usr/local/bin/rclone"; got != want {
		t.Errorf("POST /setup: rclone binary = %q, want %q", got, want)
	}
	if got, want := repo.PasswordFile, "/etc/snapback/edited"; got != want {
		t.Errorf("POST /setup: password file = %q, want %q", got, want)
	}
	if got, want := b.saved[0].Roots[0].LocalPath, "/home/edited"; got != want {
		t.Errorf("POST /setup: root local path = %q, want %q", got, want)
	}
}

// TestSetupFormFirstRunSavesConfigWithDefaults pins the first-run contract:
// a config assembled by the Setup form on a machine with no config file
// carries the same defaults as a parsed one, so it validates and is written.
func TestSetupFormFirstRunSavesConfigWithDefaults(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "state"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdgconfig"))
	pass := filepath.Join(tmp, "password")
	if err := os.WriteFile(pass, []byte("pw\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(password) error = %v", err)
	}
	repo, root := filepath.Join(tmp, "repo"), filepath.Join(tmp, "home")
	for _, d := range []string{repo, root} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatalf("os.MkdirAll(%q) error = %v", d, err)
		}
	}
	cfgPath := filepath.Join(tmp, "xdgconfig", "snapback", "config.yaml")
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: cfgPath}})

	form := url.Values{
		csrfField:         {csrf},
		"restic_path":     {"/usr/local/bin/restic"},
		"repo_uri":        {repo},
		"credential_file": {pass},
		"roots":           {root + "\n"},
	}
	w := do(t, srv, http.MethodPost, "/setup", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /setup on first run: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load(saved config) error = %v", err)
	}
	if err := config.Validate(saved); err != nil {
		t.Errorf("config.Validate(saved config) = %v, want nil", err)
	}
	if len(saved.Repositories) == 0 {
		t.Fatalf("saved repositories = 0, want 1")
	}
	if got, want := saved.Repositories[0].LockMode, "normal"; got != want {
		t.Errorf("saved repositories[0].lock_mode = %q, want %q", got, want)
	}
}

// formTree is the temporary directory tree an every-section Config form
// points at: real paths, so config.Validate accepts the submitted values.
type formTree struct {
	dir      string
	cfgPath  string
	repo     string
	root     string
	password string
	cache    string
	state    string
	history  string
	backends string
}

// newFormTree creates the directories and the password file the form names.
func newFormTree(t *testing.T) formTree {
	t.Helper()
	dir := t.TempDir()
	tree := formTree{
		dir:      dir,
		cfgPath:  filepath.Join(dir, "config.yaml"),
		repo:     filepath.Join(dir, "repo"),
		root:     filepath.Join(dir, "home"),
		password: filepath.Join(dir, "password"),
		cache:    filepath.Join(dir, "cache"),
		state:    filepath.Join(dir, "state"),
		history:  filepath.Join(dir, "mnt", "history"),
		backends: filepath.Join(dir, "mnt", "backends"),
	}
	for _, d := range []string{tree.repo, tree.root, tree.cache, tree.state, tree.history, tree.backends} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatalf("os.MkdirAll(%q) error = %v", d, err)
		}
	}
	if err := os.WriteFile(tree.password, []byte("pw"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", tree.password, err)
	}
	return tree
}

// everySectionForm is a Config form that carries every configuration section
// under its YAML key path, the naming scheme the form model decodes.
func everySectionForm(csrf string, tree formTree) url.Values {
	return url.Values{
		csrfField:           {csrf},
		"version":           {"1"},
		"link_name":         {".snapshot"},
		"timestamps":        {"utc"},
		"state_dir":         {tree.state},
		"history_mount":     {tree.history},
		"backend_mount_dir": {tree.backends},

		"web.enabled":      {"on"},
		"web.listen":       {"127.0.0.1:7899"},
		"web.open_browser": {"off"},

		"catalog.refresh_interval":              {"11m0s"},
		"catalog.prewarm_snapshots":             {"3"},
		"catalog.prewarm_concurrency":           {"2"},
		"catalog.probe_concurrency":             {"4"},
		"catalog.presence_cache_entries":        {"512"},
		"catalog.presence_cache_ttl":            {"30s"},
		"catalog.reader_policy.deny_processes":  {"mds", "mdworker"},
		"catalog.reader_policy.burst_limit":     {"7"},
		"views.rsnapshot":                       {"on"},
		"views.rsnapshot_keep.hourly":           {"6"},
		"views.rsnapshot_keep.daily":            {"7"},
		"views.rsnapshot_keep.weekly":           {"4"},
		"views.rsnapshot_keep.monthly":          {"3"},
		"discovery.mode":                        {"seed"},
		"discovery.shell":                       {"on"},
		"discovery.seed.inode_threshold":        {"0.5"},
		"discovery.seed.max_links_per_path":     {"9"},
		"discovery.on_access.allow_processes":   {"finder"},
		"discovery.on_access.handler_timeout":   {"5s"},
		"repositories[0].id":                    {"main"},
		"repositories[0].repository":            {tree.repo},
		"repositories[0].restic_binary":         {"/usr/local/bin/restic"},
		"repositories[0].password_file":         {tree.password},
		"repositories[0].cache_dir":             {tree.cache},
		"repositories[0].no_cache":              {"off"},
		"repositories[0].lock_mode":             {"none"},
		"repositories[0].environment":           {"RESTIC_COMPRESSION=max"},
		"roots[0].id":                           {"home"},
		"roots[0].local_path":                   {tree.root},
		"roots[0].repository_id":                {"main"},
		"roots[0].prefix_map[0].hostname":       {"demo-host"},
		"roots[0].prefix_map[0].source_path":    {"/srv/demo"},
		"roots[0].prefix_map[0].tree_prefix":    {"demo"},
		"roots[0].snapshots.hostname":           {"demo-host"},
		"roots[0].snapshots.tags_all":           {"nightly"},
		"roots[0].snapshots.source_paths_exact": {"/srv/demo"},
		"roots[0].seed_paths[0].path":           {filepath.Join(tree.root, "projects")},
		"roots[0].seed_paths[0].max_depth":      {"2"},
		"roots[0].exclude_relative_paths":       {"tmp-excluded"},
		"roots[0].snap.tags":                    {"adhoc"},
		"service.manager":                       {"systemd"},
		"service.scope":                         {"user"},
		"service.run_as_user":                   {"demo"},
		"telemetry.enabled":                     {"off"},
	}
}

// wantEverySection is the configuration everySectionForm must save.
func wantEverySection(tree formTree) *config.Config {
	cfg := &config.Config{
		Version:         1,
		LinkName:        ".snapshot",
		Timestamps:      "utc",
		StateDir:        tree.state,
		HistoryMount:    tree.history,
		BackendMountDir: tree.backends,
		Web:             config.Web{Enabled: true, Listen: "127.0.0.1:7899"},
		Catalog: config.Catalog{
			RefreshInterval:      11 * time.Minute,
			PrewarmSnapshots:     3,
			PrewarmConcurrency:   2,
			ProbeConcurrency:     4,
			PresenceCacheEntries: 512,
			PresenceCacheTTL:     30 * time.Second,
			ReaderPolicy: config.ReaderPolicy{
				DenyProcesses: []string{"mds", "mdworker"},
				BurstLimit:    7,
			},
		},
		Views: config.Views{
			Rsnapshot:     true,
			RsnapshotKeep: config.RsnapshotKeep{Hourly: 6, Daily: 7, Weekly: 4, Monthly: 3},
		},
		Discovery: config.Discovery{
			Mode:     "seed",
			Shell:    true,
			Seed:     config.SeedSettings{InodeThreshold: 0.5, MaxLinksPerPath: 9},
			OnAccess: config.OnAccess{AllowProcesses: []string{"finder"}, HandlerTimeout: 5 * time.Second},
		},
		Repositories: []config.Repository{{
			ID:           "main",
			Repository:   tree.repo,
			ResticBinary: "/usr/local/bin/restic",
			PasswordFile: tree.password,
			CacheDir:     tree.cache,
			LockMode:     "none",
			Environment:  map[string]string{"RESTIC_COMPRESSION": "max"},
		}},
		Roots: []config.Root{{
			ID:                   "home",
			LocalPath:            tree.root,
			RepositoryID:         "main",
			PrefixMap:            []config.PrefixMapping{{Hostname: "demo-host", SourcePath: "/srv/demo", TreePrefix: "demo"}},
			Snapshots:            config.SnapshotFilter{Hostname: "demo-host", TagsAll: []string{"nightly"}, SourcePathsExact: []string{"/srv/demo"}},
			SeedPaths:            []config.SeedPath{{Path: filepath.Join(tree.root, "projects"), MaxDepth: 2}},
			ExcludeRelativePaths: []string{"tmp-excluded"},
			Snap:                 config.SnapSettings{Tags: []string{"adhoc"}},
		}},
		Service: config.Service{Manager: "systemd", Scope: "user", RunAsUser: "demo"},
	}
	config.ApplyDefaults(cfg)
	return cfg
}

// TestConfigFormSavesEverySection pins that posting the whole configuration
// under its YAML key paths saves every section: repositories, roots,
// prefix_map, seed_paths, catalog, views, discovery, service, web and the
// reader policy. It compares the decoded configuration, not the file bytes.
func TestConfigFormSavesEverySection(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	want := wantEverySection(tree)
	for _, c := range []struct {
		section string
		got     any
		want    any
	}{
		{"web", saved.Web, want.Web},
		{"catalog", saved.Catalog, want.Catalog},
		{"catalog.reader_policy", saved.Catalog.ReaderPolicy, want.Catalog.ReaderPolicy},
		{"views", saved.Views, want.Views},
		{"discovery", saved.Discovery, want.Discovery},
		{"service", saved.Service, want.Service},
		{"repositories", saved.Repositories, want.Repositories},
		{"roots", saved.Roots, want.Roots},
	} {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("POST /config: saved %s =\n%+v\nwant\n%+v", c.section, c.got, c.want)
		}
	}
	if !reflect.DeepEqual(saved, want) {
		t.Errorf("POST /config: saved config =\n%+v\nwant\n%+v", saved, want)
	}
}

// TestConfigFormInvalidLockModeAnchorsErrorToField pins that an invalid
// lock_mode is rejected with a 4xx re-render that puts the message on that
// field's control and keeps every other submitted value in the form.
func TestConfigFormInvalidLockModeAnchorsErrorToField(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	form.Set("repositories[0].lock_mode", "sometimes")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got := w.Code; got < 400 || got > 499 {
		t.Fatalf("POST /config with a bad lock_mode: status = %d, want 4xx (body %q)", got, w.Body.String())
	}
	if _, err := os.Stat(tree.cfgPath); err == nil {
		t.Errorf("POST /config with a bad lock_mode: %q was written, want no save", tree.cfgPath)
	}

	body := html.UnescapeString(w.Body.String())
	const (
		control = `name="repositories[0].lock_mode"`
		msg     = "must be normal or none"
		window  = 400
	)
	at := strings.Index(body, control)
	if at < 0 {
		t.Fatalf("POST /config with a bad lock_mode: body has no %s control:\n%s", control, body)
	}
	near := body[max(0, at-window):min(len(body), at+window)]
	if !strings.Contains(near, msg) {
		t.Errorf("POST /config with a bad lock_mode: %q is not within %d bytes of the %s control; got:\n%s",
			msg, window, control, near)
	}
	if !strings.Contains(near, "sometimes") {
		t.Errorf("POST /config with a bad lock_mode: the rejected value is not kept on its control; got:\n%s", near)
	}
	for _, want := range []string{
		`name="roots[0].local_path" value="` + tree.root + `"`,
		`name="roots[0].prefix_map[0].hostname" value="demo-host"`,
		`name="roots[0].seed_paths[0].path" value="` + filepath.Join(tree.root, "projects") + `"`,
		`name="catalog.reader_policy.burst_limit" value="7"`,
		`name="views.rsnapshot_keep.daily" value="7"`,
		`name="web.listen" value="127.0.0.1:7899"`,
		`name="service.run_as_user" value="demo"`,
		`name="repositories[0].repository" value="` + tree.repo + `"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("POST /config with a bad lock_mode: body does not echo %s", want)
		}
	}
}

// TestConfigFormTypedPasswordGoesToCredentialStore pins that a password typed
// into the form is written to the credential store and that only its path,
// never the secret, reaches config.yaml.
func TestConfigFormTypedPasswordGoesToCredentialStore(t *testing.T) {
	const secret = "correct-horse-battery-staple"
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})

	form := everySectionForm(csrf, tree)
	form.Del("repositories[0].password_file")
	form.Set("repositories[0].password_mode", "typed")
	form.Set("repositories[0].password", secret)
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config with a typed password: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	raw, err := os.ReadFile(tree.cfgPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", tree.cfgPath, err)
	}
	if strings.Contains(string(raw), secret) {
		t.Errorf("POST /config with a typed password: config.yaml contains the secret:\n%s", raw)
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", tree.cfgPath, err)
	}
	if len(saved.Repositories) == 0 {
		t.Fatalf("POST /config with a typed password: saved repositories = 0, want 1")
	}
	want := filepath.Join(srv.opts.StateDir, "credentials", "main.pass")
	if got := saved.Repositories[0].PasswordFile; got != want {
		t.Fatalf("POST /config with a typed password: password_file = %q, want %q", got, want)
	}
	stored, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", want, err)
	}
	if got := string(stored); got != secret {
		t.Errorf("POST /config with a typed password: credential file = %q, want %q", got, secret)
	}
	info, err := os.Stat(want)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", want, err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Errorf("POST /config with a typed password: credential file mode = %v, want %v", got, want)
	}
}
