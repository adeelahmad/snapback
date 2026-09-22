package shellhook

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/cli"
	"github.com/adeelahmad/snapback/internal/ipc"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

func bufEnv(getenv func(string) string) (cli.Env, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	return cli.Env{Stdout: &stdout, Stderr: &stderr, Getenv: getenv}, &stdout, &stderr
}

func TestShellHookCommandPrintsScript(t *testing.T) {
	if got := Command().Name; got != "shell-hook" {
		t.Fatalf("Command().Name = %q, want %q", got, "shell-hook")
	}
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			want, err := Script(shell)
			if err != nil {
				t.Fatalf("Script(%q) = %v", shell, err)
			}
			env, stdout, stderr := bufEnv(mapEnv(nil))
			if got := Command().Run(context.Background(), env, []string{shell}); got != 0 {
				t.Errorf("Run(%q) = %d, want 0", shell, got)
			}
			if got := stdout.String(); got != want {
				t.Errorf("Run(%q) stdout = %q, want Script(%q)", shell, got, shell)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("Run(%q) stderr = %q, want empty", shell, got)
			}
		})
	}
}

func TestShellHookCommandUsage(t *testing.T) {
	const usage = "usage: snapback shell-hook bash|zsh|fish"
	tests := []struct {
		name string
		args []string
	}{
		{"none", nil},
		{"unknown", []string{"tcsh"}},
		{"two", []string{"bash", "zsh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, stdout, stderr := bufEnv(mapEnv(nil))
			if got := Command().Run(context.Background(), env, tt.args); got != 2 {
				t.Errorf("Run(%q) = %d, want 2", tt.args, got)
			}
			if got := stdout.String(); got != "" {
				t.Errorf("Run(%q) stdout = %q, want empty", tt.args, got)
			}
			if got := stderr.String(); !strings.Contains(got, usage) {
				t.Errorf("Run(%q) stderr = %q, want it to contain %q", tt.args, got, usage)
			}
		})
	}
}

func TestNotifyCommandSilentWithoutDaemon(t *testing.T) {
	if got := NotifyCommand().Name; got != "notify" {
		t.Fatalf("NotifyCommand().Name = %q, want %q", got, "notify")
	}
	tmp := shortTempDir(t)
	run := filepath.Join(tmp, "run")
	if err := os.Mkdir(run, 0o700); err != nil {
		t.Fatal(err)
	}
	env, stdout, stderr := bufEnv(mapEnv(map[string]string{"XDG_RUNTIME_DIR": run, "HOME": tmp}))
	args := []string{"--timeout", "200ms", "--session", "9", "--", tmp}
	start := time.Now()
	got := NotifyCommand().Run(context.Background(), env, args)
	elapsed := time.Since(start)
	if got != 0 {
		t.Errorf("Run(%q) = %d, want 0", args, got)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("Run(%q) stdout = %q, stderr = %q, want both empty", args, stdout, stderr)
	}
	if elapsed >= 300*time.Millisecond {
		t.Errorf("Run(%q) took %v, want < 300ms", args, elapsed)
	}
}

func TestNotifyCommandSilentOnBadArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"none", nil},
		{"bad timeout", []string{"--timeout", "nope", "--", "/"}},
		{"unknown flag", []string{"--bogus"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := shortTempDir(t)
			env, stdout, stderr := bufEnv(mapEnv(map[string]string{"XDG_RUNTIME_DIR": tmp, "HOME": tmp}))
			if got := NotifyCommand().Run(context.Background(), env, tt.args); got != 0 {
				t.Errorf("Run(%q) = %d, want 0", tt.args, got)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Errorf("Run(%q) stdout = %q, stderr = %q, want both empty", tt.args, stdout, stderr)
			}
		})
	}
}

func TestNotifyCommandSendsPathAfterDashDash(t *testing.T) {
	tmp := shortTempDir(t)
	sock := filepath.Join(tmp, "d.sock")
	got := fakeDaemon(t, sock, ipc.Response{OK: true})
	badConfig := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(badConfig, []byte("{: not yaml ["), 0o600); err != nil {
		t.Fatal(err)
	}
	env, stdout, stderr := bufEnv(mapEnv(map[string]string{"HOME": tmp}))
	env.ConfigPath = badConfig
	args := []string{"--socket", sock, "--timeout", "200ms", "--session", "7", "--", "--json"}
	if code := NotifyCommand().Run(context.Background(), env, args); code != 0 {
		t.Errorf("Run(%q) = %d, want 0", args, code)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Errorf("Run(%q) stdout = %q, stderr = %q, want both empty", args, stdout, stderr)
	}
	req := receive(t, got)
	if req.Op != ipc.OpDirEvent {
		t.Errorf("request Op = %q, want %q", req.Op, ipc.OpDirEvent)
	}
	if want := rawpath.Path("--json"); !bytes.Equal(req.Path, want) {
		t.Errorf("request Path = %q, want %q", req.Path, want)
	}
	if req.Session != "7" {
		t.Errorf("request Session = %q, want %q", req.Session, "7")
	}
}
