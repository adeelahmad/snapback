package web

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
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
