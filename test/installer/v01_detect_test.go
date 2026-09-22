package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// helpCommand matches a subcommand line in `snapback help` output.
var helpCommand = regexp.MustCompile(`(?m)^\s+([a-z][a-z0-9-]*)\s`)

// claimsSupported matches "supported" as a word, so "unsupported" does not count.
var claimsSupported = regexp.MustCompile(`(?i)\bsupported\b`)

// buildSnapback builds the real CLI into a temp dir and returns its path.
func buildSnapback(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "snapback")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/snapback")
	cmd.Dir = filepath.Join("..", "..")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/snapback: %v\n%s", err, out)
	}
	return bin
}

// helpCommands returns the subcommands the built binary's `snapback help` lists.
func helpCommands(t *testing.T, bin string) map[string]bool {
	t.Helper()
	out, err := exec.Command(bin, "help").CombinedOutput()
	if err != nil {
		t.Fatalf("snapback help: %v\n%s", err, out)
	}
	cmds := map[string]bool{}
	for _, m := range helpCommand.FindAllStringSubmatch(string(out), -1) {
		cmds[m[1]] = true
	}
	if len(cmds) == 0 {
		t.Fatalf("snapback help lists no commands; output=%q", out)
	}
	return cmds
}

func TestV01NextStepsNameRealCommands(t *testing.T) {
	stdout, stderr, code := runInstaller(t, dryRunEnv("Linux", "amd64"))
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	_, steps, ok := strings.Cut(stdout+stderr, "Next steps:")
	if !ok {
		t.Fatalf("output has no %q section; output=%q", "Next steps:", stdout+stderr)
	}

	for _, want := range []string{"snapback config", "snapback install service", "snapback doctor", "fuse3"} {
		if !strings.Contains(steps, want) {
			t.Errorf("next steps do not contain %q; steps=%q", want, steps)
		}
	}

	matches := snapbackSubcommand.FindAllStringSubmatch(steps, -1)
	if len(matches) == 0 {
		t.Fatalf("next steps name no snapback command; steps=%q", steps)
	}
	if got, want := matches[len(matches)-1][1], "version"; got != want {
		t.Errorf("final next step names `snapback %s`, want `snapback %s`; steps=%q", got, want, steps)
	}

	known := helpCommands(t, buildSnapback(t))
	for _, m := range matches {
		if !known[m[1]] {
			t.Errorf("next steps name `snapback %s`, which `snapback help` does not list; steps=%q", m[1], steps)
		}
	}
}

func TestV01DarwinWarnsUnverified(t *testing.T) {
	stdout, stderr, code := runInstaller(t, dryRunEnv("Darwin", "arm64"))
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	for _, want := range []string{"not verified", "follow-up"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr does not contain %q; stderr=%q", want, stderr)
		}
	}
	combined := stdout + stderr
	if !strings.Contains(combined, "macFUSE") {
		t.Errorf("output does not contain %q; output=%q", "macFUSE", combined)
	}
	for _, banned := range []string{"supported on macOS", "static", "production-ready"} {
		if strings.Contains(combined, banned) {
			t.Errorf("output contains banned claim %q; output=%q", banned, combined)
		}
	}
}

func TestV01NoFalsePlatformClaims(t *testing.T) {
	cases := []struct {
		osName, arch string
		wantFail     bool
		wantWarning  string // substring required on stderr; "" means no warning at all
		hasAsset     bool
	}{
		{osName: "Linux", arch: "amd64", hasAsset: true},
		{osName: "Linux", arch: "arm64", hasAsset: true},
		{osName: "Linux", arch: "armv7l", wantWarning: "unverified", hasAsset: true},
		{osName: "Linux", arch: "mips", wantWarning: "unverified", hasAsset: true},
		{osName: "Linux", arch: "mipsel", wantWarning: "unverified", hasAsset: true},
		{osName: "Darwin", arch: "amd64", wantWarning: "not verified", hasAsset: true},
		{osName: "Darwin", arch: "arm64", wantWarning: "not verified", hasAsset: true},
		{osName: "FreeBSD", arch: "amd64", wantFail: true},
		{osName: "Windows_NT", arch: "x86_64", wantFail: true},
	}
	for _, tc := range cases {
		t.Run(tc.osName+"/"+tc.arch, func(t *testing.T) {
			stdout, stderr, code := runInstaller(t, dryRunEnv(tc.osName, tc.arch))
			combined := stdout + stderr

			if tc.wantFail {
				if code == 0 {
					t.Errorf("exit code = 0, want non-zero")
				}
				if !strings.Contains(stderr, "unsupported") {
					t.Errorf("stderr does not contain %q; stderr=%q", "unsupported", stderr)
				}
			} else if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}

			switch {
			case tc.wantFail:
			case tc.wantWarning == "":
				if strings.Contains(stderr, "warning:") && !strings.Contains(stderr, "not on your PATH") {
					t.Errorf("stderr has a platform warning, want none; stderr=%q", stderr)
				}
				if strings.Contains(combined, "unverified") || strings.Contains(combined, "not verified") {
					t.Errorf("output warns the target is unverified, want no warning; output=%q", combined)
				}
			default:
				if !strings.Contains(stderr, tc.wantWarning) {
					t.Errorf("stderr does not contain %q; stderr=%q", tc.wantWarning, stderr)
				}
			}

			if !tc.hasAsset && claimsSupported.MatchString(combined) {
				t.Errorf("output claims %q for a target without a release asset; output=%q", "supported", combined)
			}
		})
	}
}

// recordingStub is a shell script that appends its name and args to a log.
const recordingStub = "#!/bin/sh\necho \"$(basename \"$0\") $*\" >> \"$STUB_LOG\"\nexit 0\n"

func TestV01NeverInstallsFuse(t *testing.T) {
	data, err := os.ReadFile(installScript)
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("install.sh is empty, want the installer script")
	}
	for i, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "apt-get install") || strings.Contains(line, "brew install") {
			t.Errorf("install.sh:%d has a package install line: %s", i+1, strings.TrimSpace(line))
		}
	}

	stubDir := t.TempDir()
	stubLog := filepath.Join(t.TempDir(), "stubs.log")
	stubs := []string{"apt-get", "brew", "dnf", "sudo"}
	for _, name := range stubs {
		if err := os.WriteFile(filepath.Join(stubDir, name), []byte(recordingStub), 0o755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}

	archive := fakeArchive(t)
	srv := newReleaseServer(t, map[string][]byte{
		linuxAmd64Asset: archive,
		"checksums.txt": []byte(sha256Hex(archive) + "  " + linuxAmd64Asset + "\n"),
	})
	installDir := t.TempDir()
	env := installEnv(srv.URL, installDir)
	env["PATH"] = stubDir + string(os.PathListSeparator) + os.Getenv("PATH")
	env["STUB_LOG"] = stubLog

	_, stderr, code := runInstaller(t, env)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(installDir, "snapback")); err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}

	logged, err := os.ReadFile(stubLog)
	switch {
	case os.IsNotExist(err):
	case err != nil:
		t.Fatalf("read stub log: %v", err)
	case len(logged) > 0:
		t.Errorf("installer invoked package manager stubs %v: %s", stubs, logged)
	}
}
