package release_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

// S5-27/T3: the QEMU smoke workflow (.github/workflows/qemu.yml, S5-27/T1-T2) runs
// linux/arm, linux/mips and linux/mipsle binaries under qemu-user-static on every CI
// run, so those targets are no longer unverified. Nothing the release ships, and no
// CI job name, may still call them that.

const (
	ciWorkflowFile = ".github/workflows/ci.yml"
	readmeFile     = "README.md"
)

// smokeTestedArches are the targets the QEMU smoke workflow covers.
var smokeTestedArches = []string{"arm", "mips", "mipsle"}

func TestUnverifiedSuffixGoneFromArchiveNames(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	templates := archiveNameTemplates(text)
	if len(templates) == 0 {
		t.Fatalf("%s declares no archives name_template", goreleaserFile)
	}
	for _, tmplText := range templates {
		if strings.Contains(strings.ToLower(tmplText), "unverified") {
			t.Errorf("archives name_template %q still marks artifacts unverified; QEMU smoke-tests arm/mips/mipsle on every CI run", tmplText)
		}
	}
}

func TestUnverifiedSuffixGoneFromRenderedAssets(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	tmplText := archiveNameTemplate(text)
	if tmplText == "" {
		t.Fatalf("no archives name_template found in %s", goreleaserFile)
	}
	tmpl, err := template.New("archive").Parse(tmplText)
	if err != nil {
		t.Fatalf("parse name_template %q: %v", tmplText, err)
	}
	for _, arch := range smokeTestedArches {
		var buf bytes.Buffer
		data := map[string]string{"ProjectName": "snapback", "Os": "linux", "Arch": arch}
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("execute name_template for linux/%s: %v", arch, err)
		}
		want := "snapback_linux_" + arch + ".tar.gz"
		if got := buf.String() + ".tar.gz"; got != want {
			t.Errorf("asset for linux/%s = %q, want %q", arch, got, want)
		}
	}
}

func TestUnverifiedLabelGoneFromCIJobNames(t *testing.T) {
	text := readRepoFile(t, ciWorkflowFile)
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(strings.ToLower(line), "unverified") {
			t.Errorf("%s:%d still labels a build job unverified: %q", ciWorkflowFile, i+1, strings.TrimSpace(line))
		}
	}
}

func TestUnverifiedWordingGoneFromReadme(t *testing.T) {
	text := readRepoFile(t, readmeFile)
	for i, line := range strings.Split(text, "\n") {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "unverified") {
			continue
		}
		for _, arch := range smokeTestedArches {
			if strings.Contains(lower, "linux/"+arch) || strings.Contains(lower, "_"+arch+"_unverified") {
				t.Errorf("%s:%d still calls linux/%s unverified: %q", readmeFile, i+1, arch, strings.TrimSpace(line))
				break
			}
		}
	}
}

func TestReadmeStatesQEMUSmokeCoverage(t *testing.T) {
	text := readRepoFile(t, readmeFile)
	var found string
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "QEMU") {
			found = line
			break
		}
	}
	if found == "" {
		t.Fatalf("%s does not mention QEMU; it must state that linux/arm, linux/mips and linux/mipsle are smoke-tested under QEMU on every CI run", readmeFile)
	}
	lower := strings.ToLower(found)
	for _, want := range []string{"smoke", "every ci run"} {
		if !strings.Contains(lower, want) {
			t.Errorf("%s QEMU sentence %q does not say %q", readmeFile, strings.TrimSpace(found), want)
		}
	}
	for _, arch := range smokeTestedArches {
		if !strings.Contains(lower, "linux/"+arch) {
			t.Errorf("%s QEMU sentence %q does not name linux/%s", readmeFile, strings.TrimSpace(found), arch)
		}
	}
}
