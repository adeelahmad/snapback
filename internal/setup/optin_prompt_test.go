package setup

import (
	"strings"
	"testing"
)

func TestAskOptIn(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		interactive bool
		wantEnabled bool
		wantAsked   bool
	}{
		{name: "y", in: "y\n", interactive: true, wantEnabled: true, wantAsked: true},
		{name: "yes", in: "yes\n", interactive: true, wantEnabled: true, wantAsked: true},
		{name: "uppercase yes", in: "YES\n", interactive: true, wantEnabled: true, wantAsked: true},
		{name: "n", in: "n\n", interactive: true, wantEnabled: false, wantAsked: true},
		{name: "enter", in: "\n", interactive: true, wantEnabled: false, wantAsked: true},
		{name: "eof", in: "", interactive: true, wantEnabled: false, wantAsked: true},
		{name: "not interactive", in: "y\n", interactive: false, wantEnabled: false, wantAsked: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			enabled, asked := AskOptIn(strings.NewReader(tt.in), &out, tt.interactive)
			if enabled != tt.wantEnabled || asked != tt.wantAsked {
				t.Errorf("AskOptIn(%q, %t) = %t, %t, want %t, %t",
					tt.in, tt.interactive, enabled, asked, tt.wantEnabled, tt.wantAsked)
			}
			want := 0
			if tt.wantAsked {
				want = 1
			}
			if got := strings.Count(out.String(), OptInQuestion); got != want {
				t.Errorf("AskOptIn(%q, %t) printed the question %d times, want %d",
					tt.in, tt.interactive, got, want)
			}
			if tt.wantAsked && out.String() != OptInQuestion {
				t.Errorf("AskOptIn(%q, %t) wrote %q, want exactly %q",
					tt.in, tt.interactive, out.String(), OptInQuestion)
			}
		})
	}
}
