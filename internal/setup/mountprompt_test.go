package setup

import (
	"fmt"
	"strings"
	"testing"
)

// questionLines counts the lines of out that carry the question mark, so a
// re-ask is visible and a second unasked question is not.
func questionLines(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "?") {
			n++
		}
	}
	return n
}

func TestAskMountPointDefault(t *testing.T) {
	const def = "/mnt/home-nas"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader("\n"), &out, true, def)
	if err != nil {
		t.Fatalf("AskMountPoint(%q) returned an error: %v", "\n", err)
	}
	if got != def {
		t.Errorf("AskMountPoint(%q) = %q, want the default %q", "\n", got, def)
	}
	if !strings.Contains(out.String(), def) {
		t.Errorf("AskMountPoint(%q) wrote %q, want it to show the default %q", "\n", out.String(), def)
	}
	if n := questionLines(out.String()); n != 1 {
		t.Errorf("AskMountPoint(%q) asked %d questions, want exactly 1: %q", "\n", n, out.String())
	}
	if want := fmt.Sprintf(MountPointQuestion, def); out.String() != want {
		t.Errorf("AskMountPoint(%q) wrote %q, want exactly %q", "\n", out.String(), want)
	}
}

func TestAskMountPointAbsolute(t *testing.T) {
	const def = "/mnt/home-nas"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader("/srv/restore\n"), &out, true, def)
	if err != nil {
		t.Fatalf("AskMountPoint(%q) returned an error: %v", "/srv/restore\n", err)
	}
	if want := "/srv/restore"; got != want {
		t.Errorf("AskMountPoint(%q) = %q, want %q", "/srv/restore\n", got, want)
	}
	if n := questionLines(out.String()); n != 1 {
		t.Errorf("AskMountPoint(%q) asked %d questions, want exactly 1: %q", "/srv/restore\n", n, out.String())
	}
}

func TestAskMountPointReasksOnceOnRelative(t *testing.T) {
	const def = "/mnt/home-nas"
	const in = "relative\n/srv/x\n"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader(in), &out, true, def)
	if err != nil {
		t.Fatalf("AskMountPoint(%q) returned an error: %v", in, err)
	}
	if want := "/srv/x"; got != want {
		t.Errorf("AskMountPoint(%q) = %q, want the second answer %q", in, got, want)
	}
	if want := fmt.Sprintf(MountPointRelative, "relative"); !strings.Contains(out.String(), want) {
		t.Errorf("AskMountPoint(%q) wrote %q, want it to contain %q", in, out.String(), want)
	}
	if n := questionLines(out.String()); n != 2 {
		t.Errorf("AskMountPoint(%q) asked %d questions, want the question and exactly one re-ask: %q", in, n, out.String())
	}
}

func TestAskMountPointFallsBackAfterSecondRelative(t *testing.T) {
	const def = "/mnt/home-nas"
	const in = "relative\nstill/relative\n"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader(in), &out, true, def)
	if err != nil {
		t.Fatalf("AskMountPoint(%q) returned an error: %v", in, err)
	}
	if got != def {
		t.Errorf("AskMountPoint(%q) = %q, want the default %q", in, got, def)
	}
	if want := fmt.Sprintf(MountPointFallback, def); !strings.Contains(out.String(), want) {
		t.Errorf("AskMountPoint(%q) wrote %q, want it to say it fell back: %q", in, out.String(), want)
	}
	if n := questionLines(out.String()); n != 2 {
		t.Errorf("AskMountPoint(%q) asked %d questions, want the question and exactly one re-ask: %q", in, n, out.String())
	}
}

func TestAskMountPointNotInteractive(t *testing.T) {
	const def = "/mnt/home-nas"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader("/srv/restore\n"), &out, false, def)
	if err != nil {
		t.Fatalf("AskMountPoint(interactive=false) returned an error: %v", err)
	}
	if got != def {
		t.Errorf("AskMountPoint(interactive=false) = %q, want the default %q", got, def)
	}
	if out.String() != "" {
		t.Errorf("AskMountPoint(interactive=false) wrote %q, want nothing", out.String())
	}
}

func TestAskMountPointEOF(t *testing.T) {
	const def = "/mnt/home-nas"

	var out strings.Builder
	got, err := AskMountPoint(strings.NewReader(""), &out, true, def)
	if err != nil {
		t.Fatalf("AskMountPoint(EOF) returned an error: %v", err)
	}
	if got != def {
		t.Errorf("AskMountPoint(EOF) = %q, want the default %q", got, def)
	}
}
