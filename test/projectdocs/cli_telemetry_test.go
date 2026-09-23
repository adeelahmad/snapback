package projectdocs

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// cliPage is the published CLI reference page.
const cliPage = "docs-site/cli.md"

// telemetryArgLine matches one "Args:" entry of the real `telemetry -h`
// output, e.g. "  status   report whether telemetry is enabled".
var telemetryArgLine = regexp.MustCompile(`^\s{2}(\S+)\s+(.+)$`)

// telemetryFlagLine matches a flag entry of the real `telemetry -h` output,
// e.g. "  --json".
var telemetryFlagLine = regexp.MustCompile(`^\s{2}--?([A-Za-z][A-Za-z0-9-]*)`)

// buildSnapbackBinary builds the CLI into a temporary directory and returns
// its path.
func buildSnapbackBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "snapback")
	build := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	build.Dir = repoRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	return bin
}

// telemetryHelp runs the actual telemetry help screen and returns its output.
func telemetryHelp(t *testing.T, bin string) string {
	t.Helper()
	// telemetry -h exits 0; the text is what matters either way.
	out, _ := exec.Command(bin, "telemetry", "-h").CombinedOutput()
	return string(out)
}

// telemetryArgDescriptions parses the "Args:" block of the real telemetry
// help screen into verb -> one-line description.
func telemetryArgDescriptions(t *testing.T, help string) map[string]string {
	t.Helper()
	descriptions := map[string]string{}
	inArgs := false
	for _, line := range strings.Split(help, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Args:" {
			inArgs = true
			continue
		}
		if !inArgs {
			continue
		}
		if trimmed == "" {
			break
		}
		m := telemetryArgLine.FindStringSubmatch(line)
		if m == nil {
			t.Fatalf("telemetry -h Args line %q does not match %q", line, telemetryArgLine)
		}
		descriptions[m[1]] = m[2]
	}
	return descriptions
}

// telemetryFlags parses the "Flags:" block of the real telemetry help screen
// into the set of flag names it registers.
func telemetryFlags(help string) map[string]bool {
	flags := map[string]bool{}
	inFlags := false
	for _, line := range strings.Split(help, "\n") {
		if strings.TrimSpace(line) == "Flags:" {
			inFlags = true
			continue
		}
		if !inFlags {
			continue
		}
		if m := telemetryFlagLine.FindStringSubmatch(line); m != nil {
			flags[m[1]] = true
		}
	}
	return flags
}

// TestCLIDocsPinTelemetryVerbHelpText asserts docs-site/cli.md's telemetry
// section carries the exact, verbatim one-line help text the binary prints
// for each of its four verbs, so the doc cannot silently drift from the CLI.
func TestCLIDocsPinTelemetryVerbHelpText(t *testing.T) {
	bin := buildSnapbackBinary(t)
	help := telemetryHelp(t, bin)

	descriptions := telemetryArgDescriptions(t, help)
	wantVerbs := []string{"status", "show", "enable", "disable"}
	if len(descriptions) != len(wantVerbs) {
		t.Fatalf("telemetry -h Args block has %d verbs %v, want %d %v",
			len(descriptions), descriptions, len(wantVerbs), wantVerbs)
	}

	doc := readDoc(t, cliPage)
	telemetrySection := section(doc, "## snapback telemetry")
	if strings.TrimSpace(telemetrySection) == "" {
		t.Fatalf("%s has no `## snapback telemetry` section", cliPage)
	}

	for _, verb := range wantVerbs {
		desc, ok := descriptions[verb]
		if !ok {
			t.Fatalf("telemetry -h Args block does not name verb %q:\n%s", verb, help)
		}
		if !lineWithAll(telemetrySection, verb, desc) {
			t.Errorf("%s telemetry section has no line naming verb %q with its exact help text %q",
				cliPage, verb, desc)
		}
	}
}

// TestCLIDocsPinTelemetryFlags asserts docs-site/cli.md documents exactly the
// flags the real telemetry command registers, no more and no less.
func TestCLIDocsPinTelemetryFlags(t *testing.T) {
	bin := buildSnapbackBinary(t)
	flags := telemetryFlags(telemetryHelp(t, bin))
	if len(flags) == 0 {
		t.Fatalf("telemetry -h registers no flags")
	}

	doc := readDoc(t, cliPage)
	telemetrySection := section(doc, "## snapback telemetry")
	if strings.TrimSpace(telemetrySection) == "" {
		t.Fatalf("%s has no `## snapback telemetry` section", cliPage)
	}

	for flag := range flags {
		if !strings.Contains(telemetrySection, "`--"+flag+"`") {
			t.Errorf("%s telemetry section does not document flag --%s", cliPage, flag)
		}
	}

	docFlagToken := regexp.MustCompile("`--([A-Za-z][A-Za-z0-9-]*)`")
	for _, m := range docFlagToken.FindAllStringSubmatch(telemetrySection, -1) {
		if !flags[m[1]] {
			t.Errorf("%s telemetry section documents flag --%s, which the binary does not register",
				cliPage, m[1])
		}
	}
}

// TestCLIDocsNoteDoctorBundleUnaffectedByTelemetry pins the D8 acceptance
// criterion: `doctor --bundle` behaves the same regardless of telemetry
// settings, and the CLI reference says so.
func TestCLIDocsNoteDoctorBundleUnaffectedByTelemetry(t *testing.T) {
	doc := readDoc(t, cliPage)
	doctorSection := section(doc, "## snapback doctor")
	if strings.TrimSpace(doctorSection) == "" {
		t.Fatalf("%s has no `## snapback doctor` section", cliPage)
	}
	if !lineWithAll(doctorSection, "--bundle", "unaffected", "telemetry") {
		t.Errorf("%s doctor section has no line saying `doctor --bundle` is unaffected by "+
			"telemetry settings", cliPage)
	}
}
