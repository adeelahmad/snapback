package resticfx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

type fakeCall struct {
	name string
	args []string
}

// fakeRunner records every call and answers through respond (if set).
type fakeRunner struct {
	calls   []fakeCall
	respond func(name string, args []string) ([]byte, error)
}

func (f *fakeRunner) Run(_ context.Context, name string, args []string) ([]byte, error) {
	f.calls = append(f.calls, fakeCall{name: name, args: slices.Clone(args)})
	if f.respond == nil {
		return nil, nil
	}
	return f.respond(name, args)
}

func (f *fakeRunner) hasCallWith(word string) bool {
	for _, c := range f.calls {
		if slices.Contains(c.args, word) {
			return true
		}
	}
	return false
}

// resticLike simulates restic: "init" creates the repo dir with a config file,
// "snapshots" returns snapshotsJSON.
func resticLike(t *testing.T, repo, snapshotsJSON string) func(string, []string) ([]byte, error) {
	t.Helper()
	return func(_ string, args []string) ([]byte, error) {
		switch {
		case slices.Contains(args, "init"):
			if err := os.MkdirAll(repo, 0o700); err != nil {
				return nil, err
			}
			return []byte("created restic repository\n"), os.WriteFile(filepath.Join(repo, "config"), []byte("cfg"), 0o600)
		case slices.Contains(args, "snapshots"):
			return []byte(snapshotsJSON), nil
		default:
			return nil, nil
		}
	}
}

func newTestGuard(t *testing.T) Guard {
	t.Helper()
	return Guard{TempRoot: t.TempDir(), Home: t.TempDir()}
}

func TestNewFixtureRefusesUnguardedRepo(t *testing.T) {
	g := newTestGuard(t)
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	for label, repo := range map[string]string{"home": filepath.Join(g.Home, "repo"), "other-remote": "rclone:gdrive:other"} {
		t.Run(label, func(t *testing.T) {
			fr := &fakeRunner{}
			if _, err := NewFixture(fr, repo, pw, g); err == nil {
				t.Errorf("NewFixture(%q): expected guard error, got nil", repo)
			}
			if len(fr.calls) != 0 {
				t.Errorf("NewFixture(%q): runner recorded %d calls, want 0", repo, len(fr.calls))
			}
		})
	}
}

func TestFixtureInitBackupSnapshots(t *testing.T) {
	ctx := context.Background()
	g := newTestGuard(t)
	repo := filepath.Join(g.TempRoot, "repo")
	src := t.TempDir()
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	snaps := `[{"time":"2026-09-22T01:02:03Z","hostname":"h","paths":["` + src + `"],"id":"` + fullIDA + `","short_id":"` + fullIDA[:8] + `"}]`
	fr := &fakeRunner{respond: resticLike(t, repo, snaps)}

	f, err := NewFixture(fr, repo, pw, g)
	if err != nil {
		t.Fatalf("NewFixture: %v", err)
	}
	if err := f.Init(ctx); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := f.Backup(ctx, src); err != nil {
		t.Fatalf("Backup: %v", err)
	}
	got, err := f.Snapshots(ctx)
	if err != nil {
		t.Fatalf("Snapshots: %v", err)
	}

	want := []fakeCall{
		{name: "restic", args: InitArgs(repo, pw)},
		{name: "restic", args: BackupArgs(repo, pw, src)},
		{name: "restic", args: SnapshotsArgs(repo, pw)},
	}
	if len(fr.calls) != len(want) {
		t.Fatalf("recorded %d calls %v, want %d %v", len(fr.calls), fr.calls, len(want), want)
	}
	for i := range want {
		if fr.calls[i].name != want[i].name || !slices.Equal(fr.calls[i].args, want[i].args) {
			t.Errorf("call %d = %s %q, want %s %q", i, fr.calls[i].name, fr.calls[i].args, want[i].name, want[i].args)
		}
	}
	if len(got) != 1 || got[0].ID != fullIDA {
		t.Fatalf("Snapshots = %+v, want one snapshot with full ID %s", got, fullIDA)
	}
}

func TestFixtureInitRefusesExistingRepo(t *testing.T) {
	g := newTestGuard(t)
	repo := filepath.Join(g.TempRoot, "repo")
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "keep"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	fr := &fakeRunner{respond: resticLike(t, repo, "[]")}

	f, err := NewFixture(fr, repo, pw, g)
	if err != nil {
		t.Fatalf("NewFixture: %v", err)
	}
	if err := f.Init(context.Background()); err == nil {
		t.Error("Init on existing non-empty repo dir: expected error, got nil")
	}
	if fr.hasCallWith("init") {
		t.Errorf("runner recorded an init call %v; want none", fr.calls)
	}
}

func TestFixtureDestroyOnlyOwnRepo(t *testing.T) {
	g := newTestGuard(t)
	sibling := filepath.Join(g.TempRoot, "sibling.txt")
	if err := os.WriteFile(sibling, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}

	// A fixture that never ran Init must not delete a pre-existing repo dir.
	foreign := filepath.Join(g.TempRoot, "foreign")
	if err := os.MkdirAll(foreign, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, "config"), []byte("cfg"), 0o600); err != nil {
		t.Fatal(err)
	}
	f1, err := NewFixture(&fakeRunner{}, foreign, pw, g)
	if err != nil {
		t.Fatalf("NewFixture(foreign): %v", err)
	}
	_ = f1.Destroy()
	if _, err := os.Stat(filepath.Join(foreign, "config")); err != nil {
		t.Errorf("Destroy on a fixture that did not Init removed %s: %v", foreign, err)
	}

	// A fixture whose Init created the repo dir removes it.
	own := filepath.Join(g.TempRoot, "own")
	fr := &fakeRunner{respond: resticLike(t, own, "[]")}
	f2, err := NewFixture(fr, own, pw, g)
	if err != nil {
		t.Fatalf("NewFixture(own): %v", err)
	}
	if err := f2.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := os.Stat(own); err != nil {
		t.Fatalf("after Init, repo dir %s missing: %v", own, err)
	}
	if err := f2.Destroy(); err != nil {
		t.Errorf("Destroy(own): %v", err)
	}
	if _, err := os.Stat(own); !os.IsNotExist(err) {
		t.Errorf("after Destroy, repo dir %s still exists (stat err=%v)", own, err)
	}

	if _, err := os.Stat(sibling); err != nil {
		t.Errorf("sibling file in TempRoot did not survive Destroy: %v", err)
	}
}

func TestToolVersions(t *testing.T) {
	ctx := context.Background()
	g := newTestGuard(t)
	repo := filepath.Join(g.TempRoot, "repo")
	pw, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	const resticOut = "restic 0.19.0 compiled with go1.24.5 on darwin/arm64\n"
	const rcloneOut = "rclone v1.68.2\n- os/version: darwin 15.0 (64 bit)\n"

	t.Run("both present", func(t *testing.T) {
		fr := &fakeRunner{respond: func(name string, _ []string) ([]byte, error) {
			switch name {
			case "restic":
				return []byte(resticOut), nil
			case "rclone":
				return []byte(rcloneOut), nil
			}
			return nil, exec.ErrNotFound
		}}
		f, err := NewFixture(fr, repo, pw, g)
		if err != nil {
			t.Fatalf("NewFixture: %v", err)
		}
		restic, rclone, err := f.ToolVersions(ctx)
		if err != nil {
			t.Fatalf("ToolVersions: %v", err)
		}
		if restic != "0.19.0" {
			t.Errorf("restic version = %q, want %q", restic, "0.19.0")
		}
		if rclone != "rclone v1.68.2" {
			t.Errorf("rclone version = %q, want %q", rclone, "rclone v1.68.2")
		}
	})

	t.Run("rclone not on PATH", func(t *testing.T) {
		fr := &fakeRunner{respond: func(name string, _ []string) ([]byte, error) {
			if name == "restic" {
				return []byte(resticOut), nil
			}
			return nil, &exec.Error{Name: name, Err: exec.ErrNotFound}
		}}
		f, err := NewFixture(fr, repo, pw, g)
		if err != nil {
			t.Fatalf("NewFixture: %v", err)
		}
		restic, rclone, err := f.ToolVersions(ctx)
		if err != nil {
			t.Fatalf("ToolVersions with rclone absent: expected nil error, got %v", err)
		}
		if restic != "0.19.0" {
			t.Errorf("restic version = %q, want %q", restic, "0.19.0")
		}
		if rclone != "" {
			t.Errorf("rclone version = %q, want empty when rclone is not on PATH", rclone)
		}
	})
}
