package ci_test

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

const crossCompileJob = "cross-compile"

var (
	verifiedTargets   = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"}
	unverifiedTargets = []string{"linux/arm", "linux/mips", "linux/mipsle"}
)

type matrixEntry struct {
	goos, goarch string
	fields       map[string]string
}

func (e matrixEntry) target() string { return e.goos + "/" + e.goarch }

// jobBlock returns the text of jobs.<name> in a workflow, or "" if absent.
func jobBlock(text, name string) string {
	lines := strings.Split(text, "\n")
	inJobs := false
	for i, line := range lines {
		if indentOf(line) == 0 && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			inJobs = strings.HasPrefix(strings.TrimRight(line, " \r"), "jobs:")
			continue
		}
		if !inJobs {
			continue
		}
		key := strings.TrimSuffix(strings.TrimSpace(strings.TrimRight(line, "\r")), ":")
		if strings.Trim(key, `"'`) != name {
			continue
		}
		base := indentOf(line)
		end := i + 1
		for end < len(lines) {
			l := lines[end]
			if strings.TrimSpace(l) != "" && indentOf(l) <= base {
				break
			}
			end++
		}
		return strings.Join(lines[i:end], "\n")
	}
	return ""
}

func crossCompileBlock(t *testing.T) string {
	t.Helper()
	block := jobBlock(readCI(t), crossCompileJob)
	if block == "" {
		t.Fatalf("no jobs.%s block in %s", crossCompileJob, ciWorkflowPath)
	}
	return block
}

var kvRe = regexp.MustCompile(`^([A-Za-z0-9_-]+)\s*:\s*(.*)$`)

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " #"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return strings.Trim(s, `"'`)
}

// matrixEntries extracts every list item carrying goos/goarch keys from the job text.
// Supports block-style items (`- goos: linux` + sibling keys) and flow maps (`- {goos: linux, ...}`).
func matrixEntries(block string) []matrixEntry {
	lines := strings.Split(block, "\n")
	var entries []matrixEntry
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		fields := map[string]string{}
		if strings.HasPrefix(item, "{") {
			for _, part := range strings.Split(strings.Trim(item, "{} "), ",") {
				if m := kvRe.FindStringSubmatch(strings.TrimSpace(part)); m != nil {
					fields[m[1]] = unquote(m[2])
				}
			}
		} else {
			m := kvRe.FindStringSubmatch(item)
			if m == nil {
				continue
			}
			fields[m[1]] = unquote(m[2])
			itemIndent := indentOf(line) + 2
			for _, l := range lines[i+1:] {
				if strings.TrimSpace(l) == "" {
					continue
				}
				if indentOf(l) != itemIndent || strings.HasPrefix(strings.TrimSpace(l), "- ") {
					break
				}
				if m := kvRe.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
					fields[m[1]] = unquote(m[2])
				}
			}
		}
		if fields["goos"] == "" || fields["goarch"] == "" {
			continue
		}
		entries = append(entries, matrixEntry{goos: fields["goos"], goarch: fields["goarch"], fields: fields})
	}
	return entries
}

func TestMatrixHasAllSevenTargets(t *testing.T) {
	entries := matrixEntries(crossCompileBlock(t))
	got := make([]string, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.target())
	}
	want := append(append([]string{}, verifiedTargets...), unverifiedTargets...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("cross-compile matrix targets = %v, want exactly %v", got, want)
	}
}

func TestMatrixUnverifiedLabelExact(t *testing.T) {
	entries := matrixEntries(crossCompileBlock(t))
	if len(entries) == 0 {
		t.Fatalf("no goos/goarch entries in jobs.%s matrix", crossCompileJob)
	}
	isUnverified := map[string]bool{}
	for _, tgt := range unverifiedTargets {
		isUnverified[tgt] = true
	}
	labelled := map[string]bool{}
	for _, e := range entries {
		hasLabel := false
		for k, v := range e.fields {
			if k != "goos" && k != "goarch" && strings.Contains(strings.ToLower(v), "unverified") {
				hasLabel = true
			}
		}
		if hasLabel {
			labelled[e.target()] = true
		}
		if hasLabel && !isUnverified[e.target()] {
			t.Errorf("verified target %s is labelled unverified: %v", e.target(), e.fields)
		}
	}
	for _, tgt := range unverifiedTargets {
		if !labelled[tgt] {
			t.Errorf("unverified target %s lacks an `unverified` label", tgt)
		}
	}
}

func TestMatrixJobNameShowsLabel(t *testing.T) {
	block := crossCompileBlock(t)
	var name string
	for _, line := range strings.Split(block, "\n")[1:] {
		if m := regexp.MustCompile(`^name:\s*(.*)$`).FindStringSubmatch(strings.TrimSpace(line)); m != nil && !strings.HasPrefix(strings.TrimSpace(line), "- ") {
			name = m[1]
			break
		}
	}
	if name == "" {
		t.Fatalf("jobs.%s has no name:", crossCompileJob)
	}
	if !regexp.MustCompile(`\$\{\{\s*matrix\.[A-Za-z0-9_-]+\s*\}\}`).MatchString(name) {
		t.Fatalf("jobs.%s name %q does not interpolate a matrix field", crossCompileJob, name)
	}
	labelKeys := map[string]bool{}
	for _, e := range matrixEntries(block) {
		for k, v := range e.fields {
			if strings.Contains(strings.ToLower(v), "unverified") {
				labelKeys[k] = true
			}
		}
	}
	found := false
	for k := range labelKeys {
		if regexp.MustCompile(`\$\{\{[^}]*matrix\.` + regexp.QuoteMeta(k) + `\b[^}]*\}\}`).MatchString(name) {
			found = true
		}
	}
	if !found {
		t.Fatalf("jobs.%s name %q does not interpolate the matrix field carrying `unverified` (fields: %v)", crossCompileJob, name, labelKeys)
	}
}

func TestMatrixLinuxCgoDisabled(t *testing.T) {
	block := crossCompileBlock(t)
	if !regexp.MustCompile(`CGO_ENABLED(:\s*|=)['"]?0['"]?`).MatchString(block) {
		t.Fatalf("jobs.%s does not set CGO_ENABLED: 0", crossCompileJob)
	}
}

func TestMatrixRunsFileOnLinuxBinaries(t *testing.T) {
	block := crossCompileBlock(t)
	fileRe := regexp.MustCompile(`(?m)(run:\s*|^\s+|&&\s*|;\s*)file\s+\S`)
	for _, step := range stepsContaining(block, "file ") {
		if !fileRe.MatchString(step) {
			continue
		}
		if regexp.MustCompile(`if:.*linux`).MatchString(step) {
			return
		}
	}
	t.Fatalf("jobs.%s has no linux-guarded step running `file` on the built binary", crossCompileJob)
}
