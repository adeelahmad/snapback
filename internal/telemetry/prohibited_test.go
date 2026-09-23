package telemetry

import (
	"reflect"
	"testing"
)

// wantProhibitedRules is the closed rule set, in the order ProhibitedRules must
// return it. Adding a rule is one row here and one row in the matcher table.
var wantProhibitedRules = []string{
	"absolute_path",
	"windows_path",
	"relative_path",
	"uri_scheme",
	"at_host",
	"hostname",
	"snapshot_id",
	"short_id",
	"ipv4",
	"ipv6",
	"home_prefix",
}

// cleanPayload is an OTLP-shaped body with nothing prohibited in it.
const cleanPayload = `{"resourceMetrics":[{"resource":{"attributes":` +
	`[{"key":"os","value":{"stringValue":"linux"}}]}}]}`

func TestProhibitedRulesAreTheClosedRuleSet(t *testing.T) {
	got := ProhibitedRules()
	if !reflect.DeepEqual(got, wantProhibitedRules) {
		t.Fatalf("ProhibitedRules() = %q, want %q", got, wantProhibitedRules)
	}
}

func TestProhibitedRulesReturnsACopy(t *testing.T) {
	first := ProhibitedRules()
	if len(first) == 0 {
		t.Fatalf("ProhibitedRules() = %q, want %d rules", first, len(wantProhibitedRules))
	}
	first[0] = "mutated"
	if got := ProhibitedRules()[0]; got == "mutated" {
		t.Fatalf("ProhibitedRules()[0] = %q after mutating an earlier result, want the rule name", got)
	}
}

func TestScanProhibitedFindsHostileValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		rule  string
	}{
		{"unix absolute path", "/home/u/x", "absolute_path"},
		{"windows path", `C:\u\x`, "windows_path"},
		{"relative path", "dir/file", "relative_path"},
		{"sftp uri", "sftp://h/r", "uri_scheme"},
		{"s3 uri", "s3:bucket", "uri_scheme"},
		{"rest uri", "rest:https://x", "uri_scheme"},
		{"b2 uri", "b2:bucket", "uri_scheme"},
		{"user at host", "user@backup", "at_host"},
		{"mdns hostname", "nas.local", "hostname"},
		{"lan hostname", "backup.lan", "hostname"},
		{"fqdn hostname", "backup.example.com", "hostname"},
		{"lookalike module path", "github.com/other/repo", "hostname"},
		{
			"snapshot id",
			"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			"snapshot_id",
		},
		{"short id", "snapshot 1a2b3c4d", "short_id"},
		{"ipv4 literal", "192.168.1.10", "ipv4"},
		{"ipv6 literal", "fe80::1", "ipv6"},
		{"home prefix", "~/x", "home_prefix"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScanProhibited([]byte(tt.input))
			if len(got) == 0 {
				t.Fatalf("ScanProhibited(%q) = no findings, want at least one with rule %q", tt.input, tt.rule)
			}
			for _, f := range got {
				if f.Rule == tt.rule {
					return
				}
			}
			t.Fatalf("ScanProhibited(%q) = %+v, want a finding with rule %q", tt.input, got, tt.rule)
		})
	}
}

func TestScanProhibitedReportsEveryHit(t *testing.T) {
	const input = "/home/u/x /home/u/y"
	got := ScanProhibited([]byte(input))
	var absolute int
	for _, f := range got {
		if f.Rule == "absolute_path" {
			absolute++
		}
	}
	if absolute < 2 {
		t.Fatalf("ScanProhibited(%q) = %+v with %d absolute_path findings, want one per hit (2)", input, got, absolute)
	}
}

func TestScanProhibitedNamesOnlyKnownRules(t *testing.T) {
	known := make(map[string]bool, len(wantProhibitedRules))
	for _, rule := range wantProhibitedRules {
		known[rule] = true
	}
	got := ScanProhibited([]byte("/home/u/x sftp://h/r 192.168.1.10"))
	if len(got) == 0 {
		t.Fatal("ScanProhibited(hostile payload) = no findings, want findings naming known rules")
	}
	for _, f := range got {
		if !known[f.Rule] {
			t.Errorf("ScanProhibited returned rule %q, want one of %q", f.Rule, wantProhibitedRules)
		}
		if f.Match == "" {
			t.Errorf("ScanProhibited finding %+v has an empty Match, want the matched text", f)
		}
	}
}

// TestScanProhibitedExemptsSnapbackModulePath pins ruling S6-R1: Snapback's
// own module path is not a hostname, filesystem path or URI belonging to a
// user, so a crash frame naming it must yield zero findings.
func TestScanProhibitedExemptsSnapbackModulePath(t *testing.T) {
	const payload = `{"module":"github.com/adeelahmad/snapback/internal/mount",` +
		`"function":"(*Catalog).Refresh"}`
	if got := ScanProhibited([]byte(payload)); len(got) != 0 {
		t.Fatalf("ScanProhibited(%q) = %+v, want no findings", payload, got)
	}
}

// TestScanProhibitedExemptsCrashReleaseString pins ruling S6-R4: the crash
// envelope's own "snapback@<version>" release identifier is not a leaked
// email or user@host string, so it must yield zero findings.
func TestScanProhibitedExemptsCrashReleaseString(t *testing.T) {
	const payload = `{"release":"snapback@0.1.0"}`
	if got := ScanProhibited([]byte(payload)); len(got) != 0 {
		t.Fatalf("ScanProhibited(%q) = %+v, want no findings", payload, got)
	}
}

// TestScanProhibitedStillFlagsNonVersionSnapbackAtHost pins the narrow half
// of ruling S6-R4: a value that merely starts with "snapback@" but is not
// version-shaped (attacker- or user-controlled content) must still be
// caught by at_host, proving the exemption is not "snapback@<anything>".
func TestScanProhibitedStillFlagsNonVersionSnapbackAtHost(t *testing.T) {
	const payload = `{"release":"snapback@evil.example.com"}`
	got := ScanProhibited([]byte(payload))
	for _, f := range got {
		if f.Rule == "at_host" {
			return
		}
	}
	t.Fatalf("ScanProhibited(%q) = %+v, want a finding with rule %q", payload, got, "at_host")
}

func TestScanProhibitedPassesACleanPayload(t *testing.T) {
	payloads := []struct {
		name string
		body string
	}{
		{"otlp body", cleanPayload},
		{"duration label", "<100ms"},
		{"outcome label", "ok"},
		{"failure label", "mount_failure"},
		{"version label", "1.4.1"},
		{
			"otlp body with labels",
			cleanPayload + ` <100ms ok mount_failure 1.4.1`,
		},
	}
	for _, p := range payloads {
		t.Run(p.name, func(t *testing.T) {
			if got := ScanProhibited([]byte(p.body)); len(got) != 0 {
				t.Fatalf("ScanProhibited(%q) = %+v, want no findings", p.body, got)
			}
		})
	}
}
