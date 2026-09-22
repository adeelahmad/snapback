package community

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	bugFormPath         = ".github/ISSUE_TEMPLATE/bug.yml"
	featureFormPath     = ".github/ISSUE_TEMPLATE/feature.yml"
	issueConfigPath     = ".github/ISSUE_TEMPLATE/config.yml"
	blankIssuesDisabled = "blank_issues_enabled: false"
	securityAdvisoryURL = "https://github.com/adeelahmad/snapback/security/advisories/new"
	discussionsURL      = "https://github.com/adeelahmad/snapback/discussions"
)

var (
	issueForms       = []string{bugFormPath, featureFormPath}
	retiredTemplates = []string{".github/ISSUE_TEMPLATE/bug_report.md", ".github/ISSUE_TEMPLATE/feature_request.md"}
	topLevelKeyRe    = regexp.MustCompile(`^([a-z_]+):\s*(.*)$`)
	bodyItemRe       = regexp.MustCompile(`^\s+- type:\s*\S`)
)

// issueForm is the top-level shape of a GitHub issue form.
type issueForm struct {
	scalars   map[string]string
	bodyItems int
}

// parseIssueForm parses the top level of a GitHub issue form YAML file: its
// `key: value` scalars and the number of `- type:` items under `body:`. It
// reports lines that cannot be YAML at this level (tabs, stray unindented text).
func parseIssueForm(text string) (issueForm, []string) {
	f := issueForm{scalars: map[string]string{}}
	var problems []string
	inBody := false
	for n, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "\t") {
			problems = append(problems, fmt.Sprintf("line %d: tab character", n+1))
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if line[0] != ' ' && line[0] != '-' {
			m := topLevelKeyRe.FindStringSubmatch(line)
			if m == nil {
				problems = append(problems, fmt.Sprintf("line %d: not a top-level `key:` line: %q", n+1, line))
				continue
			}
			f.scalars[m[1]] = strings.Trim(strings.TrimSpace(m[2]), `"'`)
			inBody = m[1] == "body"
			continue
		}
		if inBody && bodyItemRe.MatchString(line) {
			f.bodyItems++
		}
	}
	return f, problems
}

func TestParseIssueForm(t *testing.T) {
	good := "name: Bug\ndescription: d\nbody:\n  - type: input\n    id: v\n  - type: textarea\n"
	f, problems := parseIssueForm(good)
	if len(problems) != 0 {
		t.Errorf("parseIssueForm(good) problems = %q, want none", problems)
	}
	if got, want := f.bodyItems, 2; got != want {
		t.Errorf("parseIssueForm(good).bodyItems = %d, want %d", got, want)
	}
	if got, want := f.scalars["name"], "Bug"; got != want {
		t.Errorf("parseIssueForm(good) name = %q, want %q", got, want)
	}
	if _, problems := parseIssueForm("name: x\n\tbody: y\n"); len(problems) == 0 {
		t.Errorf("parseIssueForm(tab-indented) problems = none, want a tab problem")
	}
}

func TestIssueFormsParseAsYAML(t *testing.T) {
	for _, rel := range issueForms {
		t.Run(rel, func(t *testing.T) {
			f, problems := parseIssueForm(readOwned(t, rel))
			for _, p := range problems {
				t.Errorf("%s: %s", rel, p)
			}
			for _, key := range []string{"name", "description"} {
				if f.scalars[key] == "" {
					t.Errorf("%s: top-level %q is missing or empty", rel, key)
				}
			}
			if _, ok := f.scalars["body"]; !ok {
				t.Errorf("%s: no top-level body:", rel)
			}
			if f.bodyItems == 0 {
				t.Errorf("%s: body has no `- type:` items", rel)
			}
		})
	}
}

func TestBugFormAsksForSnapbackVersion(t *testing.T) {
	body := readOwned(t, bugFormPath)
	if !strings.Contains(body, "snapback version") {
		t.Errorf("%s does not ask for `snapback version` output", bugFormPath)
	}
	if strings.Contains(body, "snapback --version") {
		t.Errorf("%s asks for `snapback --version`, want `snapback version` (the command that exists)", bugFormPath)
	}
}

func TestRetiredMarkdownTemplatesAreGone(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range retiredTemplates {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			t.Errorf("%s still exists, want it replaced by the YAML issue forms", rel)
		}
	}
}

// configURLs returns every `url:` value in the issue chooser config.
func configURLs(text string) []string {
	var urls []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		if rest, ok := strings.CutPrefix(line, "url:"); ok {
			urls = append(urls, strings.Trim(strings.TrimSpace(rest), `"'`))
		}
	}
	return urls
}

func hasLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func TestIssueConfigRoutesSecurityPrivately(t *testing.T) {
	text := readOwned(t, issueConfigPath)
	if !hasLine(text, blankIssuesDisabled) {
		t.Errorf("%s: missing line %q", issueConfigPath, blankIssuesDisabled)
	}
	if urls := configURLs(text); !slices.Contains(urls, securityAdvisoryURL) {
		t.Errorf("%s: contact link urls = %q, want %q", issueConfigPath, urls, securityAdvisoryURL)
	}
}

func TestIssueConfigLinksDiscussions(t *testing.T) {
	text := readOwned(t, issueConfigPath)
	if !hasLine(text, blankIssuesDisabled) {
		t.Errorf("%s: missing line %q", issueConfigPath, blankIssuesDisabled)
	}
	if urls := configURLs(text); !slices.Contains(urls, discussionsURL) {
		t.Errorf("%s: contact link urls = %q, want %q", issueConfigPath, urls, discussionsURL)
	}
}
