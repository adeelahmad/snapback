package telemetry

import (
	"bytes"
	"regexp"
	"strings"
)

// Finding is one hit of a prohibited-value rule in a scanned payload.
type Finding struct {
	// Rule names the rule that matched, as listed by ProhibitedRules.
	Rule string
	// Match is the matched text.
	Match string
}

// prohibitedRule is one row of the closed matcher table: a rule name and the
// pattern whose every match becomes a Finding.
type prohibitedRule struct {
	name string
	re   *regexp.Regexp
}

// hostnamePattern matches a bare hostname or FQDN (e.g. a machine name or a
// dotted username like "j.doe"). versionAttrs mirrors this rule to reject the
// same shape in a version string, so the two checks stay consistent.
var hostnamePattern = regexp.MustCompile(`\b[A-Za-z][\w-]*(?:\.[A-Za-z0-9][\w-]*)*\.[A-Za-z]{2,}\b`)

// snapshotIDPattern matches a 64-character hex snapshot id. versionAttrs
// mirrors this rule to reject the same shape in a version string, so the two
// checks stay consistent.
var snapshotIDPattern = regexp.MustCompile(`\b[0-9a-f]{64}\b`)

// prohibitedRules is the closed rule set, in report order. Adding a rule is one
// row here; the patterns stay tight enough that version strings, duration
// buckets and outcome labels never match.
var prohibitedRules = []prohibitedRule{
	{"absolute_path", regexp.MustCompile(`/[\w.-]+/[\w.-]+`)},
	{"windows_path", regexp.MustCompile(`[A-Za-z]:\\[\w.\\-]*`)},
	{"relative_path", regexp.MustCompile(`(?:^|[^:/\w])\w+/\w+`)},
	{"uri_scheme", regexp.MustCompile(`(?i)\b(?:sftp|scp|ssh|s3|b2|gs|azure|swift|rclone|rest|https?|ftps?):(?://)?[\w.~@/-]+`)},
	{"at_host", regexp.MustCompile(`\w[\w.-]*@[\w.-]+`)},
	{"hostname", hostnamePattern},
	{"snapshot_id", snapshotIDPattern},
	{"short_id", regexp.MustCompile(`\b[0-9a-f]{8}\b`)},
	{"ipv4", regexp.MustCompile(`\b\d{1,3}(?:\.\d{1,3}){3}\b`)},
	{"ipv6", regexp.MustCompile(`(?i)\b[0-9a-f]{1,4}(?::[0-9a-f]{0,4}){2,7}`)},
	{"home_prefix", regexp.MustCompile(`~/[\w./-]*`)},
}

// snapbackModulePath matches Snapback's own module path, per ruling S6-R1:
// identical on every install, never a hostname or filesystem path belonging
// to a user, so it is exempt from every rule below. The \b after "snapback"
// admits both the bare module path (the OTLP scope name in
// internal/telemetry/otlp/encode.go) and any deeper package path, while
// still refusing to match a different, merely similarly-prefixed path such
// as "github.com/adeelahmad/snapbackup/...".
var snapbackModulePath = regexp.MustCompile(`github\.com/adeelahmad/snapback\b(?:/[\w./-]*)?`)

// eventNamePattern matches the closed, exhaustive set of telemetry event
// names in eventNames (event.go). Four of the five are dotted lowercase
// words (e.g. "setup.completed") that are shaped exactly like hostnamePattern
// expects a hostname to look, but they are schema names Snapback itself
// chose, never a leaked hostname, so they are exempt from every rule below.
var eventNamePattern = regexp.MustCompile(eventNameAlternation())

// eventNameAlternation builds the regexp alternation source for
// eventNamePattern from the closed event-name list, so the exemption can
// never drift from the schema it exists to describe.
func eventNameAlternation() string {
	parts := make([]string, len(eventNames))
	for i, name := range eventNames {
		parts[i] = regexp.QuoteMeta(name)
	}
	return `\b(?:` + strings.Join(parts, "|") + `)\b`
}

// blank replaces m with spaces of the same length, so masking a match never
// shifts the byte offsets later rules see.
func blank(m []byte) []byte {
	return bytes.Repeat([]byte{' '}, len(m))
}

// ScanProhibited reports every prohibited value in b, one Finding per hit.
func ScanProhibited(b []byte) []Finding {
	masked := snapbackModulePath.ReplaceAllFunc(b, blank)
	masked = eventNamePattern.ReplaceAllFunc(masked, blank)
	var findings []Finding
	for _, rule := range prohibitedRules {
		for _, m := range rule.re.FindAll(masked, -1) {
			findings = append(findings, Finding{Rule: rule.name, Match: string(m)})
		}
	}
	return findings
}

// ProhibitedRules lists the rule names ScanProhibited can report, in table
// order. Each call returns a fresh slice the caller may mutate freely.
func ProhibitedRules() []string {
	names := make([]string, 0, len(prohibitedRules))
	for _, rule := range prohibitedRules {
		names = append(names, rule.name)
	}
	return names
}
