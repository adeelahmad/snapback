package community

import (
	"regexp"
	"strings"
	"testing"
)

const (
	bugTemplatePath     = ".github/ISSUE_TEMPLATE/bug_report.md"
	featureTemplatePath = ".github/ISSUE_TEMPLATE/feature_request.md"
	issueConfigPath     = ".github/ISSUE_TEMPLATE/config.yml"
	frontMatterFence    = "---"
	blankIssuesDisabled = "blank_issues_enabled: false"
	securityAdvisoryURL = "https://github.com/adeelahmad/snapback/security/advisories/new"
)

var (
	frontMatterKey      = regexp.MustCompile(`^[a-z_]+$`)
	requiredFrontMatter = []string{"name", "about", "title", "labels"}
	issueTemplates      = []string{bugTemplatePath, featureTemplatePath}
)

// splitFrontMatter parses the leading `---` block of an issue template into
// key/value pairs and returns the body that follows the closing fence.
func splitFrontMatter(t *testing.T, rel string) (map[string]string, string) {
	t.Helper()
	lines := strings.Split(readOwned(t, rel), "\n")
	if strings.TrimRight(lines[0], " \t\r") != frontMatterFence {
		t.Fatalf("%s: line 1 = %q, want %q", rel, lines[0], frontMatterFence)
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t\r") == frontMatterFence {
			end = i
			break
		}
	}
	if end < 0 {
		t.Fatalf("%s: no closing %q for front matter", rel, frontMatterFence)
	}

	fields := map[string]string{}
	for n, line := range lines[1:end] {
		key, value, ok := strings.Cut(strings.TrimRight(line, " \t\r"), ":")
		if !ok || !frontMatterKey.MatchString(key) {
			t.Fatalf("%s: front matter line %d %q is not `key: value`", rel, n+2, line)
		}
		fields[key] = strings.TrimSpace(value)
	}
	return fields, strings.Join(lines[end+1:], "\n")
}

func TestIssueTemplatesHaveValidFrontMatter(t *testing.T) {
	for _, rel := range issueTemplates {
		t.Run(rel, func(t *testing.T) {
			fields, _ := splitFrontMatter(t, rel)
			for _, key := range requiredFrontMatter {
				v, ok := fields[key]
				if !ok {
					t.Errorf("%s: front matter missing key %q", rel, key)
					continue
				}
				if strings.Trim(v, `"'`) == "" {
					t.Errorf("%s: front matter key %q is empty", rel, key)
				}
			}
		})
	}
}

func TestIssueTemplateTitlesAreConventional(t *testing.T) {
	want := map[string]string{
		bugTemplatePath:     "fix: ",
		featureTemplatePath: "feat: ",
	}
	for _, rel := range issueTemplates {
		t.Run(rel, func(t *testing.T) {
			fields, _ := splitFrontMatter(t, rel)
			raw := fields["title"]
			got := strings.Trim(raw, `"'`)
			if got != want[rel] {
				t.Errorf("%s: title = %q (raw %q), want %q", rel, got, raw, want[rel])
			}
		})
	}
}

func TestIssueTemplateBodiesAreNotPlaceholder(t *testing.T) {
	want := map[string][]string{
		bugTemplatePath:     {"Steps to reproduce", "Expected", "Actual", "snapback version"},
		featureTemplatePath: {"Problem", "Proposal", "Alternatives"},
	}
	for _, rel := range issueTemplates {
		t.Run(rel, func(t *testing.T) {
			_, body := splitFrontMatter(t, rel)
			for _, prompt := range want[rel] {
				if !strings.Contains(body, prompt) {
					t.Errorf("%s: body missing prompt %q", rel, prompt)
				}
			}
		})
	}
}

func TestIssueConfigRoutesSecurityPrivately(t *testing.T) {
	var hasBlankDisabled, hasAdvisoryURL bool
	for _, line := range strings.Split(readOwned(t, issueConfigPath), "\n") {
		line = strings.TrimSpace(line)
		if line == blankIssuesDisabled {
			hasBlankDisabled = true
		}
		if rest, ok := strings.CutPrefix(line, "url:"); ok && strings.Trim(strings.TrimSpace(rest), `"'`) == securityAdvisoryURL {
			hasAdvisoryURL = true
		}
	}
	if !hasBlankDisabled {
		t.Errorf("%s: missing line %q", issueConfigPath, blankIssuesDisabled)
	}
	if !hasAdvisoryURL {
		t.Errorf("%s: missing `url: %s` line", issueConfigPath, securityAdvisoryURL)
	}
}
