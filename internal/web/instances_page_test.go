package web

import (
	"fmt"
	"html"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

// instanceNote is the honesty note the instance cards carry verbatim: named
// instances are a schema v2 feature that does not exist yet. It is pinned in
// internal/webui by the card partial and must survive onto the page.
const instanceNote = "Named instances arrive with config schema v2 (SPEC-ADDENDUM-A §2) and are not implemented yet."

// twoInstanceConfig is a configuration with two repositories, each covering
// one root, so "one card per repository" is more than a single-element case.
func twoInstanceConfig() *config.Config {
	return &config.Config{
		Version:   1,
		Discovery: config.Discovery{Mode: "seed"},
		Repositories: []config.Repository{{
			ID:           "main",
			Repository:   "/srv/restic-main",
			PasswordFile: "/etc/snapback/main",
			MountPoint:   "/mnt/snapback-main",
		}, {
			ID:           "offsite",
			Repository:   "sftp:backup@nas:/srv/restic-offsite",
			PasswordFile: "/etc/snapback/offsite",
		}},
		Roots: []config.Root{{
			ID: "home", LocalPath: "/home/demo", RepositoryID: "main",
		}, {
			ID: "work", LocalPath: "/srv/work", RepositoryID: "offsite",
		}},
	}
}

// pageHeader returns the headers of a logged-in HTML page request.
func pageHeader(cookie *http.Cookie) http.Header {
	return http.Header{"Cookie": {cookie.String()}}
}

// TestInstancesPageIsSessionGuarded pins that /instances sits behind the
// session guard like every other page, and that a logged-in request renders.
func TestInstancesPageIsSessionGuarded(t *testing.T) {
	b := &fakeBackend{cfg: twoInstanceConfig(), rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/instances", nil, http.Header{})
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("GET /instances without a session: status = %d, want %d", got, want)
	}

	w = do(t, srv, http.MethodGet, "/instances", nil, pageHeader(cookie))
	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("GET /instances with a session: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
}

// TestInstancesPageRendersOneCardPerRepository pins the page against the card
// model: one <article class="instance" data-instance="<id>"> per configured
// repository, in configuration order, and the honesty note exactly once.
func TestInstancesPageRendersOneCardPerRepository(t *testing.T) {
	cfg := twoInstanceConfig()
	b := &fakeBackend{cfg: cfg, rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/instances", nil, pageHeader(cookie))
	body := w.Body.String()

	var at []int
	for _, repo := range cfg.Repositories {
		want := fmt.Sprintf(`<article class="instance" data-instance=%q>`, repo.ID)
		if got := strings.Count(body, want); got != 1 {
			t.Errorf("GET /instances: card %s appears %d times, want 1\nbody = %q", want, got, body)
			continue
		}
		at = append(at, strings.Index(body, want))
	}
	for i := 1; i < len(at); i++ {
		if at[i-1] >= at[i] {
			t.Errorf("GET /instances: card %q is not before card %q",
				cfg.Repositories[i-1].ID, cfg.Repositories[i].ID)
		}
	}
	if got := strings.Count(html.UnescapeString(body), instanceNote); got != 1 {
		t.Errorf("GET /instances: honesty note %q appears %d times, want 1\nbody = %q", instanceNote, got, body)
	}
}

// TestConfigSectionsPageSplitsBasicAndAdvanced pins that the Config page
// renders the T2 split: one always-visible basic block holding the first-run
// keys, and one collapsed advanced disclosure.
func TestConfigSectionsPageSplitsBasicAndAdvanced(t *testing.T) {
	b := &fakeBackend{cfg: formConfig("/opt/restic/bin/restic"), rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})

	w := do(t, srv, http.MethodGet, "/config", nil, pageHeader(cookie))
	body := w.Body.String()

	if got := strings.Count(body, `data-js="form-basic"`); got != 1 {
		t.Errorf(`GET /config: data-js="form-basic" appears %d times, want 1`+"\nbody = %q", got, body)
	}
	if got := strings.Count(body, `<details data-js="form-advanced"`); got != 1 {
		t.Errorf(`GET /config: <details data-js="form-advanced" appears %d times, want 1`+"\nbody = %q", got, body)
	}
	if strings.Contains(body, `<details data-js="form-advanced" open`) {
		t.Errorf("GET /config: the advanced disclosure is open, want it collapsed\nbody = %q", body)
	}
	basic, _, ok := strings.Cut(body, `<details data-js="form-advanced"`)
	if !ok {
		t.Fatalf("GET /config: no advanced disclosure to split the basic block from\nbody = %q", body)
	}
	if want := `name="repositories[0].repository"`; !strings.Contains(basic, want) {
		t.Errorf("GET /config: the basic block does not carry %s\nbasic = %q", want, basic)
	}
}

// TestConfigSectionsAdvancedBadgeMatchesSplitCount pins that the disclosure
// summary tells the truth: its badge is the number of fields Split hid.
func TestConfigSectionsAdvancedBadgeMatchesSplitCount(t *testing.T) {
	cfg := formConfig("/opt/restic/bin/restic")
	b := &fakeBackend{cfg: cfg, rev: "r1"}
	srv, cookie, _ := newTestServer(t, Options{Backend: b})
	_, advanced := Split(Fields(cfg))

	w := do(t, srv, http.MethodGet, "/config", nil, pageHeader(cookie))
	body := w.Body.String()

	want := fmt.Sprintf(`<summary>Advanced <span class="badge">%d</span></summary>`, advanced.Count)
	if !strings.Contains(body, want) {
		t.Errorf("GET /config: summary badge is not %s\nbody = %q", want, body)
	}
}

// TestConfigSectionsFormSavesCardMountPoint pins the full round trip of a
// field the instance card owns: posting the whole configuration with a
// changed repositories[0].mount_point redirects, writes that mount point, and
// the sectioned page renders it back on its own control.
func TestConfigSectionsFormSavesCardMountPoint(t *testing.T) {
	tree := newFormTree(t)
	srv, cookie, csrf := newTestServer(t, Options{Backend: fileBackend{path: tree.cfgPath}})
	mount := filepath.Join(tree.dir, "instance-mnt")

	form := everySectionForm(csrf, tree)
	form.Set("repositories[0].mount_point", mount)
	w := do(t, srv, http.MethodPost, "/config", strings.NewReader(form.Encode()), formHeader(cookie))

	if got, want := w.Code, http.StatusSeeOther; got != want {
		t.Fatalf("POST /config with a card mount point: status = %d, want %d (body %q)", got, want, w.Body.String())
	}
	saved, _, err := config.Load(tree.cfgPath)
	if err != nil {
		t.Fatalf("config.Load(saved config) error = %v", err)
	}
	if got := saved.Repositories[0].MountPoint; got != mount {
		t.Errorf("saved repositories[0].mount_point = %q, want %q", got, mount)
	}

	body := do(t, srv, http.MethodGet, "/config", nil, pageHeader(cookie)).Body.String()
	want := fmt.Sprintf(`name="repositories[0].mount_point" value=%q`, mount)
	if !strings.Contains(body, want) {
		t.Errorf("GET /config after the save: body does not carry %s\nbody = %q", want, body)
	}
}
