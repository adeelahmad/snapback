//go:build integration

package acceptance

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	// statusPoll is the interval between status polls.
	statusPoll = 250 * time.Millisecond
	// statusSettle is how long a failure case waits for a repository error
	// code to appear in status before giving up.
	statusSettle = 30 * time.Second
	// liveSentinel marks live file bytes that must never appear in history.
	liveSentinel = "LIVE-ONLY-BYTES-acc14\n"
)

// statusRepo is one repository entry in snapback status --json.
type statusRepo struct {
	ID    string
	State string
	Code  string
}

// statusEnvelope is the snapback status --json output, success or error.
type statusEnvelope struct {
	OK   bool   `json:"ok"`
	Code string `json:"code"`
	Data struct {
		State    string
		Repos    []statusRepo
		Recovery *struct {
			Unmounted []string
			Foreign   []string
		} `json:"recovery"`
	} `json:"data"`
}

// readStatus runs snapback status --json once and decodes it.
func readStatus(t *testing.T, e env) (statusEnvelope, string) {
	t.Helper()
	stdout, stderr, _ := runSnapback(t, e, "status", "--json")
	var s statusEnvelope
	if err := json.Unmarshal([]byte(stdout), &s); err != nil {
		return statusEnvelope{}, "status --json not JSON: " + err.Error() + "; stdout: " + stdout + "; stderr: " + stderr
	}
	return s, stdout
}

// failureCode polls status until a repository or envelope error code shows
// up, the daemon exits, or statusSettle passes, and returns the code seen.
func failureCode(t *testing.T, e env, d *daemon) (code, raw string) {
	t.Helper()
	deadline := time.Now().Add(statusSettle)
	for {
		s, out := readStatus(t, e)
		raw = out
		for _, r := range s.Data.Repos {
			if r.Code != "" {
				return r.Code, raw
			}
		}
		// An unreachable daemon is not a repository failure state.
		if s.Code != "" && !strings.Contains(out, "daemon not running") {
			return s.Code, raw
		}
		select {
		case <-d.done:
			return "", raw + " (daemon exited: " + errString(d.err) + ")"
		default:
		}
		if time.Now().After(deadline) {
			return "", raw
		}
		time.Sleep(statusPoll)
	}
}

func errString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

// holdResticLock starts restic args against the repo, waits until it has
// written a lock file, then stops it with SIGSTOP so its lock stays held by
// a live process (restic treats locks of dead local processes as stale).
func holdResticLock(t *testing.T, fx fixture, args ...string) {
	t.Helper()
	full := append([]string{"-r", fx.Repo, "--password-file", fx.PasswordFile}, args...)
	cmd := exec.Command("restic", full...)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start restic %s: %v", strings.Join(args, " "), err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGCONT)
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	locks := filepath.Join(fx.Repo, "locks")
	deadline := time.Now().Add(10 * time.Second)
	for {
		if ents, err := os.ReadDir(locks); err == nil && len(ents) > 0 {
			if err := cmd.Process.Signal(syscall.SIGSTOP); err != nil {
				t.Fatalf("stop restic %s holding its lock: %v", strings.Join(args, " "), err)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("restic %s wrote no lock under %s within 10s", strings.Join(args, " "), locks)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// liveBytesUnder reports the first file under dir whose bytes contain the
// live sentinel, or "" when none does.
func liveBytesUnder(dir string) string {
	var hit string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || hit != "" {
			return nil
		}
		if d.Type().IsRegular() {
			if b, err := os.ReadFile(p); err == nil && strings.Contains(string(b), liveSentinel) {
				hit = p
			}
		}
		return nil
	})
	return hit
}

func TestAcc14FailuresDistinctFromEmpty(t *testing.T) {
	recordEvidence(t, "acc14")
	requireFUSE(t)

	cases := []struct {
		name     string
		wantCode string // "" means any code distinct from the others
		breakIt  func(t *testing.T, h histRepo, e env)
	}{
		{"offline", "repository_unavailable", func(t *testing.T, h histRepo, _ env) {
			if err := os.Rename(h.fx.Repo, h.fx.Repo+".gone"); err != nil {
				t.Fatalf("rename repo offline: %v", err)
			}
		}},
		{"lock", "", func(t *testing.T, h histRepo, _ env) {
			holdResticLock(t, h.fx, "forget", "--keep-last", "1000")
		}},
		{"prune", "", func(t *testing.T, h histRepo, _ env) {
			holdResticLock(t, h.fx, "prune")
		}},
		{"mount", "mount_failure", func(t *testing.T, _ histRepo, e env) {
			mnt := filepath.Join(e.Root, "state", "mounts", "history")
			if err := os.MkdirAll(filepath.Dir(mnt), 0o755); err != nil {
				t.Fatalf("mkdir mounts: %v", err)
			}
			if err := os.WriteFile(mnt, []byte("not a directory\n"), 0o444); err != nil {
				t.Fatalf("make history mountpoint unusable: %v", err)
			}
		}},
	}

	seen := map[string]string{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHistRepo(t)
			writeFiles(t, h.proj, map[string]string{"a.txt": "history bytes\n"})
			backup(t, h.fx, "", histHost, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "daily", h.fx.Root)
			writeFiles(t, h.proj, map[string]string{"a.txt": liveSentinel})
			e := writeHistConfig(t, h)
			c.breakIt(t, h, e)

			d := startDaemon(t, e)
			t.Cleanup(func() { _ = d.Stop() })
			_, _, _ = runSnapback(t, e, "link", h.proj)

			code, raw := failureCode(t, e, d)
			if code == "" {
				t.Errorf("%s: status --json error code = \"\", want a non-empty failure code; status: %s", c.name, raw)
			}
			if c.wantCode != "" && code != c.wantCode {
				t.Errorf("%s: status --json code = %q, want %q", c.name, code, c.wantCode)
			}
			if code != "" {
				if prev, dup := seen[code]; dup {
					t.Errorf("%s: code %q equals the code of case %s, want distinct failure states", c.name, code, prev)
				}
				seen[code] = c.name
			}

			link := filepath.Join(h.proj, ".snapshot")
			names, err := listNames(link)
			if err == nil {
				t.Errorf("%s: list %s = %v, <nil>, want an error rather than a (possibly empty) listing", c.name, link, names)
			}
			if hit := liveBytesUnder(link); hit != "" {
				t.Errorf("%s: live file bytes found under history at %s, want none", c.name, hit)
			}
			if err != nil && errors.Is(err, fs.ErrNotExist) {
				t.Logf("%s: .snapshot entry absent (%v)", c.name, err)
			}
		})
	}
}
