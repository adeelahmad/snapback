package doctor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func TestRunAllHealthy(t *testing.T) {
	f := healthyProbes(t)

	got := Run(context.Background(), f.cfg, nil, f.probes)

	wantNames := []string{
		"config", "restic", "rclone", "fuse_device", "fusermount3", "password_file",
		"repository:repoA", "mapping:repoA", "service_manager", "inode_headroom",
		"daemon_socket", "on_access",
	}
	var names []string
	for _, c := range got {
		names = append(names, c.Name)
	}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("Run() check names = %v, want %v", names, wantNames)
	}
	for _, c := range got {
		if c.Name == "on_access" {
			if c.Status != "unavailable" || c.Code != errcode.OnAccessUnavailable {
				t.Errorf("on_access = %s/%s, want unavailable/%s", c.Status, c.Code, errcode.OnAccessUnavailable)
			}
			if !strings.Contains(c.Detail+" "+c.Fix, "unavailable in v0.1") {
				t.Errorf("on_access text = %q / %q, want it to contain %q", c.Detail, c.Fix, "unavailable in v0.1")
			}
			continue
		}
		if c.Status != "ok" && c.Status != "skip" {
			t.Errorf("%s status = %q (code %q, detail %q), want ok or skip", c.Name, c.Status, c.Code, c.Detail)
		}
	}
}

func TestFuseMissingExplainsNeverInstalls(t *testing.T) {
	f := healthyProbes(t)
	stat, look := f.probes.Stat, f.probes.LookPath
	f.probes.Stat = func(path string) (fs.FileInfo, error) {
		if path == "/dev/fuse" {
			return nil, fmt.Errorf("stat %s: %w", path, fs.ErrNotExist)
		}
		return stat(path)
	}
	f.probes.LookPath = func(name string) (string, error) {
		if name == "fusermount3" {
			return "", errors.New("executable file not found in $PATH")
		}
		return look(name)
	}

	got := Run(context.Background(), f.cfg, nil, f.probes)

	wantFix := "fuse3"
	if runtime.GOOS == "darwin" {
		wantFix = "macFUSE"
	}
	for _, name := range []string{"fuse_device", "fusermount3"} {
		c := mustCheck(t, got, name)
		if c.Status != "fail" || c.Code != errcode.PrereqMissing {
			t.Errorf("%s = %s/%s, want fail/%s", name, c.Status, c.Code, errcode.PrereqMissing)
		}
		if !strings.Contains(c.Fix, wantFix) {
			t.Errorf("%s fix = %q, want it to contain %q", name, c.Fix, wantFix)
		}
	}
	installers := []string{"apt", "apt-get", "dnf", "yum", "pacman", "zypper", "brew", "sudo"}
	for _, argv := range f.runner.recorded() {
		if len(argv) > 0 && slices.Contains(installers, argv[0]) {
			t.Errorf("Run() invoked package manager %v, want no installer call", argv)
		}
	}
}

func TestResticMissingAndVersion(t *testing.T) {
	f := healthyProbes(t)
	look := f.probes.LookPath
	f.probes.LookPath = func(name string) (string, error) {
		if name == "restic" {
			return "", errors.New("executable file not found in $PATH")
		}
		return look(name)
	}

	missing := mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "restic")
	if missing.Status != "fail" || missing.Code != errcode.PrereqMissing {
		t.Errorf("restic missing = %s/%s, want fail/%s", missing.Status, missing.Code, errcode.PrereqMissing)
	}
	if !strings.Contains(missing.Fix, "restic_binary") {
		t.Errorf("restic missing fix = %q, want it to name %q", missing.Fix, "restic_binary")
	}

	f.probes.LookPath = look
	found := mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "restic")
	if found.Status != "ok" {
		t.Errorf("restic found status = %q, want ok", found.Status)
	}
	if !strings.Contains(found.Detail, "0.18.1") {
		t.Errorf("restic found detail = %q, want it to contain %q", found.Detail, "0.18.1")
	}
}

func TestRepositoryUnreachableIsDistinct(t *testing.T) {
	f := healthyProbes(t)
	f.repo.validateErr = errcode.New(errcode.RepoUnavailable, "validate",
		errors.New("Fatal: unable to open repo at rclone:gd:secret-bucket with password hunter2"))

	got := Run(context.Background(), f.cfg, nil, f.probes)

	repo := mustCheck(t, got, "repository:repoA")
	if repo.Status != "fail" || repo.Code != errcode.RepoUnavailable {
		t.Errorf("repository:repoA = %s/%s, want fail/%s", repo.Status, repo.Code, errcode.RepoUnavailable)
	}
	if m := mustCheck(t, got, "mapping:repoA"); m.Status != "skip" {
		t.Errorf("mapping:repoA status = %q, want skip", m.Status)
	}
	for _, c := range got {
		for _, field := range []string{c.Name, c.Status, string(c.Code), c.Detail, c.Fix} {
			for _, secret := range []string{"hunter2", "secret-bucket"} {
				if strings.Contains(field, secret) {
					t.Errorf("check %s field %q leaks %q", c.Name, field, secret)
				}
			}
		}
	}
}

func TestMappingAbsentAndPasswordMode(t *testing.T) {
	f := healthyProbes(t)
	f.repo.snapshots[0].Paths = []string{"/srv/other"}
	if err := os.Chmod(f.cfg.Repositories[0].PasswordFile, 0o644); err != nil {
		t.Fatalf("chmod password file: %v", err)
	}

	got := Run(context.Background(), f.cfg, nil, f.probes)

	if m := mustCheck(t, got, "mapping:repoA"); m.Status != "fail" || m.Code != errcode.MappingAbsent {
		t.Errorf("mapping:repoA = %s/%s, want fail/%s", m.Status, m.Code, errcode.MappingAbsent)
	}
	pw := mustCheck(t, got, "password_file")
	if pw.Status != "fail" || pw.Code != errcode.PermissionDenied {
		t.Errorf("password_file = %s/%s, want fail/%s", pw.Status, pw.Code, errcode.PermissionDenied)
	}
	if !strings.Contains(pw.Fix, "chmod 600") {
		t.Errorf("password_file fix = %q, want it to contain %q", pw.Fix, "chmod 600")
	}
}

func TestConfigInvalidAndSocketDown(t *testing.T) {
	f := healthyProbes(t)
	loadErr := errcode.New(errcode.InvalidConfig, "load", errors.New("repositories[0].id: required"))

	invalid := Run(context.Background(), nil, loadErr, f.probes)

	if c := mustCheck(t, invalid, "config"); c.Status != "fail" || c.Code != errcode.InvalidConfig {
		t.Errorf("config = %s/%s, want fail/%s", c.Status, c.Code, errcode.InvalidConfig)
	}
	for _, name := range []string{"rclone", "password_file", "inode_headroom"} {
		if c := mustCheck(t, invalid, name); c.Status != "skip" {
			t.Errorf("%s status with invalid config = %q, want skip", name, c.Status)
		}
	}
	for _, c := range invalid {
		if (strings.HasPrefix(c.Name, "repository:") || strings.HasPrefix(c.Name, "mapping:")) && c.Status != "skip" {
			t.Errorf("%s status with invalid config = %q, want skip", c.Name, c.Status)
		}
	}

	f.probes.DialStatus = func(context.Context, string) (string, error) {
		return "", errcode.New(errcode.PrereqMissing, "dial", errors.New("connect: no such file or directory"))
	}
	down := mustCheck(t, Run(context.Background(), f.cfg, nil, f.probes), "daemon_socket")
	if down.Status != "warn" {
		t.Errorf("daemon_socket status = %q, want warn", down.Status)
	}
	if !strings.Contains(down.Fix, "snapback install service") {
		t.Errorf("daemon_socket fix = %q, want it to name %q", down.Fix, "snapback install service")
	}
}

func TestNeverStatsThroughMounts(t *testing.T) {
	f := healthyProbes(t)
	mounts := []string{f.cfg.HistoryMount, f.cfg.BackendMountDir}
	hang := make(chan struct{})
	t.Cleanup(func() { close(hang) })
	stat := f.probes.Stat
	f.probes.Stat = func(path string) (fs.FileInfo, error) {
		for _, m := range mounts {
			if path == m || strings.HasPrefix(path, m+string(filepath.Separator)) {
				t.Errorf("Stat(%q) called on a path at or under mount %q", path, m)
				<-hang
				return nil, errors.New("hung mount")
			}
		}
		return stat(path)
	}
	f.probes.Mountinfo = func() (io.Reader, error) {
		return strings.NewReader(fmt.Sprintf(
			"36 25 0:40 / %s rw,nosuid,nodev,relatime shared:1 - fuse snapback rw,user_id=1000,group_id=1000\n"+
				"37 25 0:41 / %s rw,nosuid,nodev,relatime shared:2 - fuse restic rw,user_id=1000,group_id=1000\n",
			f.cfg.HistoryMount, f.cfg.BackendMountDir)), nil
	}
	f.probes.Statfs = func(string) (uint64, uint64, error) { return 50, 1000, nil }
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan []Check, 1)
	go func() { done <- Run(ctx, f.cfg, nil, f.probes) }()
	var got []Check
	select {
	case got = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run() did not return before the 2s deadline; it blocked on a mount path")
	}

	if c := mustCheck(t, got, "inode_headroom"); c.Status != "warn" {
		t.Errorf("inode_headroom with 5%% free and threshold 0.90 = %q, want warn", c.Status)
	}
}

func TestFuseFixUsesOSRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte("ID=debian\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want no error", path, err)
	}
	saved := osReleasePath
	osReleasePath = path
	t.Cleanup(func() { osReleasePath = saved })

	want := "sudo apt install fuse3"
	if runtime.GOOS == "darwin" {
		want = "install macFUSE yourself; snapback doctor never installs it"
	}
	if got := fuseFix(); got != want {
		t.Errorf("fuseFix() = %q, want %q", got, want)
	}
}
