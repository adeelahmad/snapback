package community

import (
	"strings"
	"testing"
)

const (
	securityPath      = "SECURITY.md"
	headingSupported  = "## Supported versions"
	headingReporting  = "## Reporting a vulnerability"
	advisoriesURL     = "https://github.com/adeelahmad/snapback/security/advisories/new"
	privateReporting  = "private vulnerability reporting"
	noPublicIssue     = "do not open a public issue"
	preReleaseMention = "pre-release"
	minTableRows      = 3
)

// securitySection returns the body between heading and the next "## " line.
func securitySection(t *testing.T, heading string) string {
	t.Helper()
	lines := strings.Split(readOwned(t, securityPath), "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t\r") == heading {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: heading %q not found", securityPath, heading)
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func TestSecurityHasRequiredHeadings(t *testing.T) {
	body := readOwned(t, securityPath)

	got := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			got[strings.TrimRight(line, " \t\r")] = true
		}
	}

	for _, want := range []string{headingSupported, headingReporting} {
		if !got[want] {
			t.Errorf("%s: missing heading %q", securityPath, want)
		}
	}
}

func TestSecurityHasPrivateReportingChannel(t *testing.T) {
	section := securitySection(t, headingReporting)

	if !strings.Contains(section, advisoriesURL) {
		t.Errorf("%s %q section: missing advisory URL %q", securityPath, headingReporting, advisoriesURL)
	}
	if !strings.Contains(strings.ToLower(section), privateReporting) {
		t.Errorf("%s %q section: missing phrase %q (case-insensitive)", securityPath, headingReporting, privateReporting)
	}
}

func TestSecurityWarnsAgainstPublicIssues(t *testing.T) {
	section := securitySection(t, headingReporting)

	if !strings.Contains(strings.ToLower(section), noPublicIssue) {
		t.Fatalf("%s %q section: missing phrase %q (case-insensitive)", securityPath, headingReporting, noPublicIssue)
	}
}

func TestSecuritySupportedVersionsTable(t *testing.T) {
	section := securitySection(t, headingSupported)

	var rows []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			rows = append(rows, trimmed)
		}
	}
	if len(rows) < minTableRows {
		t.Fatalf("%s %q section: %d table rows, want header + separator + at least one data row", securityPath, headingSupported, len(rows))
	}
	if strings.Trim(rows[1], "|:- \t") != "" || !strings.Contains(rows[1], "-") {
		t.Errorf("%s %q section: second table row %q is not a separator row", securityPath, headingSupported, rows[1])
	}
	if !strings.Contains(strings.ToLower(section), preReleaseMention) {
		t.Errorf("%s %q section: missing %q (case-insensitive)", securityPath, headingSupported, preReleaseMention)
	}
}
