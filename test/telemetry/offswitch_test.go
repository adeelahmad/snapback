// This file is S6-11/T3's acceptance test: it drives the real snapback
// binary as a subprocess (reusing collector_test.go's buildSnapbackBinary,
// runSnapback, newFakeCollector and writeConfig helpers) to check the
// telemetry off switch end to end, per docs/agents/sprint6-telemetry/plan.md
// row S6-11/T3:
//
//   - after `snapback telemetry disable`, the install-id file is gone
//   - the next run makes zero requests
//   - a `doctor -bundle` taken afterward contains no telemetry endpoint, no
//     install id and no event names (the redaction path)
//
// See TestTelemetryDisableStopsSendingAndBundleRedactsIdentity's comment for
// what this run found: `doctor -bundle`'s config.yaml member is NOT put
// through the telemetry-aware part of the redaction seam, so a configured
// telemetry.endpoint survives into the bundle verbatim. That assertion is
// left RED with this explanation rather than patched here, per this task's
// red-worker mandate (tests only, no production code beyond compile shims)
// and the sprint's practice of documenting a real finding instead of forcing
// a fix into a task dispatched to write tests.
package telemetry_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// readBundleMembers extracts every member of the gzipped tar at path into a
// name -> content map, mirroring internal/doctor/bundle_test.go's own
// readBundle helper (unexported there, so this test -- in a different
// package -- reads the real archive format itself rather than guessing it).
func readBundleMembers(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open(%q) = _, %v, want nil error", path, err)
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip.NewReader(%q) = _, %v, want nil error", path, err)
	}
	defer func() { _ = gz.Close() }()

	members := make(map[string]string)
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("tar.Next() = _, %v, want nil error", err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read tar member %s: %v, want nil error", hdr.Name, err)
		}
		members[hdr.Name] = string(body)
	}
	return members
}

// memberNames returns the sorted names of members, for readable failure
// messages.
func memberNames(members map[string]string) []string {
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// eventNamePattern matches any of the closed set of telemetry event names as
// a whole word, so a bundle member's incidental use of the common English
// word inside "error" (one of the five names) does not by itself make this
// pattern any looser than it needs to be -- it still only matches "error" (or
// one of the four dotted names) as a standalone token, never as part of a
// longer word such as "errors" or "no-such-restic-binary".
func eventNamePattern(t *testing.T) *regexp.Regexp {
	t.Helper()
	names := telemetry.Names()
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = regexp.QuoteMeta(name)
	}
	return regexp.MustCompile(`\b(?:` + strings.Join(parts, "|") + `)\b`)
}

// TestTelemetryDisableStopsSendingAndBundleRedactsIdentity pins S6-11/T3
// end to end against the real binary and a fake OTLP/HTTP collector:
//
//  1. telemetry enabled with an endpoint, a real doctor run against a
//     deliberately-broken restic binary sends at least one doctor.failed
//     event and mints an install-id file on disk (the same scenario
//     TestEnabledConfigDoctorRunProducesOnlyDoctorFailedEvents in
//     collector_test.go establishes for S6-11/T1).
//  2. `snapback telemetry disable` runs for real.
//  3. the install-id file is gone from disk (ForgetInstallID's real,
//     end-to-end effect).
//  4. a further real doctor run against the SAME fake collector makes zero
//     new requests: disable genuinely stops sending.
//  5. a real `doctor -bundle` run's archive is checked for the telemetry
//     endpoint, the (deleted) install id and every one of the five closed
//     event names.
func TestTelemetryDisableStopsSendingAndBundleRedactsIdentity(t *testing.T) {
	bin := buildSnapbackBinary(t)
	collector := newFakeCollector(t)
	dir := t.TempDir()
	tel := "telemetry:\n" +
		"  enabled: true\n" +
		"  endpoint: " + collector.srv.URL + "\n"
	cfgPath := writeConfig(t, dir, tel)
	env := os.Environ()
	installIDPath := filepath.Join(dir, "state", "telemetry", "install_id")

	// Step 1: enabled config, broken restic binary -> a real doctor.failed
	// emission mints the install id on disk.
	_, stderr, _ := runSnapback(t, bin, env, "--config", cfgPath, "doctor")
	if got := collector.Requests(); got == 0 {
		t.Fatalf("collector.Requests() after the first doctor run = 0, want at least 1 (stderr %q)", stderr)
	}
	installIDBytes, err := os.ReadFile(installIDPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) = _, %v, want the install id to exist after a real emission", installIDPath, err)
	}
	installID := strings.TrimSpace(string(installIDBytes))
	if installID == "" {
		t.Fatalf("install id file %q is empty, want a real id", installIDPath)
	}

	// Step 2: run the real `telemetry disable` command.
	_, disableStderr, disableCode := runSnapback(t, bin, env, "--config", cfgPath, "telemetry", "disable")
	if disableCode != 0 {
		t.Fatalf("telemetry disable exit = %d, want 0 (stderr %q)", disableCode, disableStderr)
	}

	// Step 3: the install-id file must be gone from disk.
	if _, err := os.Stat(installIDPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(%q) after disable err = %v, want fs.ErrNotExist", installIDPath, err)
	}

	// Step 4: a further real command against the same collector makes zero
	// NEW requests.
	reqsBeforeSecondRun := collector.Requests()
	_, stderr2, _ := runSnapback(t, bin, env, "--config", cfgPath, "doctor")
	if got := collector.Requests(); got != reqsBeforeSecondRun {
		t.Errorf("collector.Requests() after disable = %d, want unchanged %d (stderr %q)", got, reqsBeforeSecondRun, stderr2)
	}

	// Step 5: `doctor -bundle` must not leak the telemetry endpoint, the
	// deleted install id, or any of the five closed event names.
	bundleDir := t.TempDir()
	_, bundleStderr, bundleCode := runSnapback(t, bin, env, "--config", cfgPath, "doctor", "-bundle", bundleDir)
	if bundleCode != 0 {
		t.Fatalf("doctor -bundle exit = %d, want 0 (stderr %q)", bundleCode, bundleStderr)
	}
	matches, err := filepath.Glob(filepath.Join(bundleDir, "snapback-bundle-*.tar.gz"))
	if err != nil {
		t.Fatalf("Glob(snapback-bundle-*.tar.gz) = _, %v, want nil error", err)
	}
	if len(matches) != 1 {
		t.Fatalf("doctor -bundle %s wrote %v, want exactly 1 archive", bundleDir, matches)
	}
	members := readBundleMembers(t, matches[0])
	evPattern := eventNamePattern(t)

	for _, name := range memberNames(members) {
		body := members[name]
		if strings.Contains(body, collector.srv.URL) {
			t.Errorf("bundle member %q contains the telemetry endpoint %q:\n%s", name, collector.srv.URL, body)
		}
		if strings.Contains(body, installID) {
			t.Errorf("bundle member %q contains the deleted install id %q:\n%s", name, installID, body)
		}
		if m := evPattern.FindString(body); m != "" {
			t.Errorf("bundle member %q contains telemetry event name %q:\n%s", name, m, body)
		}
	}
}
