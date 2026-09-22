package setup

import (
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// consentLines splits the notice into its lines, dropping the trailing newline
// so an empty last element never counts as a line.
func consentLines(t *testing.T) []string {
	t.Helper()
	return strings.Split(strings.TrimSuffix(ConsentText(), "\n"), "\n")
}

func TestConsentTextNamesEveryEvent(t *testing.T) {
	got := ConsentText()
	names := telemetry.Names()
	if len(names) != 5 {
		t.Fatalf("telemetry.Names() has %d names, want 5: %q", len(names), names)
	}
	for _, name := range names {
		if !strings.Contains(got, name) {
			t.Errorf("ConsentText() does not name %q; text is:\n%s", name, got)
		}
	}
}

func TestConsentTextStatesOperatorSuppliesCollector(t *testing.T) {
	got := ConsentText()
	for _, want := range []string{"collector", "nothing is sent", "off by default"} {
		if !strings.Contains(got, want) {
			t.Errorf("ConsentText() does not contain %q; text is:\n%s", want, got)
		}
	}
}

func TestConsentTextLinksPrivacyPage(t *testing.T) {
	const want = "https://snapback.run/privacy"
	if got := ConsentText(); !strings.Contains(got, want) {
		t.Errorf("ConsentText() does not link %q; text is:\n%s", want, got)
	}
}

func TestConsentTextHasNoPlaceholder(t *testing.T) {
	got := ConsentText()
	if got == "" {
		t.Fatalf("ConsentText() is empty, want the telemetry notice")
	}
	placeholders := []string{
		"<host>", "<HOST>", "<url>", "<URL>", "<endpoint>", "<path>",
		"example.com", "localhost", "127.0.0.1", "your-collector",
		"YOUR_", "TODO", "http://",
	}
	for _, bad := range placeholders {
		if strings.Contains(got, bad) {
			t.Errorf("ConsentText() contains placeholder %q; text is:\n%s", bad, got)
		}
	}
}

func TestConsentTextFitsATerminal(t *testing.T) {
	got := ConsentText()
	if got == "" {
		t.Fatalf("ConsentText() is empty, want the telemetry notice")
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("ConsentText() does not end with a newline: %q", got)
	}
	lines := consentLines(t)
	if len(lines) > 12 {
		t.Errorf("ConsentText() has %d lines, want at most 12", len(lines))
	}
	for i, line := range lines {
		if len(line) > 80 {
			t.Errorf("ConsentText() line %d is %d chars, want at most 80: %q",
				i+1, len(line), line)
		}
	}
}

func TestConsentTextIsStable(t *testing.T) {
	first, second := ConsentText(), ConsentText()
	if first == "" {
		t.Fatalf("ConsentText() is empty, want the telemetry notice")
	}
	if first != second {
		t.Errorf("ConsentText() changed between calls:\nfirst:\n%s\nsecond:\n%s",
			first, second)
	}
}
