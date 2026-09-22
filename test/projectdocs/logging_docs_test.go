package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// loggingPage is the published page that documents Snapback's logging.
const loggingPage = "docs-site/logging.md"

// loggingDoc returns loggingPage, or "" when it is missing or unreadable. It
// reports the read failure instead of aborting, so every content assertion
// below still runs and names what the page has to say.
func loggingDoc(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), loggingPage))
	if err != nil {
		t.Errorf("read %s: %v", loggingPage, err)
		return ""
	}
	return string(data)
}

// yamlBlockWith returns the first ```yaml fenced block of doc containing want,
// or "" when no such block exists.
func yamlBlockWith(doc, want string) string {
	const fence = "```"
	lines := strings.Split(doc, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != fence+"yaml" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) != fence {
				continue
			}
			if block := strings.Join(lines[i+1:j], "\n"); strings.Contains(block, want) {
				return block
			}
			i = j
			break
		}
	}
	return ""
}

// lineWithAll reports whether some line of doc contains every want, compared
// without case. It pins a claim to one sentence rather than to the whole page.
func lineWithAll(doc string, want ...string) bool {
	for _, line := range strings.Split(doc, "\n") {
		lower := strings.ToLower(line)
		found := true
		for _, w := range want {
			if !strings.Contains(lower, strings.ToLower(w)) {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

func TestLoggingDocsPageExists(t *testing.T) {
	if doc := loggingDoc(t); strings.TrimSpace(doc) == "" {
		t.Errorf("%s is empty, want a page documenting Snapback's logging", loggingPage)
	}
}

func TestLoggingDocsNamesEveryFlag(t *testing.T) {
	doc := loggingDoc(t)

	for _, flag := range []string{"--log-level", "--log-format", "--log-file"} {
		if !strings.Contains(doc, flag) {
			t.Errorf("%s does not mention the %s flag", loggingPage, flag)
		}
	}
}

func TestLoggingDocsShowsLoggingConfigKeys(t *testing.T) {
	doc := loggingDoc(t)

	block := yamlBlockWith(doc, "logging:")
	if block == "" {
		t.Fatalf("%s has no ```yaml example containing a `logging:` block", loggingPage)
	}
	for _, key := range []string{"level:", "format:", "file:"} {
		if !strings.Contains(block, key) {
			t.Errorf("%s `logging:` example %q has no %q key", loggingPage, block, key)
		}
	}
}

func TestLoggingDocsNamesEveryLevelAndFormat(t *testing.T) {
	doc := loggingDoc(t)

	for _, word := range []string{"debug", "info", "warn", "error", "text", "json"} {
		re := regexp.MustCompile(`(?i)\b` + word + `\b`)
		if !re.MatchString(doc) {
			t.Errorf("%s does not name %q", loggingPage, word)
		}
	}
}

func TestLoggingDocsSaysResticPasswordIsRedacted(t *testing.T) {
	doc := loggingDoc(t)

	if !lineWithAll(doc, "debug", "restic", "password", "redacted") {
		t.Errorf("%s has no sentence saying the restic command line is logged at debug "+
			"with the password redacted", loggingPage)
	}
}

func TestLoggingDocsSaysBundleIncludesTheLogFile(t *testing.T) {
	doc := loggingDoc(t)

	if !lineWithAll(doc, "--bundle", "log file") {
		t.Errorf("%s has no sentence saying `doctor --bundle` picks up the log file", loggingPage)
	}
}

// unshippedLoggingRe names logging features Snapback does not have, so the page
// cannot promise them.
var unshippedLoggingRe = regexp.MustCompile(`(?i)(syslog|journald|log rotation)`)

func TestLoggingDocsClaimsNoUnshippedFeatures(t *testing.T) {
	doc := loggingDoc(t)

	for i, line := range strings.Split(doc, "\n") {
		for _, m := range unshippedLoggingRe.FindAllString(line, -1) {
			t.Errorf("%s:%d: claims the unshipped feature %q", loggingPage, i+1, m)
		}
	}
}
