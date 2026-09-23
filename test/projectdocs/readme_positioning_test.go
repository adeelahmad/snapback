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
