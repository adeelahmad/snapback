package installer

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var pkgManagerInstall = regexp.MustCompile(`(brew|apt|apt-get|dnf|yum|opkg|pacman|apk|port)\s+(install|add)`)

func readInstallScript(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(installScript)
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}
	return strings.Split(string(data), "\n")
}

// README.md §17: the installer never installs FUSE or any other package.
func TestStaticNoPackageManagerInstall(t *testing.T) {
	lines := readInstallScript(t)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "echo") || strings.HasPrefix(trimmed, "printf") {
			continue
		}
		if m := pkgManagerInstall.FindString(line); m != "" {
			t.Errorf("install.sh:%d invokes a package manager (%q): %s", i+1, m, trimmed)
		}
	}
}

func TestStaticPosixHeader(t *testing.T) {
	lines := readInstallScript(t)

	if lines[0] != "#!/bin/sh" {
		t.Errorf("first line = %q, want %q", lines[0], "#!/bin/sh")
	}
	hasSetEU := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "set -eu" {
			hasSetEU = true
			break
		}
	}
	if !hasSetEU {
		t.Error("install.sh has no `set -eu` line")
	}
}

func TestStaticNoBashisms(t *testing.T) {
	lines := readInstallScript(t)
	bashisms := []string{"[[", "function ", "<<<", "$'", "source "}

	for i, line := range lines {
		for _, b := range bashisms {
			if strings.Contains(line, b) {
				t.Errorf("install.sh:%d contains bashism %q: %s", i+1, b, strings.TrimSpace(line))
			}
		}
	}
}
