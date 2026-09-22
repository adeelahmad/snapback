package web

import (
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/setup"
)

// mountPointPath is the one key path the Setup form exposes the mount point
// under, and mountPointName is how it appears as an input name.
const (
	mountPointPath = "repositories[0].mount_point"
	mountPointName = `name="` + mountPointPath + `"`
)

// mountPointWindow is how far from the control's name attribute a label, a
// help text or an error still counts as adjacent to the field.
const mountPointWindow = 400

// mountPointHome is the home directory every mount point test detects, so the
// platform default is the same on every host.
const mountPointHome = "/home/u"

// mountPointOptions returns Options whose detection seams are fakes rooted at
// mountPointHome, so the Setup page's platform default is pinned.
func mountPointOptions(b *fakeBackend) Options {
	return Options{
		Backend:   b,
		Validator: &fakeValidator{},
		SetupDeps: setup.Deps{
			Getenv: func(k string) string {
				if k == "HOME" {
					return mountPointHome
				}
				return detectEnv[k]
			},
			LookPath: func(name string) (string, error) { return "/fake/bin/" + name, nil },
			Getwd:    func() (string, error) { return mountPointHome, nil },
			Hostname: func() (string, error) { return "laptop", nil },
			Given:    []string{mountPointHome},
		},
	}
}

// mountPointDefault is the platform default mount point for id on this host.
func mountPointDefault(t *testing.T, id string) string {
	t.Helper()
	want, err := setup.DefaultMountPoint(runtime.GOOS, mountPointHome, id)
	if err != nil {
		t.Fatalf("setup.DefaultMountPoint(%q, %q, %q) error = %v", runtime.GOOS, mountPointHome, id, err)
	}
	return want
}

// nearMountPoint returns the bytes of body within mountPointWindow of the
// mount point control, so a test can pin what is adjacent to the field.
func nearMountPoint(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, mountPointName)
	if i < 0 {
		t.Fatalf("body has no control named %s\nbody: %s", mountPointPath, body)
	}
	return body[max(i-mountPointWindow, 0):min(i+mountPointWindow, len(body))]
}

// mountPointForm returns a complete Setup form whose mount point is mount.
func mountPointForm(csrf, mount string) url.Values {
	return url.Values{
		csrfField:         {csrf},
		"restic_path":     {"/usr/local/bin/restic"},
		"rclone_path":     {"/usr/local/bin/rclone"},
		"repo_uri":        {"/srv/edited-repo"},
		"credential_file": {"/etc/snapback/edited"},
		"roots":           {"/home/edited\n"},
		mountPointPath:    {mount},
	}
}

// TestSetupFormRendersOneMountPointField pins that a fresh Setup page offers
// exactly one mount point control, pre-filled with the platform default for
// the repository id, labelled as the mount point and explaining /mnt/<id>.
func TestSetupFormRendersOneMountPointField(t *testing.T) {
	cfg := formConfig("/opt/restic/bin/restic")
	cfg.Repositories[0].MountPoint = ""
	srv, cookie, _ := newTestServer(t, mountPointOptions(&fakeBackend{cfg: cfg, rev: "r1"}))

	w := do(t, srv, http.MethodGet, "/setup", nil, http.Header{"Cookie": {cookie.String()}})

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /setup: status = %d, want %d", got, want)
	}
	body := w.Body.String()
	if got, want := strings.Count(body, mountPointName), 1; got != want {
		t.Fatalf("GET /setup: controls named %s = %d, want %d", mountPointPath, got, want)
	}
	near := nearMountPoint(t, body)
	if want := `value="` + mountPointDefault(t, "main") + `"`; !strings.Contains(near, want) {
		t.Errorf("GET /setup: mount point control does not carry %s\nnear the field: %s", want, near)
	}
	if want := "Mount point"; !strings.Contains(near, want) {
		t.Errorf("GET /setup: mount point control is not labelled %q\nnear the field: %s", want, near)
	}
	if want := "/mnt/"; !strings.Contains(near, want) {
		t.Errorf("GET /setup: mount point help does not name %q\nnear the field: %s", want, near)
	}
}

// TestSetupFormSavesMountPoint pins that the rendered control round-trips
// through POST /setup into repositories[0].mount_point.
func TestSetupFormSavesMountPoint(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, mountPointOptions(b))

	form := mountPointForm(csrf, "/srv/restore")
	w := do(t, srv, http.MethodPost, "/setup", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /setup: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /setup: saved configs = %d, want %d", got, want)
	}
	if got, want := b.saved[0].Repositories[0].MountPoint, "/srv/restore"; got != want {
		t.Errorf("POST /setup: repositories[0].mount_point = %q, want %q", got, want)
	}
}

// TestSetupFormSavesEmptyMountPoint pins that submitting an empty value
// disables the mount point instead of re-applying the platform default.
func TestSetupFormSavesEmptyMountPoint(t *testing.T) {
	cfg := formConfig("/opt/restic/bin/restic")
	cfg.Repositories[0].MountPoint = "/srv/restore"
	b := &fakeBackend{cfg: cfg, rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, mountPointOptions(b))

	form := mountPointForm(csrf, "")
	w := do(t, srv, http.MethodPost, "/setup", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /setup: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	if got, want := len(b.saved), 1; got != want {
		t.Fatalf("POST /setup: saved configs = %d, want %d", got, want)
	}
	if got, want := b.saved[0].Repositories[0].MountPoint, ""; got != want {
		t.Errorf("POST /setup: repositories[0].mount_point = %q, want %q", got, want)
	}
}

// TestSetupFormRejectsRelativeMountPoint pins that a relative mount point is
// refused by the same config validator, saves nothing, and re-renders with the
// error next to the field it belongs to.
func TestSetupFormRejectsRelativeMountPoint(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1", newRev: "r2"}
	srv, cookie, csrf := newTestServer(t, mountPointOptions(b))

	form := mountPointForm(csrf, "relative")
	w := do(t, srv, http.MethodPost, "/setup", strings.NewReader(form.Encode()), formHeader(cookie))

	if got := w.Code; got < 400 || got > 499 {
		t.Fatalf("POST /setup with a relative mount point: status = %d, want 4xx (body %q)", got, w.Body.String())
	}
	if got, want := len(b.saved), 0; got != want {
		t.Errorf("POST /setup with a relative mount point: saved configs = %d, want %d", got, want)
	}
	near := nearMountPoint(t, w.Body.String())
	if want := "absolute"; !strings.Contains(near, want) {
		t.Errorf("POST /setup with a relative mount point: no error naming %q within %d bytes of the field\nnear the field: %s",
			want, mountPointWindow, near)
	}
	if want := "relative"; !strings.Contains(near, `value="`+want+`"`) {
		t.Errorf("POST /setup with a relative mount point: the field does not keep the submitted value %q\nnear the field: %s",
			want, near)
	}
}

// TestMountPointFieldPrefillsPlatformDefault pins the control the Setup page
// renders: the one key path, the label, the help naming /mnt/<id>, and the
// platform default only when the configuration carries no mount point.
func TestMountPointFieldPrefillsPlatformDefault(t *testing.T) {
	cfg := formConfig("/opt/restic/bin/restic")
	cfg.Repositories[0].MountPoint = ""

	got := mountPointField(cfg, "linux", mountPointHome)

	if want := mountPointPath; got.Path != want {
		t.Errorf("mountPointField().Path = %q, want %q", got.Path, want)
	}
	if want := "Mount point"; got.Label != want {
		t.Errorf("mountPointField().Label = %q, want %q", got.Label, want)
	}
	if want := "/mnt/main"; got.Value != want {
		t.Errorf("mountPointField().Value = %q, want %q", got.Value, want)
	}
	if want := "/mnt/"; !strings.Contains(got.Help, want) {
		t.Errorf("mountPointField().Help = %q, want it to name %q", got.Help, want)
	}
	if want := "restore"; !strings.Contains(strings.ToLower(got.Help), want) {
		t.Errorf("mountPointField().Help = %q, want it to explain %q", got.Help, want)
	}

	cfg.Repositories[0].MountPoint = "/srv/restore"
	if got, want := mountPointField(cfg, "linux", mountPointHome).Value, "/srv/restore"; got != want {
		t.Errorf("mountPointField().Value with a configured mount point = %q, want %q", got, want)
	}
}
