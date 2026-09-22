package projectdocs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	specSHA256    = "6bc35dadf698271b23790db43d2b521859d724081fae39dd1e281726e0fafe28"
	specSize      = 64847
	specFirstLine = "# Snapback — Implementation Specification"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestSpecHashMatchesStage0Readme(t *testing.T) {
	spec := readDoc(t, "SPEC.md")
	if got := sha256Hex(spec); got != specSHA256 {
		t.Fatalf("SPEC.md sha256 = %s, want %s: spec altered; restore from git show e13192d:README.md", got, specSHA256)
	}
}

func TestSpecSizeAndFirstLine(t *testing.T) {
	spec := readDoc(t, "SPEC.md")
	if len(spec) != specSize {
		t.Errorf("SPEC.md size = %d bytes, want %d", len(spec), specSize)
	}
	first, _, _ := strings.Cut(spec, "\n")
	if first != specFirstLine {
		t.Errorf("SPEC.md first line = %q, want %q", first, specFirstLine)
	}
}

// README.md is legitimately absent between T1 and T2; it is then not the spec
// provided the spec has moved to SPEC.md.
func TestReadmeIsNotTheSpec(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if errors.Is(err, fs.ErrNotExist) {
		if _, serr := os.Stat(filepath.Join(root, "SPEC.md")); serr != nil {
			t.Fatalf("README.md absent and SPEC.md missing under %s: %v", root, serr)
		}
		return
	}
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(data)
	if sha256Hex(readme) == specSHA256 {
		t.Errorf("README.md is still the implementation spec (sha256 %s)", specSHA256)
	}
	if strings.Contains(readme, specFirstLine) {
		t.Errorf("README.md contains spec title %q", specFirstLine)
	}
}
