package mount

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	goFuseModule  = "github.com/hanwen/go-fuse/v2"
	goFuseVersion = "v2.11.0"
)

var exactTag = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// goModRequireLines returns every require entry (single-line or inside a
// require block) whose module path is mod.
func goModRequireLines(src, mod string) []string {
	var out []string
	inBlock := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "require ("):
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		}
		entry := ""
		if inBlock {
			entry = line
		} else if rest, ok := strings.CutPrefix(line, "require "); ok {
			entry = strings.TrimSpace(rest)
		}
		if fields := strings.Fields(entry); len(fields) >= 2 && fields[0] == mod {
			out = append(out, entry)
		}
	}
	return out
}

func TestGoModPinsGoFuseExactTag(t *testing.T) {
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	src := string(data)
	if strings.TrimSpace(src) == "" {
		t.Fatal("go.mod is empty")
	}

	lines := goModRequireLines(src, goFuseModule)
	if len(lines) != 1 {
		t.Fatalf("go.mod require lines for %s = %d (%q), want exactly 1\ngo.mod:\n%s", goFuseModule, len(lines), lines, src)
	}
	entry := lines[0]
	version := strings.Fields(entry)[1]
	if strings.Contains(version, "-0.") {
		t.Errorf("go-fuse version %q is a pseudo-version, want exact tag", version)
	}
	if !exactTag.MatchString(version) {
		t.Errorf("go-fuse version %q does not match %s", version, exactTag)
	}
	if version != goFuseVersion {
		t.Errorf("go-fuse version = %q, want %q", version, goFuseVersion)
	}
	if strings.Contains(entry, "// indirect") {
		t.Errorf("go-fuse require %q is marked indirect, want a direct requirement", entry)
	}
	inReplace := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "replace ("):
			inReplace = true
			continue
		case inReplace && line == ")":
			inReplace = false
			continue
		}
		if (inReplace || strings.HasPrefix(line, "replace ")) && strings.Contains(line, "hanwen/go-fuse") {
			t.Errorf("go.mod has a replace directive naming go-fuse: %q", line)
		}
	}
}

func TestGoSumHasGoFuseHashes(t *testing.T) {
	data, err := os.ReadFile("../../go.sum")
	if err != nil {
		t.Fatalf("read go.sum: %v", err)
	}
	src := string(data)
	if strings.TrimSpace(src) == "" {
		t.Fatal("go.sum is empty")
	}
	for _, want := range []string{
		goFuseModule + " " + goFuseVersion + " h1:",
		goFuseModule + " " + goFuseVersion + "/go.mod h1:",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("go.sum lacks %q", want)
		}
	}
}
