package telemetry

import "regexp"

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

// prohibitedRules is the closed rule set, in report order. Adding a rule is one
// row here; the patterns stay tight enough that version strings, duration
// buckets and outcome labels never match.
var prohibitedRules = []prohibitedRule{
	{"absolute_path", regexp.MustCompile(`/[\w.-]+/[\w.-]+`)},
	{"windows_path", regexp.MustCompile(`[A-Za-z]:\\[\w.\\-]*`)},
	{"relative_path", regexp.MustCompile(`(?:^|[^:/\w])\w+/\w+`)},
	{"uri_scheme", regexp.MustCompile(`(?i)\b(?:sftp|scp|ssh|s3|b2|gs|azure|swift|rclone|rest|https?|ftps?):(?://)?[\w.~@/-]+`)},
	{"at_host", regexp.MustCompile(`\w[\w.-]*@[\w.-]+`)},
	{"hostname", regexp.MustCompile(`\b[A-Za-z][\w-]*(?:\.[A-Za-z0-9][\w-]*)*\.[A-Za-z]{2,}\b`)},
	{"snapshot_id", regexp.MustCompile(`\b[0-9a-f]{64}\b`)},
	{"short_id", regexp.MustCompile(`\b[0-9a-f]{8}\b`)},
	{"ipv4", regexp.MustCompile(`\b\d{1,3}(?:\.\d{1,3}){3}\b`)},
	{"ipv6", regexp.MustCompile(`(?i)\b[0-9a-f]{1,4}(?::[0-9a-f]{0,4}){2,7}`)},
	{"home_prefix", regexp.MustCompile(`~/[\w./-]*`)},
}

// ScanProhibited reports every prohibited value in b, one Finding per hit.
func ScanProhibited(b []byte) []Finding {
	var findings []Finding
	for _, rule := range prohibitedRules {
		for _, m := range rule.re.FindAll(b, -1) {
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
