package installer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const pathHintHeader = "add it to your PATH:"

// pathHintOutput dry-runs the installer for a shell and install dir and
// returns everything it printed. The base URL points at a local server that
// answers 404, so no request can reach the network.
func pathHintOutput(t *testing.T, shell, installDir string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	env := dryRunEnv("Linux", "x86_64")
	env["SNAPBACK_BASE_URL"] = srv.URL
	env["SNAPBACK_INSTALL_DIR"] = installDir
	env["SNAPBACK_SHELL"] = shell

	stdout, stderr, code := runInstaller(t, env)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	return stdout + stderr
}

// TestPathHintPerShell pins that a user whose install dir is off PATH is given
// the exact line their own shell accepts, rather than a generic warning.
func TestPathHintPerShell(t *testing.T) {
	for _, tc := range []struct{ shell, want string }{
		{"bash", `export PATH="%s:$PATH"  # add this line to ~/.bashrc`},
		{"zsh", `export PATH="%s:$PATH"  # add this line to ~/.zshrc`},
		{"fish", `fish_add_path %s`},
	} {
		t.Run(tc.shell, func(t *testing.T) {
			installDir := t.TempDir()
			out := pathHintOutput(t, "/bin/"+tc.shell, installDir)
			want := strings.Replace(tc.want, "%s", installDir, 1)

			_, after, ok := strings.Cut(out, pathHintHeader)
			if !ok {
				t.Fatalf("output has no %q; output=%q", pathHintHeader, out)
			}
			lines := strings.Split(strings.TrimPrefix(after, "\n"), "\n")
			if got := strings.TrimSpace(lines[0]); got != want {
				t.Errorf("PATH hint line = %q, want %q", got, want)
			}
			if got := strings.TrimSpace(lines[1]); got != "" {
				t.Errorf("PATH hint has a second line %q, want exactly one line", got)
			}
		})
	}
}

// TestPathHintSilentWhenOnPath pins that an install dir already on PATH earns
// no advice about PATH at all.
func TestPathHintSilentWhenOnPath(t *testing.T) {
	installDir := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	env := dryRunEnv("Linux", "x86_64")
	env["SNAPBACK_BASE_URL"] = srv.URL
	env["SNAPBACK_INSTALL_DIR"] = installDir
	env["SNAPBACK_SHELL"] = "/bin/bash"
	env["PATH"] = installDir + ":/usr/bin:/bin"

	stdout, stderr, code := runInstaller(t, env)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	out := stdout + stderr
	if strings.Contains(out, pathHintHeader) {
		t.Errorf("output advises about PATH for a dir already on PATH; output=%q", out)
	}
}
