//go:build integration

package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/mount/gofuse"
	"github.com/adeelahmad/snapback/internal/projection"
)

// Per-command deadlines. A crawler that hits toolTimeout is recorded as
// timed-out rather than failing the run.
const (
	toolTimeout    = 2 * time.Minute
	versionTimeout = 10 * time.Second
)

// crawlEnv is a mounted history catalog with a seeded live tree whose every
// .snapshot entry points into the mount.
type crawlEnv struct {
	counter *Counter
	mnt     string
	tree    string
	root    string
	dirs    []string
	links   []string
	unmount func() error
}

func skipWithoutFUSE(t *testing.T) {
	t.Helper()
	if os.Getenv("SNAPBACK_FUSE_TESTS") != "1" {
		t.Skip("SNAPBACK_FUSE_TESTS not set: FUSE crawler tests skipped")
	}
	switch runtime.GOOS {
	case "linux":
		if _, err := os.Stat("/dev/fuse"); err != nil {
			t.Skip("fuse3 device /dev/fuse not present: FUSE crawler tests skipped")
		}
	case "darwin":
		if _, err := os.Stat("/Library/Filesystems/macfuse.fs"); err != nil {
			t.Skip("macFUSE not installed: /Library/Filesystems/macfuse.fs missing, FUSE crawler tests skipped")
		}
	default:
		t.Skipf("no FUSE support on %s: FUSE crawler tests skipped", runtime.GOOS)
	}
}

// historySpec creates n history directories with a file each under base and
// returns a projection spec whose root holds one symlink per directory.
func historySpec(t *testing.T, base string, n int) projection.Spec {
	t.Helper()
	var spec projection.Spec
	for i := range n {
		name := "snap-" + string(rune('a'+i))
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sub", "old.txt"), []byte("history\n"), 0o644); err != nil {
			t.Fatalf("WriteFile in %s error = %v", dir, err)
		}
		spec.Links = append(spec.Links, projection.Link{Name: name, Target: dir})
	}
	return spec
}

func setupCrawlEnv(t *testing.T) *crawlEnv {
	t.Helper()
	skipWithoutFUSE(t)

	base := t.TempDir()
	gen, err := projection.Build(historySpec(t, filepath.Join(base, "history"), 3))
	if err != nil {
		t.Fatalf("projection.Build(historySpec) error = %v", err)
	}
	mnt := filepath.Join(base, "snapback-crawler-mnt")
	if err := os.Mkdir(mnt, 0o755); err != nil {
		t.Fatalf("Mkdir(%s) error = %v", mnt, err)
	}

	env := &crawlEnv{counter: &Counter{}, mnt: mnt, tree: filepath.Join(base, "tree")}
	a := gofuse.NewAdapter(env.counter)
	if err := a.Mount(mnt, gen); err != nil {
		t.Fatalf("Mount(%s) error = %v", mnt, err)
	}
	mounted := true
	env.unmount = func() error {
		if !mounted {
			return nil
		}
		mounted = false
		return a.Unmount()
	}
	t.Cleanup(func() {
		if err := env.unmount(); err != nil {
			t.Errorf("cleanup Unmount(%s) error = %v", mnt, err)
		}
	})

	env.root = filepath.Join(env.tree, "live")
	env.dirs, env.links, err = Seed(env.root, mnt, DefaultShape())
	if err != nil {
		t.Fatalf("Seed(%s, %s) error = %v", env.root, mnt, err)
	}
	return env
}

// positiveControl lists one .snapshot through the mount and returns the hits.
func positiveControl(t *testing.T, env *crawlEnv) int {
	t.Helper()
	env.counter.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ls", "-L", filepath.Join(env.root, ".snapshot")+"/")
	cmd.Dir = env.tree
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ls -L %s/.snapshot/ error = %v, output %q", env.root, err, out)
	}
	return env.counter.Total()
}

func requireUnmounted(t *testing.T, mnt string) {
	t.Helper()
	entries, err := os.ReadDir(mnt)
	if err != nil {
		t.Fatalf("ReadDir(%s) after unmount error = %v", mnt, err)
	}
	if len(entries) != 0 {
		t.Errorf("mountpoint %s after unmount has %d entries, want empty", mnt, len(entries))
	}
	var mst, pst syscall.Stat_t
	if err := syscall.Lstat(mnt, &mst); err != nil {
		t.Fatalf("Lstat(%s) error = %v", mnt, err)
	}
	if err := syscall.Lstat(filepath.Dir(mnt), &pst); err != nil {
		t.Fatalf("Lstat(%s) error = %v", filepath.Dir(mnt), err)
	}
	if mst.Dev != pst.Dev {
		t.Errorf("mountpoint device = %d, parent device = %d; still a mount", mst.Dev, pst.Dev)
	}
}

func TestCrawlerPositiveControl(t *testing.T) {
	env := setupCrawlEnv(t)

	if got, want := len(env.links), len(env.dirs); got != want {
		t.Errorf("Seed() links = %d, want %d (one per directory)", got, want)
	}
	if got := len(env.dirs); got < minSeedDirs {
		t.Errorf("Seed() dirs = %d, want >= %d", got, minSeedDirs)
	}
	for _, link := range env.links {
		got, err := os.Readlink(link)
		if err != nil {
			t.Fatalf("Readlink(%s) error = %v", link, err)
		}
		if got != env.mnt {
			t.Fatalf("Readlink(%s) = %q, want %q", link, got, env.mnt)
		}
	}

	if got := positiveControl(t, env); got <= 0 {
		t.Errorf("ls -L <dir>/.snapshot/: Counter.Total() = %d, want > 0", got)
	}

	if err := env.unmount(); err != nil {
		t.Fatalf("Unmount(%s) error = %v", env.mnt, err)
	}
	requireUnmounted(t, env.mnt)
}

// toolVersion returns the first line of the tool's version output, or "" when
// the tool has no version flag (BSD find).
func toolVersion(path string, args []string) string {
	ctx, cancel := context.WithTimeout(context.Background(), versionTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		return ""
	}
	first, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(first)
}

// runTool runs one tool row over the seeded tree and returns its evidence.
func runTool(t *testing.T, env *crawlEnv, tool Tool) Row {
	t.Helper()
	dest := t.TempDir()
	argv := tool.Argv(env.root, dest)
	row := Row{Tool: tool.Name, Argv: argv, Follows: tool.Follows}

	path, err := Resolve(tool, exec.LookPath)
	if err != nil {
		row.Status = StatusNotTestedHere
		row.Reason = err.Error()
		return row
	}
	row.Version = toolVersion(path, tool.VersionArgs)

	env.counter.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, argv[1:]...)
	cmd.Dir = env.tree
	start := time.Now()
	runErr := cmd.Run()
	row.Duration = time.Since(start)

	var exitErr *exec.ExitError
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		row.Status = StatusTimedOut
		row.Reason = "exceeded " + toolTimeout.String()
	case runErr != nil && !errors.As(runErr, &exitErr):
		t.Fatalf("run %v error = %v", argv, runErr)
	default:
		row.Status = StatusTested
	}
	if cmd.ProcessState != nil {
		row.ExitCode = cmd.ProcessState.ExitCode()
	}
	hits := env.counter.Total()
	row.Hits = &hits
	row.Ops = make(map[string]int)
	for op, n := range env.counter.ByOp() {
		row.Ops[op.String()] = n
	}
	return row
}

func TestCrawlerToolMatrix(t *testing.T) {
	env := setupCrawlEnv(t)
	report := Report{
		SchemaVersion:       SchemaVersion,
		GOOS:                runtime.GOOS,
		GOARCH:              runtime.GOARCH,
		Seed:                DefaultShape(),
		PositiveControlHits: positiveControl(t, env),
	}
	if report.PositiveControlHits <= 0 {
		t.Fatalf("positive control hits = %d, want > 0", report.PositiveControlHits)
	}

	for _, tool := range Tools() {
		t.Run(tool.Name, func(t *testing.T) {
			row := runTool(t, env, tool)
			report.Rows = append(report.Rows, row)
			if row.Status == StatusNotTestedHere {
				t.Skipf("%s not on PATH: %s", tool.Candidates[0], row.Reason)
			}
		})
	}
	vs := VSCodeRow()
	report.Rows = append(report.Rows, Row{Tool: vs.Name, Argv: vs.Argv, Status: vs.Status, Reason: vs.Reason})

	if got := len(report.Rows); got != len(Tools())+1 {
		t.Fatalf("report rows = %d, want %d", got, len(Tools())+1)
	}
	for _, row := range report.Rows {
		switch {
		case row.Tool == vs.Name:
			if row.Status != StatusNotTestedHere {
				t.Errorf("row %q status = %q, want %q", row.Tool, row.Status, StatusNotTestedHere)
			}
		case row.Status == StatusNotTestedHere:
			if row.Hits != nil {
				t.Errorf("missing tool row %q hits = %d, want nil", row.Tool, *row.Hits)
			}
		case !row.Follows:
			if row.Status != StatusTested {
				t.Errorf("non-following row %q status = %q, want %q", row.Tool, row.Status, StatusTested)
			}
			if row.Hits == nil || *row.Hits != 0 {
				t.Errorf("non-following row %q hits = %v (ops %v), want 0", row.Tool, row.Hits, row.Ops)
			}
		default:
			if row.Hits == nil {
				t.Errorf("following row %q hits = nil, want a recorded count", row.Tool)
			}
		}
	}
	if bad := NonFollowingViolations(report); len(bad) != 0 {
		t.Errorf("NonFollowingViolations(report) = %+v, want none", bad)
	}

	dir := os.Getenv("SNAPBACK_EVIDENCE_DIR")
	if dir == "" {
		return
	}
	if err := WriteEvidence(dir, report); err != nil {
		t.Fatalf("WriteEvidence(%s) error = %v", dir, err)
	}
	name, err := EvidenceFileName(runtime.GOOS)
	if err != nil {
		t.Fatalf("EvidenceFileName(%q) error = %v", runtime.GOOS, err)
	}
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read evidence %s: %v", name, err)
	}
	var got Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode evidence %s: %v", name, err)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("evidence %s Validate() = %v, want nil", name, err)
	}
	if got.PositiveControlHits <= 0 {
		t.Errorf("evidence %s positive_control_hits = %d, want > 0", name, got.PositiveControlHits)
	}
}
