package projectdocs

import (
	"strings"
	"testing"
)

const readmePitch = "**Backup is a solved problem. Restore isn't.**"

const readmeLeadParagraph = "snapback is a restore tool for your backups. It supports Restic today; other backends are on the [roadmap](#roadmap). Restic keeps doing the backup, on your schedule and under your retention rules. snapback does the restore, because restoring a file should be as easy as it was in 2008."

func TestReadmePitchIsBackupRestoreSplit(t *testing.T) {
	lead := readmeLead(readDoc(t, "README.md"))
	if !strings.Contains(lead, readmePitch) {
		t.Errorf("README lead missing exact pitch line %q", readmePitch)
	}
}

func TestReadmeLeadParagraphSplitsBackupFromRestore(t *testing.T) {
	lead := readmeLead(readDoc(t, "README.md"))
	if !strings.Contains(lead, readmeLeadParagraph) {
		t.Errorf("README lead missing exact paragraph %q", readmeLeadParagraph)
	}
}

func TestReadmeNoLongerHasOldPitchLine(t *testing.T) {
	readme := readDoc(t, "README.md")
	const oldPitch = "**Restoring a file should be as easy as it was in 2008.**"
	if strings.Contains(readme, oldPitch) {
		t.Errorf("README still contains old pitch line %q, want it replaced", oldPitch)
	}
}

const readmeWhyTableHeader = "| | Backup | Restore |"

func TestReadmeWhyHasTwoProblemsTable(t *testing.T) {
	why := section(readDoc(t, "README.md"), "## Why")
	if !strings.Contains(why, readmeWhyTableHeader) {
		t.Errorf(`README "## Why" section missing table header %q`, readmeWhyTableHeader)
	}
	for _, want := range []string{
		"**Today** | Solved. Restic takes deduplicated, encrypted snapshots whenever your cron job or timer runs it, and forgets and prunes them by your retention rules.",
		"**snapback** | Stays out of it. Keep running Restic as you do now; `snapback snap` asks it for one extra snapshot, and only when you run it.",
	} {
		if !strings.Contains(why, want) {
			t.Errorf(`README "## Why" section missing table row containing %q`, want)
		}
	}
}

func TestReadmeNeverMakesSnapbackTheBackupTool(t *testing.T) {
	why := section(readDoc(t, "README.md"), "## Why")
	for _, banned := range []string{"snapback is a backup tool", "snapback backs up", "snapback backup tool"} {
		if strings.Contains(strings.ToLower(why), strings.ToLower(banned)) {
			t.Errorf(`README "## Why" section contains %q, snapback must never be framed as the backup tool`, banned)
		}
	}
	if !strings.Contains(why, "Stays out of it") {
		t.Errorf(`README "## Why" section missing "Stays out of it", the line that keeps snapback out of the backup half`)
	}
}
