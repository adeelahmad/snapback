package community

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const wantModule = "github.com/adeelahmad/snapback"

const minOwnedBytes = 200

var wantOwnedFiles = []string{
	"CONTRIBUTING.md",
	"SECURITY.md",
	"CODE_OF_CONDUCT.md",
	".github/ISSUE_TEMPLATE/bug_report.md",
	".github/ISSUE_TEMPLATE/feature_request.md",
	".github/pull_request_template.md",
}

func TestRepoRootHasGoMod(t *testing.T) {
	root := repoRoot(t)

	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("repoRoot(t) = %q has no readable go.mod: %v", root, err)
	}
	defer func() { _ = f.Close() }()

	var module string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if rest, ok := strings.CutPrefix(line, "module "); ok {
			module = strings.TrimSpace(rest)
			break
		}
	}
	if module != wantModule {
		t.Fatalf("go.mod under %q declares module %q, want %q", root, module, wantModule)
	}
}

func TestOwnedFilesExistAndNonTrivial(t *testing.T) {
	if len(ownedFiles) != len(wantOwnedFiles) {
		t.Fatalf("ownedFiles has %d entries %v, want %d %v", len(ownedFiles), ownedFiles, len(wantOwnedFiles), wantOwnedFiles)
	}
	for i, want := range wantOwnedFiles {
		if ownedFiles[i] != want {
			t.Errorf("ownedFiles[%d] = %q, want %q", i, ownedFiles[i], want)
		}
	}

	root := repoRoot(t)
	for _, rel := range ownedFiles {
		t.Run(rel, func(t *testing.T) {
			path := filepath.Join(root, rel)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("%s: missing: %v", rel, err)
			}
			if !info.Mode().IsRegular() {
				t.Fatalf("%s: not a regular file (mode %v)", rel, info.Mode())
			}
			body := strings.TrimSpace(readOwned(t, rel))
			if len(body) < minOwnedBytes {
				t.Fatalf("%s: %d bytes of trimmed content, want at least %d", rel, len(body), minOwnedBytes)
			}
		})
	}
}
