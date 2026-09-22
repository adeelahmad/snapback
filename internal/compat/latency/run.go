package latency

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Samples is the repeat count for measurements that allow repetition.
const Samples = 3

// DataFileCount is the number of generated test files backed up.
const DataFileCount = 100

// DataFileBytes is the size of each generated test file.
const DataFileBytes = 4096

// Clock supplies the wall time used to bracket each timed operation.
type Clock interface {
	Now() time.Time
}

// Mounter starts the long-running restic mount.
type Mounter interface {
	Mount(ctx context.Context, name string, args []string) (Mounted, error)
}

// Mounted is a live mount; listing, reading and unmounting go through it.
type Mounted interface {
	List(ctx context.Context, dir string) ([]string, error)
	Read(ctx context.Context, path string) ([]byte, error)
	Unmount(ctx context.Context) error
}

// Config holds the injected collaborators for one latency run.
type Config struct {
	Remote  string
	Runner  Runner
	Mounter Mounter
	Clock   Clock
}

// Run performs one orchestrated latency measurement against Config.Remote.
func Run(ctx context.Context, cfg Config) (res Result, err error) {
	if err := CheckRemote(cfg.Remote); err != nil {
		return Result{}, err
	}
	if err := checkEmpty(ctx, cfg.Runner); err != nil {
		return Result{}, err
	}
	sc, err := newScratch()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = sc.Close() }()
	defer func() {
		res.RemoteDeleted = purgeAndVerify(ctx, cfg.Runner)
		if !res.RemoteDeleted {
			err = errors.Join(err, errors.New("remote "+AllowedRemote+" not deleted"))
		}
	}()
	res = Result{
		Measurements:  map[string]Stats{},
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		TimestampUTC:  time.Now().UTC(),
		Remote:        AllowedRemote,
		DataFileCount: DataFileCount,
		DataFileBytes: DataFileBytes,
	}
	r := &runner{cfg: cfg, sc: sc, res: &res}
	return res, r.run(ctx)
}

type runner struct {
	cfg Config
	sc  scratch
	res *Result
}

func (r *runner) run(ctx context.Context) error {
	if err := r.versions(ctx); err != nil {
		return err
	}
	id, err := r.prepareRepo(ctx)
	if err != nil {
		return err
	}
	dir := filepath.Join(r.sc.mountDir, "ids", id, r.sc.dataDir)
	err = r.withMount(ctx, func(m Mounted) error { return r.firstMount(ctx, m, id, dir) })
	if err != nil {
		return err
	}
	return r.withMount(ctx, func(m Mounted) error {
		return r.timeListings(ctx, m, dir, "warm_listing_after_restart", Samples)
	})
}

func (r *runner) versions(ctx context.Context) error {
	out, err := r.cfg.Runner.Run(ctx, "restic", []string{"version"})
	if err != nil {
		return fmt.Errorf("restic version: %w", err)
	}
	r.res.ResticVersion = firstLine(out)
	out, err = r.cfg.Runner.Run(ctx, "rclone", []string{"version"})
	if err != nil {
		return fmt.Errorf("rclone version: %w", err)
	}
	r.res.RcloneVersion = firstLine(out)
	return nil
}

func firstLine(out []byte) string {
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return line
}

func (r *runner) prepareRepo(ctx context.Context) (string, error) {
	if err := writeDataFiles(r.sc.dataDir); err != nil {
		return "", err
	}
	if _, err := r.restic(ctx, "init"); err != nil {
		return "", fmt.Errorf("restic init: %w", err)
	}
	if _, err := r.restic(ctx, "backup", r.sc.dataDir); err != nil {
		return "", fmt.Errorf("restic backup: %w", err)
	}
	out, err := r.restic(ctx, "snapshots", "--json")
	if err != nil {
		return "", fmt.Errorf("restic snapshots: %w", err)
	}
	var snaps []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &snaps); err != nil {
		return "", fmt.Errorf("parse restic snapshots: %w", err)
	}
	if len(snaps) == 0 || snaps[len(snaps)-1].ID == "" {
		return "", errors.New("restic snapshots: no snapshot ID")
	}
	return snaps[len(snaps)-1].ID, nil
}

func writeDataFiles(dir string) error {
	content := bytes.Repeat([]byte("s"), DataFileBytes)
	for i := range DataFileCount {
		name := filepath.Join(dir, fmt.Sprintf("file-%03d", i))
		if err := os.WriteFile(name, content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) restic(ctx context.Context, args ...string) ([]byte, error) {
	return r.cfg.Runner.Run(ctx, "restic", r.resticArgs(args))
}

func (r *runner) resticArgs(args []string) []string {
	global := []string{"-r", RepoSpec(), "--password-file", r.sc.passwordFile, "--cache-dir", r.sc.cacheDir}
	return append(global, args...)
}

func (r *runner) withMount(ctx context.Context, fn func(Mounted) error) error {
	args := r.resticArgs([]string{"mount", "--path-template", "ids/%I", r.sc.mountDir})
	m, err := r.cfg.Mounter.Mount(ctx, "restic", args)
	if err != nil {
		return fmt.Errorf("restic mount: %w", err)
	}
	err = fn(m)
	if uerr := m.Unmount(ctx); uerr != nil {
		err = errors.Join(err, fmt.Errorf("unmount: %w", uerr))
	}
	return err
}

func (r *runner) firstMount(ctx context.Context, m Mounted, id, dir string) error {
	if err := r.timeListings(ctx, m, dir, "cold_listing", 1); err != nil {
		return err
	}
	if _, err := r.restic(ctx, "ls", "--json", id); err != nil {
		return fmt.Errorf("restic ls prewarm: %w", err)
	}
	if err := r.timeListings(ctx, m, dir, "warm_prewarmed_listing", Samples); err != nil {
		return err
	}
	return r.timeRead(ctx, m, filepath.Join(dir, "file-000"))
}

func (r *runner) timeListings(ctx context.Context, m Mounted, dir, name string, n int) error {
	samples := make([]time.Duration, n)
	for i := range samples {
		start := r.cfg.Clock.Now()
		if _, err := m.List(ctx, dir); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		samples[i] = r.cfg.Clock.Now().Sub(start)
	}
	return r.record(name, samples)
}

func (r *runner) timeRead(ctx context.Context, m Mounted, path string) error {
	start := r.cfg.Clock.Now()
	if _, err := m.Read(ctx, path); err != nil {
		return fmt.Errorf("cold_first_file_read: %w", err)
	}
	return r.record("cold_first_file_read", []time.Duration{r.cfg.Clock.Now().Sub(start)})
}

func (r *runner) record(name string, samples []time.Duration) error {
	st, err := Summarize(samples)
	if err != nil {
		return err
	}
	r.res.Measurements[name] = st
	return nil
}
