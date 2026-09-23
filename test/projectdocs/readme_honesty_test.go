package projectdocs

import (
	"regexp"
	"strings"
	"testing"
)

var readOnlyRepositoryClaimRe = regexp.MustCompile(`(?i)(will not|won't|never|does not|doesn't) write to (your |the )?restic repositor|only reads restic repositor`)

func TestReadmeMakesNoReadOnlyRepositoryClaim(t *testing.T) {
	readme := readDoc(t, "README.md")
	if !strings.Contains(readme, "Restic") {
		t.Fatal("README.md does not mention Restic")
	}
	if m := readOnlyRepositoryClaimRe.FindString(readme); m != "" {
		t.Errorf("README.md claims a read-only repository: %q", m)
	}
}

func TestReadmeStatesSnapIsOnDemand(t *testing.T) {
	doesntDo := section(readDoc(t, "README.md"), "## What it doesn't do")
	if strings.TrimSpace(doesntDo) == "" {
		t.Fatal(`README "## What it doesn't do" section is missing or empty`)
	}
	if !strings.Contains(doesntDo, "`snapback snap`") {
		t.Errorf(`README "## What it doesn't do" section missing %q`, "`snapback snap`")
	}
	if !strings.Contains(doesntDo, "only when you run it") {
		t.Errorf(`README "## What it doesn't do" section missing %q`, "only when you run it")
	}
	neverRewritesRe := regexp.MustCompile(`(?i)never deletes, prunes or rewrites repository data`)
	if !neverRewritesRe.MatchString(doesntDo) {
		t.Errorf(`README "## What it doesn't do" section does not match %q`, neverRewritesRe.String())
	}
}

func TestReadmeInstallPinsNoReleaseVersion(t *testing.T) {
	install := section(readDoc(t, "README.md"), "## Install")
	if strings.TrimSpace(install) == "" {
		t.Fatal(`README "## Install" section is missing or empty`)
	}
	if !strings.Contains(install, "latest release") {
		t.Errorf(`README "## Install" section missing %q`, "latest release")
	}
	if m := regexp.MustCompile(`\(v\d+\.\d+\.\d+\)`).FindString(install); m != "" {
		t.Errorf(`README "## Install" section pins a release version: %q`, m)
	}
	if m := regexp.MustCompile(`release \(v`).FindString(install); m != "" {
		t.Errorf(`README "## Install" section pins a release version: %q`, m)
	}
}
