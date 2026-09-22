package projectdocs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// howItWorksHeading starts the section below the README's first screen.
const howItWorksHeading = "\n## How it works"

// actionsReport is the report that would license an action-count claim.
const actionsReport = "docs/reports/sprint5/actions.md"

// readmeFirstScreen returns the README text above "## How it works": what a
// reader sees before scrolling.
func readmeFirstScreen(t *testing.T) string {
	t.Helper()
	before, _, ok := strings.Cut(readDoc(t, "README.md"), howItWorksHeading)
	if !ok {
		t.Fatalf("README has no %q heading", strings.TrimSpace(howItWorksHeading))
	}
	return before
}

// TestReadmeFirstScreenNamesSetup pins the command that produces a working
// .snapshot, so the first screen is reproducible rather than aspirational.
func TestReadmeFirstScreenNamesSetup(t *testing.T) {
	screen := readmeFirstScreen(t)
	for _, want := range []string{"snapback setup", "cd ~/project", installLine} {
		if !strings.Contains(screen, want) {
			t.Errorf("README first screen (above %q) missing %q", strings.TrimSpace(howItWorksHeading), want)
		}
	}
}

// TestReadmeFirstScreenStatesPrerequisites pins the three things a reader must
// already have: FUSE, the restic CLI and a repository holding a snapshot.
func TestReadmeFirstScreenStatesPrerequisites(t *testing.T) {
	screen := strings.ToLower(readmeFirstScreen(t))
	for _, want := range []string{"fuse", "restic", "at least one snapshot"} {
		if !strings.Contains(screen, want) {
			t.Errorf("README first screen does not state the prerequisite %q", want)
		}
	}
}

// TestReadmeFirstScreenShowsListAndRestore pins the two commands the pitch
// promises: listing .snapshot and copying a file back out of it.
func TestReadmeFirstScreenShowsListAndRestore(t *testing.T) {
	screen := readmeFirstScreen(t)
	for _, want := range []string{"ls .snapshot", "cp .snapshot/latest/"} {
		if !strings.Contains(screen, want) {
			t.Errorf("README first screen missing %q", want)
		}
	}
}

// TestReadmeFirstScreenNotesMacOSDaemon pins the honest platform note: macOS
// runs the daemon in the foreground, the login service is Linux-only for now.
func TestReadmeFirstScreenNotesMacOSDaemon(t *testing.T) {
	screen := readmeFirstScreen(t)
	for _, want := range []string{"macOS", "snapback run", "foreground", "Linux-only"} {
		if !strings.Contains(screen, want) {
			t.Errorf("README first screen missing the macOS daemon note %q", want)
		}
	}
}

var actionCountRe = regexp.MustCompile(`(?i)\b(three|3|four|4) (actions|commands|steps)\b`)

// TestReadmeFirstScreenMakesNoActionCountClaim keeps the first screen from
// counting actions while no measured report backs the number.
func TestReadmeFirstScreenMakesNoActionCountClaim(t *testing.T) {
	if _, err := os.Stat(filepath.Join(repoRoot(t), actionsReport)); err == nil {
		t.Skipf("%s exists, an action-count claim is licensed", actionsReport)
	}
	if m := actionCountRe.FindAllString(readmeFirstScreen(t), -1); len(m) > 0 {
		t.Errorf("README first screen claims %q with no %s to back it", m, actionsReport)
	}
}

// TestReadmeHasNoSkeletonLeftovers keeps the pre-v0.1 skeleton wording out.
func TestReadmeHasNoSkeletonLeftovers(t *testing.T) {
	readme := strings.ToLower(readDoc(t, "README.md"))
	for _, stale := range []string{"only prints build information", "predates v0.1"} {
		if strings.Contains(readme, stale) {
			t.Errorf("README still contains the stale phrase %q", stale)
		}
	}
}
