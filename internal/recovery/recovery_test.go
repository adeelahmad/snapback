package recovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	historyPoint = "/home/u/Snapshots History"
	backendDir   = "/home/u/.local/state/snapback/backend"
)

func openFixture(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "mountinfo.txt"))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

type unmountRecorder struct {
	points []string
}

func (u *unmountRecorder) unmount(_ context.Context, point string) error {
	u.points = append(u.points, point)
	return nil
}

func mountinfoLine(id int, point, fstype string) string {
	escaped := strings.ReplaceAll(point, " ", `\040`)
	return fmt.Sprintf("%d 22 0:%d / %s ro,nosuid,nodev shared:%d - %s snapback ro\n", 40+id, 40+id, escaped, 20+id, fstype)
}

func TestParseMountinfo(t *testing.T) {
	got, err := ParseMountinfo(openFixture(t))
	if err != nil {
		t.Fatalf("ParseMountinfo(fixture) error = %v, want nil", err)
	}
	want := []Mount{
		{Point: "/", FSType: "ext4", Source: "/dev/sda1"},
		{Point: backendDir + "/repoA", FSType: "fuse.restic", Source: "restic"},
		{Point: historyPoint, FSType: "fuse", Source: "snapback"},
		{Point: "/mnt/nas", FSType: "fuse.sshfs", Source: "nas:/export"},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseMountinfo(fixture) returned %d mounts, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Point != want[i].Point || got[i].FSType != want[i].FSType {
			t.Errorf("ParseMountinfo(fixture)[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestScanUnmountsOnlyOwnedFuse(t *testing.T) {
	rec := &unmountRecorder{}
	in := Input{
		Mountinfo: openFixture(t),
		Owned:     []string{historyPoint, backendDir},
		PIDFile:   filepath.Join(t.TempDir(), "daemon.pid"),
		Alive:     func(int) bool { return false },
		Unmount:   rec.unmount,
		Repair:    func(context.Context) (RepairReport, error) { return RepairReport{}, nil },
	}
	got, err := Scan(context.Background(), in)
	if err != nil {
		t.Fatalf("Scan() error = %v, want nil", err)
	}
	wantUnmounted := []string{historyPoint, backendDir + "/repoA"}
	if !reflect.DeepEqual(got.Unmounted, wantUnmounted) {
		t.Errorf("Scan().Unmounted = %q, want %q (history first)", got.Unmounted, wantUnmounted)
	}
	if !reflect.DeepEqual(rec.points, wantUnmounted) {
		t.Errorf("Unmount calls = %q, want %q", rec.points, wantUnmounted)
	}
	if !reflect.DeepEqual(got.Foreign, []string{"/mnt/nas"}) {
		t.Errorf("Scan().Foreign = %q, want [/mnt/nas]", got.Foreign)
	}
	for _, p := range rec.points {
		if p == "/" {
			t.Errorf("Unmount called for /, want untouched")
		}
	}
}

func TestScanNeverStatsMountPoints(t *testing.T) {
	root := t.TempDir()
	hist := filepath.Join(root, "never", "History")
	backend := filepath.Join(root, "never", "backend")
	info := mountinfoLine(1, hist, "fuse") +
		mountinfoLine(2, backend+"/repoA", "fuse.restic") +
		mountinfoLine(3, backend+"X", "fuse")
	rec := &unmountRecorder{}
	in := Input{
		Mountinfo: strings.NewReader(info),
		Owned:     []string{hist, backend},
		PIDFile:   filepath.Join(root, "daemon.pid"),
		Alive:     func(int) bool { return false },
		Unmount:   rec.unmount,
		Repair:    func(context.Context) (RepairReport, error) { return RepairReport{}, nil },
	}
	got, err := Scan(context.Background(), in)
	if err != nil {
		t.Fatalf("Scan() error = %v, want nil", err)
	}
	wantUnmounted := []string{hist, backend + "/repoA"}
	if !reflect.DeepEqual(rec.points, wantUnmounted) {
		t.Errorf("Unmount calls = %q, want %q (non-existent owned points must still unmount)", rec.points, wantUnmounted)
	}
	if !reflect.DeepEqual(got.Foreign, []string{backend + "X"}) {
		t.Errorf("Scan().Foreign = %q, want [%s] (path-component match, not prefix)", got.Foreign, backend+"X")
	}
}

func TestScanSkipsWhenOwnerAlive(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "daemon.pid")
	if err := os.WriteFile(pidfile, []byte("4242\n"), 0o600); err != nil {
		t.Fatalf("write pidfile: %v", err)
	}
	wantOwned := []string{historyPoint, backendDir + "/repoA"}

	rec := &unmountRecorder{}
	repairs := 0
	in := Input{
		Mountinfo: openFixture(t),
		Owned:     []string{historyPoint, backendDir},
		PIDFile:   pidfile,
		Alive:     func(pid int) bool { return pid == 4242 },
		Unmount:   rec.unmount,
		Repair: func(context.Context) (RepairReport, error) {
			repairs++
			return RepairReport{}, nil
		},
	}
	got, err := Scan(context.Background(), in)
	if err != nil {
		t.Fatalf("Scan(owner alive) error = %v, want nil", err)
	}
	if len(rec.points) != 0 {
		t.Errorf("Scan(owner alive) Unmount calls = %q, want none", rec.points)
	}
	if !reflect.DeepEqual(got.SkippedLive, wantOwned) {
		t.Errorf("Scan(owner alive).SkippedLive = %q, want %q", got.SkippedLive, wantOwned)
	}
	if repairs != 0 {
		t.Errorf("Scan(owner alive) Repair calls = %d, want 0", repairs)
	}

	dead := &unmountRecorder{}
	in.Mountinfo = openFixture(t)
	in.Alive = func(int) bool { return false }
	in.Unmount = dead.unmount
	if _, err := Scan(context.Background(), in); err != nil {
		t.Fatalf("Scan(owner dead) error = %v, want nil", err)
	}
	if !reflect.DeepEqual(dead.points, wantOwned) {
		t.Errorf("Scan(owner dead) Unmount calls = %q, want %q", dead.points, wantOwned)
	}
}
