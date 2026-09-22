package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/daemon"
)

// signalTimeout bounds both daemon startup and its exit after a signal.
const signalTimeout = 10 * time.Second

// writeSignalConfig writes a one-repository, one-root config under tmp that
// uses the given restic binary, and returns the config path and state dir.
func writeSignalConfig(t *testing.T, tmp, restic string) (cfgPath, stateDir string) {
	t.Helper()
	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("secret\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", pw, err)
	}
	root := filepath.Join(tmp, "work")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", root, err)
	}
	stateDir = filepath.Join(tmp, "state")
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("state_dir: " + stateDir + "\n")
	b.WriteString("repositories:\n")
	b.WriteString("  - id: personal\n")
	b.WriteString("    repository: " + filepath.Join(tmp, "repo") + "\n")
	b.WriteString("    restic_binary: " + restic + "\n")
	b.WriteString("    password_file: " + pw + "\n")
	b.WriteString("roots:\n")
	b.WriteString("  - id: work\n")
	b.WriteString("    local_path: " + root + "\n")
	b.WriteString("    repository_id: personal\n")
	cfgPath = filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", cfgPath, err)
	}
	return cfgPath, stateDir
}

// waitForSocket polls until sock exists, failing if the daemon exits first
// or the timeout passes.
func waitForSocket(t *testing.T, sock string, done chan error, stderr *bytes.Buffer) {
	t.Helper()
	deadline := time.Now().Add(signalTimeout)
	for {
		if _, err := os.Stat(sock); err == nil {
			return
		}
		select {
		case err := <-done:
			done <- err
			t.Fatalf("snapback run exited before serving: %v\nstderr:\n%s", err, stderr.String())
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("socket %s not created within %v\nstderr:\n%s", sock, signalTimeout, stderr.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestRunExitsCleanlyOnSIGTERM(t *testing.T) {
	bin := buildBinary(t, "")
	tests := []struct {
		name string
		sig  syscall.Signal
	}{
		{name: "SIGTERM", sig: syscall.SIGTERM},
		{name: "SIGINT", sig: syscall.SIGINT},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp, err := os.MkdirTemp("/tmp", "sb")
			if err != nil {
				t.Fatalf("os.MkdirTemp(/tmp, sb) = %v", err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(tmp) })
			pathDir := filepath.Join(tmp, "bin")
			if err := os.MkdirAll(pathDir, 0o755); err != nil {
				t.Fatalf("os.MkdirAll(%s) = %v", pathDir, err)
			}
			restic := filepath.Join(pathDir, "restic")
			if err := os.WriteFile(restic, []byte(fakeRestic), 0o755); err != nil {
				t.Fatalf("os.WriteFile(%s) = %v", restic, err)
			}
			cfgPath, stateDir := writeSignalConfig(t, tmp, restic)
			runtimeDir := filepath.Join(tmp, "rt")
			sock := filepath.Join(runtimeDir, "snapback", "daemon.sock")

			var stderr bytes.Buffer
			cmd := exec.Command(bin, "--config", cfgPath, "run")
			cmd.Env = append(os.Environ(),
				"PATH="+pathDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"XDG_RUNTIME_DIR="+runtimeDir,
			)
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				t.Fatalf("start snapback run = %v", err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
				<-done
			})

			waitForSocket(t, sock, done, &stderr)

			if err := cmd.Process.Signal(tc.sig); err != nil {
				t.Fatalf("cmd.Process.Signal(%v) = %v", tc.sig, err)
			}
			select {
			case err := <-done:
				done <- err
				if err != nil {
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) {
						t.Fatalf("snapback run after %v: exit = %v (code %d), want exit status 0\nstderr:\n%s",
							tc.sig, err, exitErr.ExitCode(), stderr.String())
					}
					t.Fatalf("snapback run after %v: wait = %v, want exit status 0", tc.sig, err)
				}
			case <-time.After(signalTimeout):
				t.Fatalf("snapback run did not exit within %v after %v\nstderr:\n%s",
					signalTimeout, tc.sig, stderr.String())
			}

			unlock, err := daemon.Lock(stateDir)
			if err != nil {
				t.Fatalf("daemon.Lock(%s) after exit = %v, want the lock released", stateDir, err)
			}
			unlock()
		})
	}
}
