package config

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// pinnedDeps are the Sprint 3 requirements go.mod must carry at exact tags.
var pinnedDeps = []struct{ module, version string }{
	{"go.yaml.in/yaml/v3", "v3.0.4"},
	{"go.etcd.io/bbolt", "v1.4.3"},
	{"golang.org/x/sys", "v0.47.0"},
}

var pseudoVersion = regexp.MustCompile(`-0\.\d{14}`)

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../../" + name)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error: %v", name, err)
	}
	return string(data)
}

// requireLines returns every requirement line of a go.mod file, from both
// single-line `require m v` directives and `require ( ... )` blocks, trimmed.
func requireLines(gomod string) []string {
	var lines []string
	inBlock := false
	for _, raw := range strings.Split(gomod, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case inBlock && line == ")":
			inBlock = false
		case inBlock:
			if line != "" && !strings.HasPrefix(line, "//") {
				lines = append(lines, line)
			}
		case line == "require (":
			inBlock = true
		case strings.HasPrefix(line, "require "):
			lines = append(lines, strings.TrimSpace(strings.TrimPrefix(line, "require ")))
		}
	}
	return lines
}

func TestGoModPinsSprint3Deps(t *testing.T) {
	lines := requireLines(readRepoFile(t, "go.mod"))
	for _, dep := range pinnedDeps {
		want := dep.module + " " + dep.version
		count := 0
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[0]+" "+fields[1] != want {
				continue
			}
			count++
			if strings.Contains(line, "// indirect") {
				t.Errorf("go.mod require %q is marked indirect, want direct", line)
			}
		}
		if count != 1 {
			t.Errorf("go.mod require lines for %q = %d, want 1", want, count)
		}
	}
}

func TestGoModNoFloatingVersions(t *testing.T) {
	lines := requireLines(readRepoFile(t, "go.mod"))
	if len(lines) < 3 {
		t.Fatalf("go.mod require lines = %d, want at least 3: %q", len(lines), lines)
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		for _, dep := range pinnedDeps {
			if fields[0] != dep.module {
				continue
			}
			v := fields[1]
			if strings.Contains(v, "latest") || strings.Contains(v, "master") || pseudoVersion.MatchString(v) {
				t.Errorf("go.mod require %q uses a floating version, want tag %s", line, dep.version)
			}
		}
	}
}

func TestGoSumHasPinnedHashes(t *testing.T) {
	gosum := readRepoFile(t, "go.sum")
	var lines []string
	for _, line := range strings.Split(gosum, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		t.Fatal("go.sum has no lines, want hash lines")
	}
	for _, dep := range pinnedDeps {
		for _, prefix := range []string{
			dep.module + " " + dep.version + " h1:",
			dep.module + " " + dep.version + "/go.mod h1:",
		} {
			found := false
			for _, line := range lines {
				if strings.HasPrefix(line, prefix) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("go.sum has no line starting %q, want one", prefix)
			}
		}
	}
}
