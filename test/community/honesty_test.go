package community

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const issueTemplateDir = ".github/ISSUE_TEMPLATE"

var (
	bannedHonestyPhrases = []string{"production-ready", "cross-platform", "static", "finder-integrated"}
	placeholderTokens    = []string{"todo", "tbd", "<fill", "lorem"}
	emailPattern         = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
)

// scannedFiles returns ownedFiles plus every regular file under
// .github/ISSUE_TEMPLATE/, keyed by repo-relative path. It fails the test when
// the set is incomplete or any file is empty, so negative scans can never pass
// vacuously over nothing.
func scannedFiles(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, rel := range ownedFiles {
		files[rel] = readOwned(t, rel)
	}

	root := repoRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, issueTemplateDir))
	if err != nil {
		t.Fatalf("%s: cannot list directory: %v", issueTemplateDir, err)
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		rel := issueTemplateDir + "/" + e.Name()
		if _, seen := files[rel]; seen {
			continue
		}
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("%s: unreadable: %v", rel, err)
		}
		files[rel] = string(body)
	}

	for _, want := range wantOwnedFiles {
		if _, ok := files[want]; !ok {
			t.Fatalf("scanned file set %v is missing %s", sortedKeys(files), want)
		}
	}
	for rel, body := range files {
		if strings.TrimSpace(body) == "" {
			t.Fatalf("%s: empty content; refusing to scan an empty file", rel)
		}
	}
	return files
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestNoBannedHonestyClaims(t *testing.T) {
	files := scannedFiles(t)
	for _, rel := range sortedKeys(files) {
		lower := strings.ToLower(files[rel])
		for _, phrase := range bannedHonestyPhrases {
			if strings.Contains(lower, phrase) {
				t.Errorf("%s: contains banned honesty phrase %q", rel, phrase)
			}
		}
	}
}

func TestNoEmailAddresses(t *testing.T) {
	files := scannedFiles(t)
	for _, rel := range sortedKeys(files) {
		if m := emailPattern.FindAllString(files[rel], -1); len(m) > 0 {
			t.Errorf("%s: contains email address(es) %v", rel, m)
		}
	}
}

func TestNoPlaceholderTokens(t *testing.T) {
	files := scannedFiles(t)
	for _, rel := range sortedKeys(files) {
		lower := strings.ToLower(files[rel])
		for _, token := range placeholderTokens {
			if strings.Contains(lower, token) {
				t.Errorf("%s: contains placeholder token %q", rel, token)
			}
		}
	}
}
