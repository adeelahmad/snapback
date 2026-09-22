package webui

import (
	"regexp"
	"strings"
	"testing"
)

var (
	controlTagRE = regexp.MustCompile(`(?is)<(input|select|textarea)\b[^>]*>`)
	inputTagRE   = regexp.MustCompile(`(?is)<input\b[^>]*>`)
	optionRE     = regexp.MustCompile(`(?is)<option\b[^>]*>[^<]*</option>`)
	idAttrRE     = regexp.MustCompile(`(?i)\bid="([^"]+)"`)
	labelTagRE   = regexp.MustCompile(`(?is)<label\b[^>]*>`)
	forAttrRE    = regexp.MustCompile(`(?i)\bfor="([^"]*)"`)
	repoInputRE  = regexp.MustCompile(`(?is)<input\b[^>]*value="sftp:host:/repo"`)
	seededOptRE  = regexp.MustCompile(`(?is)<option\b[^>]*value="seeded"[^>]*>`)
)

// setupFixture returns a Setup view with two restic paths and a repo URI.
func setupFixture() SetupView {
	return SetupView{
		Chrome:      Chrome{Title: "Setup", Active: "setup", CSRFToken: "tok123"},
		ResticPaths: []string{"/usr/bin/restic", "/opt/restic/bin/restic"},
		RclonePaths: []string{"/usr/bin/rclone"},
		RepoURI:     "sftp:host:/repo",
		Roots:       []string{"/home/a"},
	}
}

// configFixture returns a Configuration view with two roots in seeded mode.
func configFixture() ConfigView {
	return ConfigView{
		Chrome:          Chrome{Title: "Configuration", Active: "config", CSRFToken: "tok123"},
		Revision:        "rev-7",
		Roots:           []string{"/home/a", "/srv"},
		Filters:         []string{"*.go"},
		Exclusions:      []string{"node_modules"},
		SeedPaths:       []string{"/home/a/src"},
		DiscoveryMode:   "seeded",
		CacheDir:        "/var/cache/snapback",
		RefreshInterval: "15m",
	}
}

// hasHiddenInput reports whether out has a hidden input named name with value.
func hasHiddenInput(out, name, value string) bool {
	for _, tag := range inputTagRE.FindAllString(out, -1) {
		if strings.Contains(tag, `type="hidden"`) &&
			strings.Contains(tag, `name="`+name+`"`) &&
			strings.Contains(tag, `value="`+value+`"`) {
			return true
		}
	}
	return false
}

// hasOption reports whether out has an <option> whose value or text is v.
func hasOption(out, v string) bool {
	for _, opt := range optionRE.FindAllString(out, -1) {
		if strings.Contains(opt, `value="`+v+`"`) || strings.Contains(opt, ">"+v+"<") {
			return true
		}
	}
	return false
}

// elementWith returns the first element whose start tag carries attr, through
// its closing tag, or "" if there is none.
func elementWith(out, attr string) string {
	i := strings.Index(out, attr)
	if i < 0 {
		return ""
	}
	start := strings.LastIndex(out[:i], "<")
	if start < 0 {
		return ""
	}
	fields := strings.Fields(out[start+1 : i])
	if len(fields) == 0 {
		return ""
	}
	end := strings.Index(out[i:], "</"+fields[0]+">")
	if end < 0 {
		return ""
	}
	return out[start : i+end]
}

// labelledBy reports whether out has a <label> whose for attribute is id.
// The attribute may sit anywhere in the start tag: the shared control
// partials write class before for.
func labelledBy(out, id string) bool {
	for _, tag := range labelTagRE.FindAllString(out, -1) {
		if m := forAttrRE.FindStringSubmatch(tag); m != nil && m[1] == id {
			return true
		}
	}
	return false
}

// checkLabelled fails t unless out has at least minControls visible form
// controls and each has an id with a matching <label for>.
func checkLabelled(t *testing.T, page, out string, minControls int) {
	t.Helper()
	var controls []string
	for _, tag := range controlTagRE.FindAllString(out, -1) {
		if strings.Contains(tag, `type="hidden"`) {
			continue
		}
		controls = append(controls, tag)
	}
	if len(controls) < minControls {
		t.Fatalf("render(%q) has %d visible form controls, want at least %d", page, len(controls), minControls)
	}
	for _, c := range controls {
		m := idAttrRE.FindStringSubmatch(c)
		if m == nil {
			t.Errorf("render(%q) control %s has no id, want one", page, c)
			continue
		}
		if !labelledBy(out, m[1]) {
			t.Errorf("render(%q) has no <label for=%q>, want one for %s", page, m[1], c)
		}
	}
}

func TestSetupFormPostsWithCSRF(t *testing.T) {
	p := mustLoad(t, "")
	out := render(t, p, "setup", setupFixture())

	if got := strings.Count(out, "<form"); got != 1 {
		t.Errorf("render(setup) has %d <form> elements, want 1", got)
	}
	if !strings.Contains(out, `method="post"`) {
		t.Errorf(`render(setup) lacks method="post", want a post-back form`)
	}
	if !hasHiddenInput(out, "csrf_token", "tok123") {
		t.Errorf("render(setup) lacks hidden csrf_token = tok123, want one")
	}
	for _, path := range []string{"/usr/bin/restic", "/opt/restic/bin/restic"} {
		if !hasOption(out, path) {
			t.Errorf("render(setup) lacks <option> for restic path %q, want one", path)
		}
	}
	if !repoInputRE.MatchString(out) {
		t.Errorf("render(setup) lacks an input with value %q, want the repo URI", "sftp:host:/repo")
	}
}

func TestSetupControlsLabelled(t *testing.T) {
	p := mustLoad(t, "")
	out := render(t, p, "setup", setupFixture())
	checkLabelled(t, "setup", out, 4)
}

func TestConfigFormRevisionAndCSRF(t *testing.T) {
	p := mustLoad(t, "")
	out := render(t, p, "config", configFixture())

	if !strings.Contains(out, `method="post"`) {
		t.Errorf(`render(config) lacks method="post", want a post-back form`)
	}
	if !hasHiddenInput(out, "revision", "rev-7") {
		t.Errorf("render(config) lacks hidden revision = rev-7, want one")
	}
	if !hasHiddenInput(out, "csrf_token", "tok123") {
		t.Errorf("render(config) lacks hidden csrf_token = tok123, want one")
	}
	for _, root := range []string{"/home/a", "/srv"} {
		if !strings.Contains(out, root) {
			t.Errorf("render(config) lacks root %q, want it present", root)
		}
	}
	seeded := seededOptRE.FindString(out)
	if seeded == "" {
		t.Fatalf(`render(config) has no <option value="seeded">, want one`)
	}
	if !strings.Contains(seeded, "selected") {
		t.Errorf("render(config) option %s is not selected, want selected", seeded)
	}
}

func TestConfigControlsLabelled(t *testing.T) {
	p := mustLoad(t, "")
	out := render(t, p, "config", configFixture())
	checkLabelled(t, "config", out, 6)
}

func TestConfigErrorsAnnounced(t *testing.T) {
	p := mustLoad(t, "")
	v := configFixture()
	v.Errors = []string{"root /x does not exist", hostile}
	out := render(t, p, "config", v)

	alert := elementWith(out, `role="alert"`)
	if alert == "" {
		t.Fatalf(`render(config) with errors has no role="alert" element, want one`)
	}
	if !strings.Contains(alert, "root /x does not exist") {
		t.Errorf("role=alert element = %q, want it to contain %q", alert, "root /x does not exist")
	}
	if !strings.Contains(alert, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("role=alert element = %q, want the escaped hostile error", alert)
	}
	if strings.Contains(out, "<script>alert(1)") {
		t.Errorf("render(config) contains raw <script>alert(1), want it escaped")
	}
}

func TestConfigSavedNotice(t *testing.T) {
	p := mustLoad(t, "")

	saved := configFixture()
	saved.Saved = true
	out := render(t, p, "config", saved)
	notice := elementWith(out, `role="status"`)
	if notice == "" {
		t.Fatalf(`render(config) with Saved = true has no role="status" element, want one`)
	}
	if !strings.Contains(notice, "Saved") {
		t.Errorf("role=status element = %q, want it to contain %q", notice, "Saved")
	}

	out = render(t, p, "config", configFixture())
	if strings.Contains(out, `role="status"`) {
		t.Errorf(`render(config) with Saved = false contains role="status", want it absent`)
	}
}
