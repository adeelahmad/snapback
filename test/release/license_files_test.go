package release_test

import (
	"strings"
	"testing"
)

// archiveFiles parses the top-level archives: block of .goreleaser.yaml and returns, per archive
// entry, the id and the literal items of its files: list.
func archiveFiles(t *testing.T, yaml string) map[string][]string {
	t.Helper()
	block := topLevelBlock(yaml, "archives")
	if strings.TrimSpace(block) == "" {
		t.Fatalf("no top-level archives: block in .goreleaser.yaml")
	}
	out := map[string][]string{}
	id := ""
	inFiles := false
	filesIndent := 0
	for _, line := range strings.Split(block, "\n")[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := indentOf(line)
		if indent == 2 && strings.HasPrefix(trimmed, "- ") {
			id = strings.TrimSpace(strings.TrimPrefix(trimmed, "- id:"))
			if id == "" {
				t.Fatalf("archive entry without an id: %q", line)
			}
			out[id] = nil
			inFiles = false
			continue
		}
		if inFiles {
			if indent > filesIndent && strings.HasPrefix(trimmed, "- ") {
				out[id] = append(out[id], yamlScalar(strings.TrimPrefix(trimmed, "- ")))
				continue
			}
			inFiles = false
		}
		if trimmed == "files:" {
			if id == "" {
				t.Fatalf("files: outside an archive entry")
			}
			inFiles = true
			filesIndent = indent
		}
	}
	return out
}

// TestArchivesShipLicenseAndNotice pins S5-28: every release tarball carries LICENSE and NOTICE
// beside the binary, so the redistributed archive is legally complete on its own.
func TestArchivesShipLicenseAndNotice(t *testing.T) {
	yaml := readRepoFile(t, ".goreleaser.yaml")
	entries := archiveFiles(t, yaml)
	if len(entries) < 2 {
		t.Fatalf("want at least the snapback and snapback_universal archives, got %d: %v", len(entries), entries)
	}
	for _, id := range []string{"snapback", "snapback_universal"} {
		if _, ok := entries[id]; !ok {
			t.Errorf("archive %q is missing from .goreleaser.yaml", id)
		}
	}
	for id, files := range entries {
		for _, want := range []string{"LICENSE", "NOTICE"} {
			if !contains(files, want) {
				t.Errorf("archive %q files: %v does not include %s", id, files, want)
			}
		}
		for _, f := range files {
			if strings.HasPrefix(f, "none") {
				t.Errorf("archive %q files: still excludes everything with %q", id, f)
			}
		}
	}
}

// TestLicenseAndNoticeExistAtRepoRoot keeps the archive contract honest: the packed paths are real.
func TestLicenseAndNoticeExistAtRepoRoot(t *testing.T) {
	for _, rel := range []string{"LICENSE", "NOTICE"} {
		if strings.TrimSpace(readRepoFile(t, rel)) == "" {
			t.Errorf("%s is empty", rel)
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
