package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/web"
	"github.com/adeelahmad/snapback/internal/webui"
)

// webUIConfigHeading is the heading whose body has to describe the
// Configuration page: every section the form renders, the basic/advanced
// split and where a typed password is stored.
const webUIConfigHeading = "## Configuration"

// webUIImageDir is the directory every image the page shows has to live in.
const webUIImageDir = "docs-site/img"

// webUIImage matches a markdown image reference, capturing its target.
var webUIImage = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)

// configSectionTitles is the list of configurable sections the Configuration
// form actually renders, derived from the code rather than copied, so a new
// top-level configuration key fails this test until the page names it.
func configSectionTitles() []string {
	sections := webui.Sections(web.Fields(&config.Config{}))
	titles := make([]string, 0, len(sections))
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	return titles
}

// namesSection reports whether doc names title the way the page has to: in
// bold, or as the first cell of a table row. A bare mention in prose does not
// count, so "Web UI" cannot stand in for the Web section.
func namesSection(doc, title string) bool {
	q := regexp.QuoteMeta(title)
	return regexp.MustCompile(`(?m)\*\*` + q + `\*\*|^\|\s*` + q + `\s*\|`).MatchString(doc)
}

func TestWebUIDocsNamesEveryConfigurableSection(t *testing.T) {
	sec := section(webUIDoc(t), webUIConfigHeading)

	for _, title := range configSectionTitles() {
		if !namesSection(sec, title) {
			t.Errorf("%s %q section does not name the configurable section %q in bold or as a "+
				"table row; webui.Sections(web.Fields(cfg)) renders it", webUIPage,
				webUIConfigHeading, title)
		}
	}
}

func TestWebUIDocsStatesTheBasicAdvancedSplit(t *testing.T) {
	sec := section(webUIDoc(t), webUIConfigHeading)

	if !lineWithAll(sec, "basic") {
		t.Errorf("%s %q says nothing about the basic block that is open by default",
			webUIPage, webUIConfigHeading)
	}
	if !lineWithAll(sec, "advanced", "count") {
		t.Errorf("%s %q does not state that the advanced block is collapsed behind a count "+
			"badge", webUIPage, webUIConfigHeading)
	}
}

func TestWebUIDocsStatesTheTypedPasswordStoragePath(t *testing.T) {
	doc := webUIDoc(t)

	if !strings.Contains(doc, "credentials/<id>.pass") {
		t.Errorf("%s does not state the path a typed repository password is written to, want the "+
			"pattern %q", webUIPage, "credentials/<id>.pass")
	}
	if !lineWithAll(doc, "password_file", "YAML") {
		t.Errorf("%s does not state that the YAML holds only %q and never the password itself",
			webUIPage, "password_file")
	}
}

func TestWebUIDocsNamesTheInstancesPage(t *testing.T) {
	doc := webUIDoc(t)

	if !strings.Contains(doc, "/instances") {
		t.Errorf("%s does not name the %q page, which shows one card per repository",
			webUIPage, "/instances")
	}
	if !lineWithAll(doc, "schema v2") {
		t.Errorf("%s does not carry the honesty note that named instances need config schema v2",
			webUIPage)
	}
}

func TestWebUIDocsStatesTheFirstRunTour(t *testing.T) {
	doc := webUIDoc(t)

	if !lineWithAll(doc, "tour", "first run") {
		t.Errorf("%s does not state that the Setup page shows a tour on first run", webUIPage)
	}
}

func TestWebUIDocsStatesTheDaemonControl(t *testing.T) {
	doc := webUIDoc(t)

	if !lineWithAll(doc, "daemon", "start", "stop") {
		t.Errorf("%s does not state that the Status page can start and stop the daemon",
			webUIPage)
	}
	if !strings.Contains(doc, "--with-daemon") {
		t.Errorf("%s does not name %q, the flag that runs a daemon for the lifetime of the "+
			"command", webUIPage, "--with-daemon")
	}
}

func TestWebUIDocsShowsNoMissingImages(t *testing.T) {
	root := repoRoot(t)

	for _, m := range webUIImage.FindAllStringSubmatch(webUIDoc(t), -1) {
		target := m[1]
		if strings.Contains(target, "://") {
			t.Errorf("%s shows a remote image %q, want a file under %s",
				webUIPage, target, webUIImageDir)
			continue
		}
		path := filepath.Join(root, filepath.Dir(webUIPage), filepath.FromSlash(target))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s shows image %q, but %s does not exist: %v",
				webUIPage, target, path, err)
		}
	}
}
