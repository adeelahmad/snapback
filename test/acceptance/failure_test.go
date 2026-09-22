//go:build integration

package acceptance

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	// sampleWindow is how long the lock and prune cases sample status and
	// the listing for an empty listing in state ready.
	sampleWindow = 10 * time.Second
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

// historyEntries drops info.json and latest, leaving one entry per snapshot.
func historyEntries(names []string) []string {
	return slices.DeleteFunc(aliasesOnly(names), func(n string) bool { return n == "info.json" })
}

// readyWithoutCode reports whether status says ready with no repository code.
func readyWithoutCode(s statusEnvelope) bool {
	if !s.OK || s.Data.State != "ready" {
		return false
	}
	for _, r := range s.Data.Repos {
		if r.Code != "" {
			return false
		}
	}
	return true
}

// neverEmptyWhileReady samples status and the listing for sampleWindow and
// reports every sample where status is ready but the listing shows no history.
func neverEmptyWhileReady(t *testing.T, e env, name, link string) {
	t.Helper()
	deadline := time.Now().Add(sampleWindow)
	for time.Now().Before(deadline) {
		s, raw := readStatus(t, e)
		names, err := listNames(link)
		if readyWithoutCode(s) && err == nil && len(historyEntries(names)) == 0 {
			t.Errorf("%s: list %s = %v in state ready, want the history aliases or a failure code; status: %s", name, link, names, raw)
			return
		}
		time.Sleep(statusPoll)
	}
}

func TestAcc14FailuresDistinctFromEmpty(t *testing.T) {
	recordEvidence(t, "acc14")
	requireFUSE(t)

	cases := []struct {
		name     string
		wantCode string // "" means the listing must never be empty while ready
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
			link := filepath.Join(h.proj, ".snapshot")

			switch c.name {
			case "lock", "prune":
				neverEmptyWhileReady(t, e, c.name, link)
			case "offline":
				code, raw := failureCode(t, e, d)
				if code != "" && code != c.wantCode {
					t.Errorf("%s: status --json code = %q, want %q; status: %s", c.name, code, c.wantCode, raw)
				}
				names, err := listNames(link)
				info, _ := os.ReadFile(filepath.Join(link, "info.json"))
				// An absent .snapshot is not a failure state; the listing
				// itself must fail (EIO) to count.
				listFailed := err != nil && !errors.Is(err, fs.ErrNotExist)
				t.Logf("%s: list = %v, %v; info.json = %q; status code = %q", c.name, names, err, info, code)
				if !listFailed && code != c.wantCode && !strings.Contains(string(info), c.wantCode) {
					t.Errorf("%s: list %s = %v, %v; info.json = %q; status code = %q; want the listing to fail or info.json or status to state %q",
						c.name, link, names, err, info, code, c.wantCode)
				}
			case "mount":
				code, raw := failureCode(t, e, d)
				if code != c.wantCode {
					t.Errorf("%s: status --json code = %q, want %q with the daemon still up; status: %s", c.name, code, c.wantCode, raw)
				}
			}

			if hit := liveBytesUnder(link); hit != "" {
				t.Errorf("%s: live file bytes found under history at %s, want none", c.name, hit)
			}
			if names, err := listNames(link); err != nil && errors.Is(err, fs.ErrNotExist) {
				t.Logf("%s: .snapshot entry absent (%v, %v)", c.name, names, err)
			}
		})
	}
}
