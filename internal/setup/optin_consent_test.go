package setup

import (
	"strings"
	"testing"
)

// wantOptInQuestion pins the exact wording and capital-N default of the
// opt-in question, independent of the OptInQuestion constant, so a change to
// the constant itself is also caught here.
const wantOptInQuestion = "Send anonymous usage counters to a collector you configure? [y/N] "

func TestOptInConsent(t *testing.T) {
	if OptInQuestion != wantOptInQuestion {
		t.Fatalf("OptInQuestion changed: got %q, want %q", OptInQuestion, wantOptInQuestion)
	}

	t.Run("interactive prints consent once immediately above the question", func(t *testing.T) {
		var out strings.Builder
		enabled, asked := AskOptIn(strings.NewReader("y\n"), &out, true)

		if !enabled || !asked {
			t.Fatalf("AskOptIn(interactive=true, \"y\") = %t, %t, want true, true", enabled, asked)
		}

		got := out.String()
		consent := ConsentText()

		if count := strings.Count(got, consent); count != 1 {
			t.Fatalf("AskOptIn output contains ConsentText() %d times, want exactly 1; output:\n%s", count, got)
		}

		consentIdx := strings.Index(got, consent)
		questionIdx := strings.Index(got, OptInQuestion)
		if questionIdx == -1 {
			t.Fatalf("AskOptIn output does not contain OptInQuestion; output:\n%s", got)
		}
		if questionIdx <= consentIdx {
			t.Fatalf("OptInQuestion does not appear after ConsentText(); output:\n%s", got)
		}

		gap := got[consentIdx+len(consent) : questionIdx]
		if gap != "" && gap != "\n" {
			t.Errorf("gap between ConsentText() and OptInQuestion is %q, want \"\" or one blank line (\"\\n\")", gap)
		}

		if rest := got[questionIdx:]; rest != OptInQuestion {
			t.Errorf("text from the question onward is %q, want exactly %q", rest, OptInQuestion)
		}
	})

	t.Run("non-interactive writes neither consent nor question", func(t *testing.T) {
		var out strings.Builder
		enabled, asked := AskOptIn(strings.NewReader("y\n"), &out, false)

		if enabled || asked {
			t.Errorf("AskOptIn(interactive=false) = %t, %t, want false, false", enabled, asked)
		}
		if got := out.String(); got != "" {
			t.Errorf("AskOptIn(interactive=false) wrote %q, want nothing (no consent text, no question)", got)
		}
	})
}
