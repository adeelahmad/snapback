package release_test

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"text/template"
)

// archiveNameContract is the S1-03 installer asset-naming contract (s1-03-installer/tasks.md).
const archiveNameContract = `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}{{ if or (eq .Arch "arm") (eq .Arch "mips") (eq .Arch "mipsle") }}_unverified{{ end }}`

var (
	archiveFormatRe   = regexp.MustCompile(`(?m)^\s*(formats:\s*\[\s*['"]?tar\.gz['"]?\s*\]|format:\s*['"]?tar\.gz['"]?\s*$|formats:\s*\n\s*-\s*['"]?tar\.gz['"]?\s*$)`)
	wrapInDirTrueRe   = regexp.MustCompile(`(?m)^\s*wrap_in_directory:\s*['"]?true['"]?\s*$`)
	checksumNameRe    = regexp.MustCompile(`(?m)^\s*name_template:\s*['"]?checksums\.txt['"]?\s*$`)
	sha256AlgorithmRe = regexp.MustCompile(`(?m)^\s*algorithm:\s*sha256\s*$`)
	publisherKeyRe    = regexp.MustCompile(`^(brews|homebrew_casks|nfpms|publishers|snapcrafts|dockers|aurs|scoops|blobs|uploads):`)
	appendModeRe      = regexp.MustCompile(`(?m)^\s*mode:\s*['"]?append['"]?\s*$`)
	disableTrueRe     = regexp.MustCompile(`(?m)^\s*disable:\s*true\s*$`)
)

// topLevelBlock returns the block under a column-0 key, so nested keys of the same name are not matched.
func topLevelBlock(text, key string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, key+":") {
			continue
		}
		out := []string{line}
		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) != "" && indentOf(next) == 0 && !strings.HasPrefix(next, "- ") && !strings.HasPrefix(next, "#") {
				break
			}
			out = append(out, next)
		}
		return strings.Join(out, "\n")
	}
	return ""
}

// yamlScalar unquotes a YAML scalar value (double-quoted with escapes, single-quoted, or bare).
func yamlScalar(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, `"`) {
		if s, err := strconv.Unquote(raw); err == nil {
			return s
		}
		return raw
	}
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") && len(raw) >= 2 {
		return strings.ReplaceAll(raw[1:len(raw)-1], "''", "'")
	}
	return raw
}

func archiveNameTemplate(text string) string {
	block := topLevelBlock(text, "archives")
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimLeft(strings.TrimSpace(line), "- ")
		if strings.HasPrefix(trimmed, "name_template:") {
			return yamlScalar(strings.TrimPrefix(trimmed, "name_template:"))
		}
	}
	return ""
}

func TestGoreleaserArchiveNameTemplate(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "archives")
	if block == "" {
		t.Fatalf("%s has no top-level archives: block", goreleaserFile)
	}
	if got := archiveNameTemplate(text); got != archiveNameContract {
		t.Errorf("archives name_template = %q, want S1-03 contract %q", got, archiveNameContract)
	}
	if !archiveFormatRe.MatchString(block) {
		t.Errorf("archives block does not declare tar.gz as the archive format:\n%s", block)
	}
	if wrapInDirTrueRe.MatchString(block) {
		t.Errorf("archives block sets wrap_in_directory: true; the binary must sit at the archive root")
	}
}

func TestGoreleaserArchiveNamesRenderForAllTargets(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	tmplText := archiveNameTemplate(text)
	if tmplText == "" {
		t.Fatalf("no archives name_template found in %s", goreleaserFile)
	}
	tmpl, err := template.New("archive").Parse(tmplText)
	if err != nil {
		t.Fatalf("parse name_template %q: %v", tmplText, err)
	}
	targets := [][2]string{
		{"linux", "amd64"}, {"linux", "arm64"}, {"darwin", "amd64"}, {"darwin", "arm64"},
		{"linux", "arm"}, {"linux", "mips"}, {"linux", "mipsle"},
	}
	want := []string{
		"snapback_linux_amd64.tar.gz",
		"snapback_linux_arm64.tar.gz",
		"snapback_darwin_amd64.tar.gz",
		"snapback_darwin_arm64.tar.gz",
		"snapback_linux_arm_unverified.tar.gz",
		"snapback_linux_mips_unverified.tar.gz",
		"snapback_linux_mipsle_unverified.tar.gz",
	}
	for i, tg := range targets {
		var buf bytes.Buffer
		data := map[string]string{"ProjectName": "snapback", "Os": tg[0], "Arch": tg[1]}
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("execute name_template for %s/%s: %v", tg[0], tg[1], err)
		}
		if got := buf.String() + ".tar.gz"; got != want[i] {
			t.Errorf("asset for %s/%s = %q, want %q", tg[0], tg[1], got, want[i])
		}
	}
}

func TestGoreleaserChecksumSha256(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "checksum")
	if block == "" {
		t.Fatalf("%s has no top-level checksum: block", goreleaserFile)
	}
	if !checksumNameRe.MatchString(block) {
		t.Errorf("checksum block lacks name_template: checksums.txt:\n%s", block)
	}
	if !sha256AlgorithmRe.MatchString(block) {
		t.Errorf("checksum block lacks algorithm: sha256:\n%s", block)
	}
}

func TestGoreleaserSignsChecksumKeyless(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "signs")
	if block == "" {
		t.Fatalf("%s has no top-level signs: block", goreleaserFile)
	}
	for _, want := range []string{"cmd: cosign", "artifacts: checksum", "sign-blob", "--output-signature"} {
		if !strings.Contains(block, want) {
			t.Errorf("signs block missing %q", want)
		}
	}
	for _, banned := range []string{"--key", "COSIGN_PRIVATE_KEY", "COSIGN_PASSWORD"} {
		if strings.Contains(text, banned) {
			t.Errorf("%s contains %q; signing must be keyless", goreleaserFile, banned)
		}
	}
}

func TestGoreleaserReleaseNotesLabelUnverified(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "release")
	if block == "" {
		t.Fatalf("%s has no top-level release: block", goreleaserFile)
	}
	notes := strings.ToLower(yamlBlock(block, "footer") + "\n" + yamlBlock(block, "header"))
	for _, want := range []string{"unverified", "linux/arm", "linux/mips", "linux/mipsle"} {
		if !strings.Contains(notes, want) {
			t.Errorf("release footer/header does not mention %q", want)
		}
	}
}

func TestGoreleaserNoPackagePublishers(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	if !strings.Contains(text, "archives:") {
		t.Fatalf("%s has no archives: section; release config is incomplete", goreleaserFile)
	}
	for i, line := range strings.Split(text, "\n") {
		if publisherKeyRe.MatchString(line) {
			t.Errorf("line %d declares an out-of-scope stage-7 channel: %q", i+1, line)
		}
	}
}

func TestGoreleaserReleaseAppendsToExisting(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	release := topLevelBlock(text, "release")
	if !appendModeRe.MatchString(release) {
		t.Errorf("release: block lacks mode: append:\n%s", release)
	}
	changelog := topLevelBlock(text, "changelog")
	if !disableTrueRe.MatchString(changelog) {
		t.Errorf("changelog: block lacks disable: true:\n%s", changelog)
	}
}

func TestGoreleaserHonestyWords(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	if !strings.Contains(text, "release:") {
		t.Fatalf("%s has no release: section; release notes are not yet configured", goreleaserFile)
	}
	lower := strings.ToLower(text)
	for _, word := range []string{"production-ready", "cross-platform", "static", "finder-integrated"} {
		if strings.Contains(lower, word) {
			t.Errorf("%s contains honesty-gate word %q", goreleaserFile, word)
		}
	}
}

// universalReplaceFalseRe pins replace: false so the per-arch darwin archives keep shipping.
var universalReplaceFalseRe = regexp.MustCompile(`(?m)^\s*replace:\s*['"]?false['"]?\s*$`)

// archiveNameTemplates returns every name_template declared under the top-level archives: key.
func archiveNameTemplates(text string) []string {
	var out []string
	for _, line := range strings.Split(topLevelBlock(text, "archives"), "\n") {
		trimmed := strings.TrimLeft(strings.TrimSpace(line), "- ")
		if strings.HasPrefix(trimmed, "name_template:") {
			out = append(out, yamlScalar(strings.TrimPrefix(trimmed, "name_template:")))
		}
	}
	return out
}

func TestGoreleaserDarwinUniversalBinary(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	block := topLevelBlock(text, "universal_binaries")
	if block == "" {
		t.Fatalf("%s has no top-level universal_binaries: block", goreleaserFile)
	}
	if !universalReplaceFalseRe.MatchString(block) {
		t.Errorf("universal_binaries block lacks replace: false; per-arch darwin builds must survive:\n%s", block)
	}
	var name string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimLeft(strings.TrimSpace(line), "- ")
		if strings.HasPrefix(trimmed, "name_template:") {
			name = yamlScalar(strings.TrimPrefix(trimmed, "name_template:"))
			break
		}
	}
	if name != "snapback" {
		t.Errorf("universal_binaries name_template = %q, want %q", name, "snapback")
	}
	if !strings.Contains(block, "snapback") {
		t.Errorf("universal_binaries block does not reference the snapback build:\n%s", block)
	}
}

func TestGoreleaserShipsDarwinUniversalArchive(t *testing.T) {
	text := readRepoFile(t, goreleaserFile)
	const want = "snapback_darwin_universal.tar.gz"
	templates := archiveNameTemplates(text)
	if len(templates) < 2 {
		t.Fatalf("archives declares %d name_template(s); want a darwin universal entry too", len(templates))
	}
	for _, tmplText := range templates {
		tmpl, err := template.New("archive").Parse(tmplText)
		if err != nil {
			t.Fatalf("parse name_template %q: %v", tmplText, err)
		}
		var buf bytes.Buffer
		data := map[string]string{"ProjectName": "snapback", "Os": "darwin", "Arch": "all"}
		if err := tmpl.Execute(&buf, data); err != nil {
			t.Fatalf("execute name_template %q: %v", tmplText, err)
		}
		if buf.String()+".tar.gz" == want {
			return
		}
	}
	t.Errorf("no archives name_template in %v yields %q", templates, want)
}

func TestGoreleaserCheck(t *testing.T) {
	bin, err := exec.LookPath("goreleaser")
	if err != nil {
		t.Skip("goreleaser not on PATH")
	}
	cmd := exec.Command(bin, "check")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goreleaser check failed: %v\n%s", err, out)
	}
}
