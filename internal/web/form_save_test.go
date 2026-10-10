package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// multiRepoConfig is a config with two repositories and one existing root
// assigned to the first, used to test that new roots are never silently
// assigned to repositories[0].
func multiRepoConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Repositories: []config.Repository{
			{ID: "main", Repository: "/srv/main-repo"},
			{ID: "backup", Repository: "/srv/backup-repo"},
		},
		Roots: []config.Root{{
			ID:           "home",
			LocalPath:    "/home/demo",
			RepositoryID: "main",
		}},
	}
}

func TestConfigFormNewRootWithSingleRepositoryDefaultsToIt(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/edited\n/srv/new-root\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /config: saved configs = %d, want %d", got, want)
	}
	roots := b.saved[0].Roots
	if got, want := len(roots), 2; got != want {
		t.Fatalf("POST /config: saved roots = %d, want %d", got, want)
	}
	newRoot := roots[1]
	if got, want := newRoot.LocalPath, "/srv/new-root"; got != want {
		t.Errorf("POST /config: new root local path = %q, want %q", got, want)
	}
	if got, want := newRoot.RepositoryID, "main"; got != want {
		t.Errorf("POST /config: new root repository id = %q, want %q", got, want)
	}
}

func TestConfigFormNewRootCanChooseNonFirstRepository(t *testing.T) {
	b := &fakeBackend{cfg: multiRepoConfig(), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/demo\n/srv/new-root=backup\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /config: saved configs = %d, want %d", got, want)
	}
	roots := b.saved[0].Roots
	if got, want := len(roots), 2; got != want {
		t.Fatalf("POST /config: saved roots = %d, want %d", got, want)
	}
	newRoot := roots[1]
	if got, want := newRoot.LocalPath, "/srv/new-root"; got != want {
		t.Errorf("POST /config: new root local path = %q, want %q", got, want)
	}
	if got, want := newRoot.RepositoryID, "backup"; got != want {
		t.Errorf("POST /config: new root repository id = %q, want %q (must not default to repositories[0])", got, want)
	}
}

func TestConfigFormNewRootWithUnknownRepositoryIsRejected(t *testing.T) {
	b := &fakeBackend{cfg: multiRepoConfig(), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/demo\n/srv/new-root=nosuchrepo\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /config with unknown repository: status = %d, want %d", got, want)
	}
	if !strings.Contains(w.Body.String(), "nosuchrepo") {
		t.Errorf("POST /config with unknown repository: body does not name the bad repository: %q", w.Body.String())
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /config with unknown repository: saved configs = %d, want %d", got, want)
	}
}

func TestConfigFormNewRootWithoutRepositoryIsAmbiguousWhenMultipleConfigured(t *testing.T) {
	b := &fakeBackend{cfg: multiRepoConfig(), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/demo\n/srv/new-root\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /config without repository: status = %d, want %d", got, want)
	}
	if !strings.Contains(w.Body.String(), "/srv/new-root") {
		t.Errorf("POST /config without repository: body does not name the ambiguous root: %q", w.Body.String())
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /config without repository: saved configs = %d, want %d", got, want)
	}
}

func TestConfigFormNewRootPathContainingEqualsIsNotMisreadAsRepository(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/edited\n/mnt/weird=path/data\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /config: saved configs = %d, want %d", got, want)
	}
	roots := b.saved[0].Roots
	if got, want := len(roots), 2; got != want {
		t.Fatalf("POST /config: saved roots = %d, want %d", got, want)
	}
	if got, want := roots[1].LocalPath, "/mnt/weird=path/data"; got != want {
		t.Errorf("POST /config: new root local path = %q, want %q", got, want)
	}
	if got, want := roots[1].RepositoryID, "main"; got != want {
		t.Errorf("POST /config: new root repository id = %q, want %q", got, want)
	}
}

func TestConfigFormNewRootWithEmptyPathIsRejected(t *testing.T) {
	b := &fakeBackend{cfg: multiRepoConfig(), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/demo\n=backup\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("POST /config with empty path: status = %d, want %d", got, want)
	}
	if !strings.Contains(w.Body.String(), "missing local path") {
		t.Errorf("POST /config with empty path: body does not explain the error: %q", w.Body.String())
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /config with empty path: saved configs = %d, want %d", got, want)
	}
}

func TestSplitRootLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantPath   string
		wantRepoID string
	}{
		{"bare path", "/home/demo", "/home/demo", ""},
		{"path with repository id", "/home/demo=main", "/home/demo", "main"},
		{"trailing segment not a valid id is part of the path", "/mnt/weird=path/data", "/mnt/weird=path/data", ""},
		{"empty path before the id", "=main", "", "main"},
		{"multiple equals, only the last is a candidate id", "/mnt/a=b=main", "/mnt/a=b", "main"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotRepoID := splitRootLine(tt.line)
			if gotPath != tt.wantPath || gotRepoID != tt.wantRepoID {
				t.Errorf("splitRootLine(%q) = (%q, %q), want (%q, %q)", tt.line, gotPath, gotRepoID, tt.wantPath, tt.wantRepoID)
			}
		})
	}
}

// TestConfigFormExistingRootLineIgnoresRepositorySuffix pins that an
// existing root's line is always its literal local path: the
// "=<repository id>" syntax only applies to brand-new root lines, so typing
// it onto an existing root's line neither reassigns the repository nor
// corrupts the path.
func TestConfigFormExistingRootLineIgnoresRepositorySuffix(t *testing.T) {
	b := &fakeBackend{cfg: multiRepoConfig(), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, Options{Backend: b})

	form := configForm(csrf)
	form.Set("roots", "/home/demo=backup\n")
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /config: saved configs = %d, want %d", got, want)
	}
	roots := b.saved[0].Roots
	if got, want := len(roots), 1; got != want {
		t.Fatalf("POST /config: saved roots = %d, want %d", got, want)
	}
	if got, want := roots[0].LocalPath, "/home/demo=backup"; got != want {
		t.Errorf("POST /config: existing root local path = %q, want %q (literal, not split)", got, want)
	}
	if got, want := roots[0].RepositoryID, "main"; got != want {
		t.Errorf("POST /config: existing root repository id = %q, want %q (must not reassign)", got, want)
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
