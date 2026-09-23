package projectdocs

import (
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// privacyPage is the user-facing legal/trust document describing exactly what
// Snapback's opt-in telemetry can and cannot report.
const privacyPage = "docs-site/privacy.md"

// privacyProhibitedEntries are the exact "will never be reported" list items
// as they existed before S6-10/T3, captured verbatim from docs-site/privacy.md
// so the rewrite cannot silently drop or paraphrase one of them.
var privacyProhibitedEntries = []string{
	"1. File names, file paths, directory names or path fragments.",
	"2. Repository URIs, bucket names, host names or IP addresses of any backend.",
	"3. Your hostname, domain, usernames or home directory name.",
	"4. Snapshot IDs, tree IDs, blob IDs or backup tags.",
	"5. Config file contents, including instance names and environment variables.",
	"6. Free-text error strings, stack traces, command lines, arguments or log lines.",
	"7. Any file content, size distribution or directory listing.",
}

// privacyGoogleAnalyticsSection is the docs-site's own analytics disclosure,
// captured verbatim before S6-10/T3. It is unrelated to Snapback's telemetry
// and must survive the rewrite byte-for-byte.
const privacyGoogleAnalyticsSection = "The site you are reading is served with Google Analytics (measurement ID\n" +
	"`G-BFWW49ZP0E`). It records page views and the usual browser-side details that\n" +
	"come with them — page URL, referrer, approximate location derived from IP\n" +
	"address, browser and device type — and Google sets cookies to count returning\n" +
	"visitors. Blocking the script or using a content blocker does not affect the\n" +
	"docs.\n" +
	"\n" +
	"Site analytics are entirely separate from snapback itself. The CLI and the\n" +
	"daemon never talk to Google Analytics, and nothing you do on this site is\n" +
	"linked to anything on your machine."

// lineContaining returns the first line of text containing substr, or "".
func lineContaining(text, substr string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, substr) {
			return line
		}
	}
	return ""
}

// attrsForEvent builds a sample event via the real internal/telemetry
// constructor for name and returns its attribute keys in schema order, so the
// docs test pins against production code rather than a second, hand-copied
// list.
func attrsForEvent(t *testing.T, name string) []string {
	t.Helper()
	now := time.Now()

	var (
		ev  telemetry.Event
		err error
	)
	switch name {
	case "setup.completed":
		ev, err = telemetry.SetupCompleted("1.4.1", "ok", time.Second, now)
	case "daemon.started":
		ev, err = telemetry.DaemonStarted("1.4.1", now)
	case "mount.ready":
		ev, err = telemetry.MountReady("1.4.1", time.Second, now)
	case "doctor.failed":
		checks := telemetry.DoctorChecks()
		if len(checks) == 0 {
			t.Fatalf("telemetry.DoctorChecks() is empty")
		}
		ev, err = telemetry.DoctorFailed("1.4.1", checks[0], now)
	case "error":
		codes := telemetry.ErrorCodes()
		if len(codes) == 0 {
			t.Fatalf("telemetry.ErrorCodes() is empty")
		}
		ev, err = telemetry.ErrorEvent("1.4.1", codes[0], now)
	default:
		t.Fatalf("attrsForEvent: unknown event %q", name)
	}
	if err != nil {
		t.Fatalf("build sample %s event: %v", name, err)
	}

	keys := make([]string, len(ev.Attrs))
	for i, a := range ev.Attrs {
		keys[i] = a.Key
	}
	return keys
}

func TestPrivacyPageNoLongerClaimsNoTelemetryCode(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if strings.Contains(doc, "Snapback ships no telemetry code") {
		t.Errorf("%s still contains the outdated claim %q", privacyPage, "Snapback ships no telemetry code")
	}
}

func TestPrivacyPageStatesTelemetryCodeIsShipped(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "snapback ships telemetry code", "otlp", "crash reporting") {
		t.Errorf("%s has no line saying Snapback ships telemetry code, an OTLP exporter and crash reporting", privacyPage)
	}
}

func TestPrivacyPageStatesOffByDefault(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "off by default", "collected", "sent") {
		t.Errorf("%s has no line saying telemetry is off by default and nothing is collected or sent otherwise", privacyPage)
	}
}

func TestPrivacyPageStatesTwoSeparateSwitches(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "telemetry.enabled", "telemetry.crash_reports", "independent") {
		t.Errorf("%s has no line naming telemetry.enabled and telemetry.crash_reports as independent switches", privacyPage)
	}
	if !lineWithAll(doc, "turning one on never turns the other on") {
		t.Errorf("%s does not state that the two switches are independent of each other", privacyPage)
	}
}

func TestPrivacyPageStatesNoOnDiskQueue(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "no on-disk queue") {
		t.Errorf("%s has no line stating there is no on-disk queue", privacyPage)
	}
	if !lineWithAll(doc, "buffered in memory", "dropped") {
		t.Errorf("%s has no line stating events are buffered in memory only and dropped when undeliverable", privacyPage)
	}
}

func TestPrivacyPageStatesInstallIDLifecycle(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "128-bit", "install_id") {
		t.Errorf("%s has no line describing the 128-bit install identifier and its file name", privacyPage)
	}
	if !lineWithAll(doc, "actually used") {
		t.Errorf("%s does not state the install identifier is created only once telemetry is actually used", privacyPage)
	}
	if !lineWithAll(doc, "state_dir") {
		t.Errorf("%s does not state the install identifier is stored under the state directory", privacyPage)
	}
	if !lineWithAll(doc, "disable", "forgetinstallid") {
		t.Errorf("%s does not name ForgetInstallID as the mechanism `disable` uses to delete the install identifier", privacyPage)
	}
}

func TestPrivacyPageStatesRetentionIsTheOperatorsCollector(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !lineWithAll(doc, "retention", "your own collector") {
		t.Errorf("%s has no line saying retention is whatever the operator's own collector is configured for", privacyPage)
	}
	if !lineWithAll(doc, "maintainers never receive or retain anything") {
		t.Errorf("%s does not state that Snapback's maintainers never receive or retain anything", privacyPage)
	}
	if !lineWithAll(doc, "no maintainer-run") {
		t.Errorf("%s does not state there is no maintainer-run collector or endpoint", privacyPage)
	}
}

// TestPrivacyPageEventNamesMatchSchemaBidirectionally mirrors
// telemetry_docs_test.go's drift-proofing: the page's own "### The five
// events" list must equal internal/telemetry.Names() exactly, in both
// directions, so the two can never silently diverge.
func TestPrivacyPageEventNamesMatchSchemaBidirectionally(t *testing.T) {
	doc := readDoc(t, privacyPage)
	eventsSection := section(doc, "### The five events")
	if strings.TrimSpace(eventsSection) == "" {
		t.Fatalf("%s has no `### The five events` section", privacyPage)
	}

	onPage := eventNameTokens(eventsSection)
	want := map[string]bool{}
	for _, name := range telemetry.Names() {
		want[name] = true
	}

	for name := range want {
		if !onPage[name] {
			t.Errorf("%s events section is missing schema event %q", privacyPage, name)
		}
	}
	for name := range onPage {
		if !want[name] {
			t.Errorf("%s events section names %q, which is not in telemetry.Names() %v",
				privacyPage, name, sortedKeys(want))
		}
	}
}

// TestPrivacyPageEventAttributesMatchSchema pins the page's per-event
// attribute list against the real attributes each internal/telemetry
// constructor produces, so the page cannot list an attribute an event does
// not carry or omit one that it does.
func TestPrivacyPageEventAttributesMatchSchema(t *testing.T) {
	doc := readDoc(t, privacyPage)
	attrSection := section(doc, "### Event attributes")
	if strings.TrimSpace(attrSection) == "" {
		t.Fatalf("%s has no `### Event attributes` section", privacyPage)
	}

	for _, name := range telemetry.Names() {
		line := lineContaining(attrSection, "`"+name+"`")
		if line == "" {
			t.Fatalf("%s event attributes section has no line naming %q", privacyPage, name)
		}
		for _, key := range attrsForEvent(t, name) {
			if !strings.Contains(line, "`"+key+"`") {
				t.Errorf("%s attributes line for %q is missing attribute `%s`: %s", privacyPage, name, key, line)
			}
		}
	}
}

func TestPrivacyPageStatesTheFiveDurationBuckets(t *testing.T) {
	doc := readDoc(t, privacyPage)
	for _, bucket := range telemetry.Buckets() {
		if !strings.Contains(doc, "`"+bucket+"`") {
			t.Errorf("%s never states the duration bucket `%s`", privacyPage, bucket)
		}
	}
}

// TestPrivacyPageKeepsProhibitionListVerbatim pins every entry of the
// pre-existing "will never be reported" list to survive word for word.
func TestPrivacyPageKeepsProhibitionListVerbatim(t *testing.T) {
	doc := readDoc(t, privacyPage)
	if !strings.Contains(doc, "### What will never be reported") {
		t.Fatalf("%s is missing the `### What will never be reported` heading", privacyPage)
	}
	for _, entry := range privacyProhibitedEntries {
		if !strings.Contains(doc, entry) {
			t.Errorf("%s no longer contains the prohibition list entry verbatim: %q", privacyPage, entry)
		}
	}
	if !strings.Contains(doc, "None of the following may leave your machine, in any field, in any encoding,\nhashed or in the clear:") {
		t.Errorf("%s no longer contains the prohibition list's introductory sentence verbatim", privacyPage)
	}
	if !strings.Contains(doc, "A field that cannot be produced without one of the above will be dropped, not\nredacted.") {
		t.Errorf("%s no longer contains the prohibition list's closing sentence verbatim", privacyPage)
	}
}

// TestPrivacyPageGoogleAnalyticsSectionUnchanged asserts the docs-site's own
// analytics disclosure is byte-identical before and after the telemetry
// rewrite; it is unrelated to Snapback's own telemetry.
func TestPrivacyPageGoogleAnalyticsSectionUnchanged(t *testing.T) {
	doc := readDoc(t, privacyPage)
	got := strings.TrimSpace(section(doc, "## This documentation site uses Google Analytics"))
	want := strings.TrimSpace(privacyGoogleAnalyticsSection)
	if got != want {
		t.Errorf("%s Google Analytics section changed.\ngot:\n%s\nwant:\n%s", privacyPage, got, want)
	}
}
