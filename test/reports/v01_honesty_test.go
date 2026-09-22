package reports_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const v01ReportPath = "docs/reports/v0.1-acceptance.md"

var (
	v01BannedPhrases = []string{"production-ready", "cross-platform", "finder integrated", "finder-integrated", "multi-backend"}
	v01OtherBackend  = regexp.MustCompile(`(?i)\b(zfs|btrfs|nilfs2?|borg|borgbackup|kopia|duplicity|duplicati|tarsnap|bup|snapper|rsnapshot)\b`)
	v01SentenceEnd   = regexp.MustCompile(`[.!?](\s|$)|\n\s*\n`)
	helpCommand      = regexp.MustCompile(`(?m)^  ([a-z][a-z-]*)  `)
	docCommand       = regexp.MustCompile("`snapback ([a-z][a-z-]*)")
	scriptCommand    = regexp.MustCompile(`snapback ([a-z][a-z-]*)`)
	reloadLatency    = regexp.MustCompile(`(?i)up to (about|around|roughly) (a|one) minute`)
	notWord          = regexp.MustCompile(`\bnot\b`)
	// v01ConfigKeyInline and v01ConfigKeyYAML match the views config keys, which
	// name the optional daily.N/weekly.N view, not another backend.
	v01ConfigKeyInline = regexp.MustCompile("`views\\.rsnapshot(_keep(\\.(hourly|daily|weekly|monthly))?)?`")
	v01ConfigKeyYAML   = regexp.MustCompile(`^\s*(views\.)?rsnapshot(_keep)?:`)
)

// scriptProse are words that follow "snapback " in install.sh prose rather than naming a command.
var scriptProse = map[string]bool{"to": true, "is": true, "will": true, "binary": true, "from": true, "into": true, "in": true}

// readRepoText returns rel's contents, failing the test if it is missing.
func readRepoText(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

// docsSiteFiles returns every markdown file under docs-site, relative to the repo root.
func docsSiteFiles(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	var files []string
	err := filepath.WalkDir(filepath.Join(root, "docs-site"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs-site: %v", err)
	}
	return files
}

// snapbackHelp builds cmd/snapback and returns its --help output.
func snapbackHelp(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "snapback")
	build := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	build.Dir = repoRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	// --help may exit non-zero; the text is what matters.
	out, _ := exec.Command(bin, "--help").CombinedOutput()
	if strings.TrimSpace(string(out)) == "" {
		t.Fatalf("snapback --help printed nothing")
	}
	return string(out)
}

// unplannedBackendLines returns, keyed by 1-based line number, each line of
// text that names another backend, except lines inside a `## Roadmap` section
// that say "planned". Any "## " heading ends the section. Exact views config
// keys are exempt: in inline code, and as YAML keys inside fenced code blocks. It mirrors
// test/docs/honesty_test.go::TestOnlyResticBackendNamed; the two _test
// packages cannot share code, so the rule is duplicated here.
func unplannedBackendLines(text string) map[int]string {
	hits := map[int]string{}
	inRoadmap, inFence := false, false
	for i, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "## ") {
			inRoadmap = strings.TrimSpace(line) == "## Roadmap"
		}
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
		}
		scan := v01ConfigKeyInline.ReplaceAllString(line, "")
		if inFence {
			scan = v01ConfigKeyYAML.ReplaceAllString(scan, "")
		}
		if !v01OtherBackend.MatchString(scan) {
			continue
		}
		if inRoadmap && strings.Contains(strings.ToLower(line), "planned") {
			continue
		}
		hits[i+1] = strings.TrimSpace(line)
	}
	return hits
}

func TestUnplannedBackendLinesExemptsOnlyConfigKeys(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"```yaml\nviews:\n  rsnapshot: false\n  rsnapshot_keep: {daily: 7}\n```", 0},
		{"| `views.rsnapshot` | boolean |", 0},
		{"| `views.rsnapshot_keep.daily` | integer |", 0},
		{"Set `views.rsnapshot_keep` to trim the list.", 0},
		{"rsnapshot: false", 1},
		{"Snapback also reads rsnapshot trees.", 1},
		{"Use `views.rsnapshot` like rsnapshot does.", 1},
		{"```yaml\n# rsnapshot is supported\n```", 1},
	}
	for _, c := range cases {
		if got := len(unplannedBackendLines(c.in)); got != c.want {
			t.Errorf("len(unplannedBackendLines(%q)) = %d, want %d", c.in, got, c.want)
		}
	}
}

func sentences(text string) []string {
	return v01SentenceEnd.Split(text, -1)
}

func TestPublicDocsMakeNoUnbackedClaims(t *testing.T) {
	sources := map[string]string{}
	for _, rel := range append([]string{"README.md", "install.sh", ".goreleaser.yaml", v01ReportPath}, docsSiteFiles(t)...) {
		sources[rel] = readRepoText(t, rel)
	}
	sources["snapback --help"] = snapbackHelp(t)
	if len(sources) == 0 {
		t.Fatalf("no public docs scanned")
	}

	for name, text := range sources {
		lower := strings.ToLower(text)
		for _, banned := range v01BannedPhrases {
			if strings.Contains(lower, banned) {
				t.Errorf("%s contains %q", name, banned)
			}
		}
		for n, line := range unplannedBackendLines(text) {
			t.Errorf("%s:%d names another backend outside a planned Roadmap line: %s", name, n, line)
		}
		for _, s := range sentences(lower) {
			if strings.Contains(s, "static") && !strings.Contains(s, "linux") {
				t.Errorf("%s: %q says static without linux", name, strings.TrimSpace(s))
			}
			i := strings.Index(s, "macos")
			if i < 0 || !strings.Contains(s, "supported") {
				continue
			}
			rest := s[i:]
			if !strings.Contains(rest, "follow-up") && !notWord.MatchString(rest) {
				t.Errorf("%s: %q calls macOS supported", name, strings.TrimSpace(s))
			}
		}
	}
}

func TestDocsCommandsExist(t *testing.T) {
	have := map[string]bool{}
	for _, m := range helpCommand.FindAllStringSubmatch(snapbackHelp(t), -1) {
		have[m[1]] = true
	}
	if len(have) == 0 {
		t.Fatalf("snapback --help lists no commands")
	}

	named := map[string]string{}
	for _, rel := range append([]string{"README.md"}, docsSiteFiles(t)...) {
		for _, m := range docCommand.FindAllStringSubmatch(readRepoText(t, rel), -1) {
			named[m[1]] = rel
		}
	}
	for _, m := range scriptCommand.FindAllStringSubmatch(readRepoText(t, "install.sh"), -1) {
		if !scriptProse[m[1]] {
			named[m[1]] = "install.sh"
		}
	}
	if len(named) == 0 {
		t.Fatalf("no snapback commands named in README, docs-site or install.sh")
	}

	cmds := make([]string, 0, len(named))
	for c := range named {
		cmds = append(cmds, c)
	}
	sort.Strings(cmds)
	for _, c := range cmds {
		if !have[c] {
			t.Errorf("%s names `snapback %s`, which snapback --help does not list", named[c], c)
		}
	}
}

func TestDocsStateSnapshotReloadLatency(t *testing.T) {
	var found []string
	for _, rel := range append([]string{"README.md"}, docsSiteFiles(t)...) {
		for _, s := range sentences(readRepoText(t, rel)) {
			if reloadLatency.MatchString(s) && strings.Contains(s, ".snapshot") {
				found = append(found, rel)
			}
		}
	}
	if len(found) == 0 {
		t.Errorf("README.md and docs-site do not say a new snapshot may take up to about a minute to appear under .snapshot")
	}
}
