package ci_test

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const golangciConfigPath = ".golangci.yml"

var pinnedRefRe = regexp.MustCompile(`@(v\d[\w.\-]*|[0-9a-f]{40})$`)

// enabledLinters returns the union of every `enable:` list in a golangci config
// (block or inline form), covering both v1 `linters.enable` and v2 `formatters.enable`.
func enabledLinters(text string) map[string]bool {
	lines := strings.Split(text, "\n")
	enabled := map[string]bool{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if !strings.HasPrefix(trimmed, "enable:") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "enable:"))
		if strings.HasPrefix(rest, "[") {
			for _, name := range strings.Split(strings.Trim(rest, "[]"), ",") {
				enabled[strings.Trim(strings.TrimSpace(name), `"'`)] = true
			}
			continue
		}
		base := indentOf(line)
		for _, l := range lines[i+1:] {
			t := strings.TrimSpace(strings.TrimRight(l, "\r"))
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			if indentOf(l) < base || (indentOf(l) == base && !strings.HasPrefix(t, "- ")) {
				break
			}
			if !strings.HasPrefix(t, "- ") {
				break
			}
			name := strings.TrimSpace(strings.TrimPrefix(t, "- "))
			if idx := strings.Index(name, "#"); idx >= 0 {
				name = strings.TrimSpace(name[:idx])
			}
			enabled[strings.Trim(name, `"'`)] = true
		}
	}
	return enabled
}

func TestGolangciConfigEnablesRequiredLinters(t *testing.T) {
	enabled := enabledLinters(readRepoFile(t, golangciConfigPath))
	for _, want := range []string{"staticcheck", "govet", "errcheck", "goimports"} {
		if !enabled[want] {
			t.Errorf("%s: %q not in any enable list (got %v)", golangciConfigPath, want, enabled)
		}
	}
}

func TestCIGolangciLintActionPinnedVersion(t *testing.T) {
	step := stepContaining(readCI(t), "golangci/golangci-lint-action")
	if step == "" {
		t.Fatalf("%s: no step uses golangci/golangci-lint-action", ciWorkflowPath)
	}
	var uses, version string
	for _, line := range strings.Split(step, "\n") {
		kv := strings.TrimPrefix(strings.TrimSpace(line), "- ")
		if v, ok := strings.CutPrefix(kv, "uses:"); ok {
			uses = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		if v, ok := strings.CutPrefix(kv, "version:"); ok {
			version = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	if !pinnedRefRe.MatchString(uses) {
		t.Errorf("golangci-lint-action not pinned to v<digit> or a 40-hex SHA: %q", uses)
	}
	if version == "" {
		t.Errorf("golangci-lint-action step has no version: input:\n%s", step)
	} else if strings.EqualFold(version, "latest") {
		t.Errorf("golangci-lint-action version must not be latest")
	}
}

func TestCIActionlint(t *testing.T) {
	bin, err := exec.LookPath("actionlint")
	if err != nil {
		t.Skip("actionlint not on PATH")
	}
	cmd := exec.Command(bin, filepath.FromSlash(ciWorkflowPath))
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("actionlint %s failed: %v\n%s", ciWorkflowPath, err, out)
	}
	if len(strings.TrimSpace(string(out))) != 0 {
		t.Errorf("actionlint %s produced output:\n%s", ciWorkflowPath, out)
	}
}
