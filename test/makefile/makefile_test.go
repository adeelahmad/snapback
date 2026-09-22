// Package makefile_test pins the repo-root Makefile: its targets, the build
// and install recipes, and the ci target that runs every standards gate. The
// tests only use `make -n`, so nothing is built or installed.
package makefile_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	repoRoot      = "../.."
	makeTimeout   = 30 * time.Second
	versionPkg    = "github.com/adeelahmad/snapback/internal/version"
	installPrefix = "/tmp/x"
)

// readMakefile returns the repo-root Makefile, failing the test when it is
// absent.
func readMakefile(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, "Makefile"))
	if err != nil {
		t.Fatalf("repo-root Makefile missing: %v", err)
	}
	return string(data)
}

// makeDryRun runs `make -n <args>` at the repo root and returns its output.
// The Makefile check comes first so a missing Makefile fails rather than
// skips when make is not installed.
func makeDryRun(t *testing.T, args ...string) string {
	t.Helper()
	readMakefile(t)
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("prerequisite missing: make not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), makeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "make", append([]string{"-n", "-C", repoRoot}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make -n %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// joinContinuations folds backslash-newline continuations into single lines.
func joinContinuations(s string) string {
	return strings.ReplaceAll(s, "\\\n", " ")
}

var ruleRE = regexp.MustCompile(`^([A-Za-z0-9_.-][A-Za-z0-9_. -]*):([^=]|$)`)

// rules returns the Makefile's explicit rule target names in order of
// appearance, splitting multi-target rules such as `vet lint:`.
func rules(makefile string) []string {
	var names []string
	for _, line := range strings.Split(joinContinuations(makefile), "\n") {
		if m := ruleRE.FindStringSubmatch(line); m != nil {
			names = append(names, strings.Fields(m[1])...)
		}
	}
	return names
}

// phony returns the set of targets declared in any .PHONY line.
func phony(makefile string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(joinContinuations(makefile), "\n") {
		rest, ok := strings.CutPrefix(line, ".PHONY:")
		if !ok {
			continue
		}
		for _, name := range strings.Fields(rest) {
			set[name] = true
		}
	}
	return set
}

var wantTargets = []string{
	"help", "build", "install", "uninstall", "test", "cover", "lint", "fmt",
	"fmt-check", "vet", "vuln", "docs", "release-check", "ci", "clean",
}

func TestMakefileTargets(t *testing.T) {
	mf := readMakefile(t)
	got := rules(mf)
	defined := map[string]bool{}
	for _, name := range got {
		defined[name] = true
	}
	phonySet := phony(mf)
	for _, want := range wantTargets {
		if !defined[want] {
			t.Errorf("Makefile rules = %v, want target %q", got, want)
		}
		if !phonySet[want] {
			t.Errorf("Makefile .PHONY lacks %q", want)
		}
	}
}

func TestMakefileHelpIsDefaultTarget(t *testing.T) {
	mf := readMakefile(t)
	var first string
	for _, name := range rules(mf) {
		if !strings.HasPrefix(name, ".") {
			first = name
			break
		}
	}
	if first != "help" {
		t.Errorf("Makefile first target = %q, want %q", first, "help")
	}
}

func TestMakeBuildDryRun(t *testing.T) {
	out := makeDryRun(t, "build")
	var line string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "go build") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("make -n build printed no go build line:\n%s", out)
	}
	for _, want := range []string{
		"CGO_ENABLED=0",
		"-o bin/snapback",
		"./cmd/snapback",
		"-ldflags",
		versionPkg + ".Version=",
		versionPkg + ".Commit=",
		versionPkg + ".Target=",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("make -n build go build line = %q, want it to contain %q", line, want)
		}
	}
}

var installModeRE = regexp.MustCompile(`\binstall\b.*-m\s*0?755\b`)

func TestMakeInstallDryRun(t *testing.T) {
	tests := []struct {
		name string
		args []string
		dest string
	}{
		{"prefix", []string{"install", "PREFIX=" + installPrefix}, installPrefix + "/bin/snapback"},
		{"destdir", []string{"install", "PREFIX=" + installPrefix, "DESTDIR=/tmp/stage"}, "/tmp/stage" + installPrefix + "/bin/snapback"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := makeDryRun(t, tc.args...)
			found := false
			for _, l := range strings.Split(out, "\n") {
				if installModeRE.MatchString(l) && strings.Contains(l, "bin/snapback") && strings.Contains(l, tc.dest) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("make -n %s printed no `install -m 0755 bin/snapback %s` line:\n%s", strings.Join(tc.args, " "), tc.dest, out)
			}
		})
	}
}

func TestMakeUninstallDryRun(t *testing.T) {
	out := makeDryRun(t, "uninstall", "PREFIX="+installPrefix)
	want := installPrefix + "/bin/snapback"
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "rm ") && strings.Contains(l, want) {
			return
		}
	}
	t.Errorf("make -n uninstall PREFIX=%s printed no `rm ... %s` line:\n%s", installPrefix, want, out)
}

func TestMakeCIRunsEveryGate(t *testing.T) {
	out := makeDryRun(t, "ci")
	gates := []string{
		"gofmt -l",
		"goimports -l",
		"go build",
		"go vet",
		"golangci-lint run",
		"go test -race",
		"-coverprofile=",
		"govulncheck",
		"actionlint",
		"shellcheck -s sh install.sh",
		"mkdocs build --strict",
		"goreleaser check",
	}
	for _, want := range gates {
		if !strings.Contains(out, want) {
			t.Errorf("make -n ci output lacks gate %q:\n%s", want, out)
		}
	}
	if !regexp.MustCompile(`\b80\b`).MatchString(out) {
		t.Errorf("make -n ci output lacks the 80%% coverage threshold check:\n%s", out)
	}
}

var (
	toolchainDefaultRE = regexp.MustCompile(`(?m)^(export\s+)?GOTOOLCHAIN\s*\?=\s*auto\s*$`)
	toolchainExportRE  = regexp.MustCompile(`(?m)^export\s+GOTOOLCHAIN\b`)
)

func TestMakefileGoToolchainDefault(t *testing.T) {
	mf := readMakefile(t)
	if !toolchainDefaultRE.MatchString(mf) {
		t.Errorf("Makefile lacks an overridable `GOTOOLCHAIN ?= auto` default")
	}
	if !toolchainExportRE.MatchString(mf) {
		t.Errorf("Makefile does not export GOTOOLCHAIN")
	}
}

func TestContributingMentionsMake(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "CONTRIBUTING.md"))
	if err != nil {
		t.Fatalf("CONTRIBUTING.md unreadable: %v", err)
	}
	for _, want := range []string{"make ci", "make install"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("CONTRIBUTING.md lacks %q", want)
		}
	}
}
