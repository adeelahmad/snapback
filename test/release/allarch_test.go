package release_test

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

// allArchiveName is the S5-26/T2 contract: one extra, version-free tarball holding every binary.
const allArchiveName = "snapback_all.tar.gz"

// splitListEntries splits a YAML block body into its "  - " entries, dropping the key line.
func splitListEntries(block string) []string {
	var entries []string
	var cur []string
	for _, line := range strings.Split(block, "\n") {
		if strings.HasPrefix(line, "  - ") {
			if len(cur) > 0 {
				entries = append(entries, strings.Join(cur, "\n"))
			}
			cur = []string{line}
			continue
		}
		if len(cur) > 0 {
			cur = append(cur, line)
		}
	}
	if len(cur) > 0 {
		entries = append(entries, strings.Join(cur, "\n"))
	}
	return entries
}

// entryScalar returns the unquoted value of key inside a single list entry.
func entryScalar(entry, key string) string {
	for _, line := range strings.Split(entry, "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		rest, ok := strings.CutPrefix(trimmed, key+":")
		if !ok {
			continue
		}
		return yamlScalar(rest)
	}
	return ""
}

// entryIDs returns the "ids:" list of a single archive entry.
func entryIDs(entry string) []string {
	lines := strings.Split(entry, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "ids:" {
			continue
		}
		var out []string
		for _, next := range lines[i+1:] {
			trimmed := strings.TrimSpace(next)
			if !strings.HasPrefix(trimmed, "- ") {
				break
			}
			out = append(out, yamlScalar(strings.TrimPrefix(trimmed, "- ")))
		}
		return out
	}
	return nil
}

// renderArchiveName renders a goreleaser archive name_template for one target.
func renderArchiveName(t *testing.T, nameTemplate, goos, goarch string) string {
	t.Helper()
	tpl, err := template.New("name").Parse(nameTemplate)
	if err != nil {
		t.Fatalf("parse name_template %q: %v", nameTemplate, err)
	}
	var buf bytes.Buffer
	data := struct{ ProjectName, Os, Arch string }{"snapback", goos, goarch}
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("render name_template %q: %v", nameTemplate, err)
	}
	return buf.String()
}

// allArchiveEntry returns the single archive entry whose rendered name is allArchiveName.
func allArchiveEntry(t *testing.T, cfg string) (string, int) {
	t.Helper()
	var match string
	count := 0
	for _, entry := range splitListEntries(topLevelBlock(cfg, "archives")) {
		name := entryScalar(entry, "name_template")
		if name == "" {
			t.Fatalf("archive entry without name_template:\n%s", entry)
		}
		if renderArchiveName(t, name, "linux", "amd64")+".tar.gz" != allArchiveName {
			continue
		}
		match = entry
		count++
	}
	return match, count
}

// TestAllArchitectureArchiveIsDeclaredOnce pins exactly one tar.gz archive named snapback_all.tar.gz.
func TestAllArchitectureArchiveIsDeclaredOnce(t *testing.T) {
	cfg := readRepoFile(t, ".goreleaser.yaml")
	entry, count := allArchiveEntry(t, cfg)
	if count != 1 {
		t.Fatalf("archives rendering to %s = %d, want exactly 1", allArchiveName, count)
	}
	if formats := entryScalar(entry, "formats"); formats != "[tar.gz]" {
		t.Errorf("all-architecture archive formats = %q, want [tar.gz]", formats)
	}
	if meta := entryScalar(entry, "meta"); meta != "true" {
		t.Errorf("all-architecture archive meta = %q, want true", meta)
	}
}

// TestAllArchitectureArchiveCoversEveryBuild pins that the all-architecture archive names every build id.
func TestAllArchitectureArchiveCoversEveryBuild(t *testing.T) {
	cfg := readRepoFile(t, ".goreleaser.yaml")

	var wantIDs []string
	for _, entry := range splitListEntries(topLevelBlock(cfg, "builds")) {
		if id := entryScalar(entry, "id"); id != "" {
			wantIDs = append(wantIDs, id)
		}
	}
	for _, entry := range splitListEntries(topLevelBlock(cfg, "universal_binaries")) {
		if id := entryScalar(entry, "id"); id != "" {
			wantIDs = append(wantIDs, id)
		}
	}
	if len(wantIDs) == 0 {
		t.Fatalf("no build ids declared in .goreleaser.yaml")
	}

	entry, _ := allArchiveEntry(t, cfg)
	got := entryIDs(entry)
	for _, want := range wantIDs {
		found := false
		for _, id := range got {
			if id == want {
				found = true
			}
		}
		if !found {
			t.Errorf("all-architecture archive ids %v do not cover build id %q", got, want)
		}
	}
}
