package web

import (
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
)

// formConfig is a config whose every field shown on the Config and Setup
// forms holds a distinct value.
func formConfig(restic string) *config.Config {
	return &config.Config{
		Version:   1,
		Catalog:   config.Catalog{RefreshInterval: 7 * time.Minute},
		Discovery: config.Discovery{Mode: "onaccess"},
		Repositories: []config.Repository{{
			ID:           "main",
			Repository:   "/srv/restic-repo",
			ResticBinary: restic,
			PasswordFile: "/etc/snapback/pass",
			CacheDir:     "/var/cache/snapback-test",
		}},
		Roots: []config.Root{{
			ID:                   "home",
			LocalPath:            "/home/demo",
			RepositoryID:         "main",
			Snapshots:            config.SnapshotFilter{Hostname: "demo-host", TagsAll: []string{"nightly"}},
			SeedPaths:            []config.SeedPath{{Path: "/home/demo/projects"}},
			ExcludeRelativePaths: []string{"tmp-excluded"},
		}},
	}
}

func TestConfigPageShowsConfigOnDisk(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/config", nil, http.Header{"Cookie": {cookie.String()}})
	body := html.UnescapeString(w.Body.String())
	for _, want := range []string{
		`name="revision" value="r1"`,
		"/home/demo\n",
		"host=demo-host\n",
		"tag=nightly\n",
		"tmp-excluded\n",
		"/home/demo/projects\n",
		`value="onaccess" selected`,
		`name="cache_dir" value="/var/cache/snapback-test"`,
		`name="refresh_interval" value="7m0s"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /config: body does not contain %q", want)
		}
	}
}

func TestSetupPageOffersDetectedResticBinaries(t *testing.T) {
	bin := t.TempDir()
	onPath := filepath.Join(bin, "restic")
	if err := os.WriteFile(onPath, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/setup", nil, http.Header{"Cookie": {cookie.String()}})
	body := html.UnescapeString(w.Body.String())
	for _, want := range []string{
		`<option value="/opt/restic/bin/restic">`,
		`<option value="` + onPath + `">`,
		`name="repo_uri" value="/srv/restic-repo"`,
		`name="credential_file" value="/etc/snapback/pass"`,
		"/home/demo\n",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET /setup: body does not contain %q", want)
		}
	}
}
