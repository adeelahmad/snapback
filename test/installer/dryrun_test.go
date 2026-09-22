package installer

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDryRunPrintsPlanWithoutNetwork(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	env := dryRunEnv("Linux", "x86_64")
	env["SNAPBACK_BASE_URL"] = srv.URL
	env["SNAPBACK_INSTALL_DIR"] = installDir

	stdout, stderr, code := runInstaller(t, env)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	for _, want := range []string{
		srv.URL + "/snapback_linux_amd64.tar.gz",
		srv.URL + "/checksums.txt",
		installDir,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
		}
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("server saw %d requests during dry-run, want 0", n)
	}
}

const defaultBaseURL = "https://github.com/adeelahmad/snapback/releases/latest/download"

func TestDryRunDefaultBaseURLIsGitHubReleases(t *testing.T) {
	stdout, stderr, code := runInstaller(t, dryRunEnv("Linux", "x86_64"))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if want := defaultBaseURL + "/snapback_linux_amd64.tar.gz"; !strings.Contains(stdout, want) {
		t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
	}
	if strings.Contains(stdout+stderr, "example.invalid") {
		t.Errorf("output contains placeholder %q; stdout=%q", "example.invalid", stdout)
	}
}

func TestNextStepsDarwin(t *testing.T) {
	assertNextSteps(t, "Darwin", "arm64", "macFUSE")
}

func TestNextStepsLinux(t *testing.T) {
	assertNextSteps(t, "Linux", "amd64", "fuse3")
}

func assertNextSteps(t *testing.T, osName, arch, fuseHint string) {
	t.Helper()
	stdout, stderr, code := runInstaller(t, dryRunEnv(osName, arch))

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	combined := stdout + stderr
	for _, want := range []string{"snapback version", fuseHint} {
		if !strings.Contains(combined, want) {
			t.Errorf("output does not contain %q; output=%q", want, combined)
		}
	}
}

// snapbackSubcommand matches a "snapback <subcommand>" invocation in prose.
var snapbackSubcommand = regexp.MustCompile(`\bsnapback ([a-z][a-z0-9-]*)`)

// TestNextStepsNameOnlyExistingCommands guards against the installer telling
// users to run a snapback subcommand that does not exist: every command the
// next steps name must be listed by the built binary's `snapback help`, and
// `snapback version` stays the last step.
func TestNextStepsNameOnlyExistingCommands(t *testing.T) {
	known := helpCommands(t, buildSnapback(t))
	for _, tc := range []struct {
		osName, arch string
	}{
		{"Darwin", "arm64"},
		{"Linux", "amd64"},
	} {
		t.Run(tc.osName, func(t *testing.T) {
			stdout, stderr, code := runInstaller(t, dryRunEnv(tc.osName, tc.arch))
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}

			combined := stdout + stderr
			_, steps, ok := strings.Cut(combined, "Next steps:")
			if !ok {
				t.Fatalf("output has no %q section; output=%q", "Next steps:", combined)
			}

			matches := snapbackSubcommand.FindAllStringSubmatch(steps, -1)
			if len(matches) == 0 {
				t.Fatalf("next steps name no snapback command, want %q; steps=%q", "snapback version", steps)
			}
			if got, want := matches[len(matches)-1][1], "version"; got != want {
				t.Errorf("final next step names `snapback %s`, want `snapback %s`; steps=%q", got, want, steps)
			}
			for _, m := range matches {
				if !known[m[1]] {
					t.Errorf("next steps name `snapback %s`, which `snapback help` does not list; steps=%q", m[1], steps)
				}
			}
		})
	}
}
