package projectdocs

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// telemetryDocsPage is the user-facing "what telemetry does" overview: config
// keys, event names, duration buckets, the opt-in contract and the endpoint
// rules. It is distinct from telemetry-collector.md (how to run a collector)
// and privacy.md (the full prohibition list and per-event attributes).
const telemetryDocsPage = "docs-site/telemetry.md"

// backtickToken matches one inline-code span's contents.
var backtickToken = regexp.MustCompile("`([^`]+)`")

func TestTelemetryDocsPageExists(t *testing.T) {
	if doc := readDoc(t, telemetryDocsPage); strings.TrimSpace(doc) == "" {
		t.Errorf("%s is empty, want a page describing what telemetry reports", telemetryDocsPage)
	}
}

// telemetryConfigKeys are the exact `telemetry.*` YAML keys from
// internal/config/types.go's Telemetry struct.
var telemetryConfigKeys = []string{
	"telemetry.enabled",
	"telemetry.endpoint",
	"telemetry.crash_reports",
	"telemetry.crash_endpoint",
}

func TestTelemetryDocsStatesEveryConfigKey(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	for _, key := range telemetryConfigKeys {
		if !strings.Contains(doc, "`"+key+"`") {
			t.Errorf("%s never states the config key `%s`", telemetryDocsPage, key)
		}
	}
}

// eventNameTokens extracts the backtick-quoted tokens that look like an event
// name (letters, optionally one or more dot-separated segments) from a page
// section, so the test can compare the page's own list against the schema
// without hardcoding the schema a second time.
func eventNameTokens(section string) map[string]bool {
	nameLike := regexp.MustCompile(`^[a-z]+(\.[a-z]+)*$`)
	names := map[string]bool{}
	for _, m := range backtickToken.FindAllStringSubmatch(section, -1) {
		if nameLike.MatchString(m[1]) {
			names[m[1]] = true
		}
	}
	return names
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestTelemetryDocsEventNamesMatchSchemaBidirectionally pins the page's event
// list to internal/telemetry.Names(): every schema name must appear on the
// page, and the page must name nothing beyond the schema, so the two can
// never silently drift apart as events are added or removed.
func TestTelemetryDocsEventNamesMatchSchemaBidirectionally(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	eventsSection := section(doc, "## The five events")
	if strings.TrimSpace(eventsSection) == "" {
		t.Fatalf("%s has no `## The five events` section", telemetryDocsPage)
	}

	onPage := eventNameTokens(eventsSection)
	want := map[string]bool{}
	for _, name := range telemetry.Names() {
		want[name] = true
	}

	for name := range want {
		if !onPage[name] {
			t.Errorf("%s events section is missing schema event %q", telemetryDocsPage, name)
		}
	}
	for name := range onPage {
		if !want[name] {
			t.Errorf("%s events section names %q, which is not in telemetry.Names() %v",
				telemetryDocsPage, name, sortedKeys(want))
		}
	}
}

func TestTelemetryDocsStatesTheFiveDurationBuckets(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	for _, bucket := range telemetry.Buckets() {
		if !strings.Contains(doc, "`"+bucket+"`") {
			t.Errorf("%s never states the duration bucket `%s`", telemetryDocsPage, bucket)
		}
	}
}

func TestTelemetryDocsStatesOffByDefault(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	if !lineWithAll(doc, "off", "default") {
		t.Errorf("%s has no sentence saying telemetry is off by default", telemetryDocsPage)
	}
}

func TestTelemetryDocsStatesNoShippedCollectorOrEndpoint(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	if !lineWithAll(doc, "snapback", "no", "collector") {
		t.Errorf("%s has no sentence saying Snapback ships no collector", telemetryDocsPage)
	}
	if !lineWithAll(doc, "no default", "endpoint") {
		t.Errorf("%s has no sentence saying Snapback ships no default endpoint", telemetryDocsPage)
	}
}

func TestTelemetryDocsStatesEnableRefusesWithoutEndpoint(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	if !lineWithAll(doc, "enable", "endpoint", "must be set") {
		t.Errorf("%s has no sentence saying `enable` refuses without telemetry.endpoint set", telemetryDocsPage)
	}
}

func TestTelemetryDocsStatesTheLoopbackHTTPSRule(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	for _, host := range []string{"127.0.0.1", "::1", "localhost"} {
		if !strings.Contains(doc, host) {
			t.Errorf("%s never names the loopback literal %q", telemetryDocsPage, host)
		}
	}
	if !lineWithAll(doc, "http", "loopback") {
		t.Errorf("%s has no sentence tying the http scheme to a loopback host", telemetryDocsPage)
	}
	if !lineWithAll(doc, "https") {
		t.Errorf("%s never requires https for non-loopback endpoints", telemetryDocsPage)
	}
}

func TestTelemetryDocsStatesTheComposeCommand(t *testing.T) {
	doc := readDoc(t, telemetryDocsPage)
	if !strings.Contains(doc, "docker compose up") {
		t.Errorf("%s never shows the `docker compose up` command", telemetryDocsPage)
	}
	for _, name := range []string{"otel-collector", "glitchtip"} {
		if !strings.Contains(doc, name) {
			t.Errorf("%s never names the %q compose service", telemetryDocsPage, name)
		}
	}
}
