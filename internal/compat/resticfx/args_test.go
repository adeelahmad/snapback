package resticfx

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	testRepo = "/tmp/x/repo"
	testPW   = "/tmp/x/pw"
	testID   = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

func assertArgs(t *testing.T, name string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}

func TestInitArgs(t *testing.T) {
	got := InitArgs(testRepo, testPW)
	assertArgs(t, "InitArgs", got, []string{"-r", testRepo, "--password-file", testPW, "init"})
}

func TestBackupArgs(t *testing.T) {
	dirs := []string{"/tmp/x/src", "/tmp/x/dir with space;rm -rf ~"}
	for _, d := range dirs {
		got := BackupArgs(testRepo, testPW, d)
		assertArgs(t, "BackupArgs", got,
			[]string{"-r", testRepo, "--password-file", testPW, "backup", "--quiet", d})
	}
}

func TestSnapshotsAndLsArgs(t *testing.T) {
	snaps := SnapshotsArgs(testRepo, testPW)
	assertArgs(t, "SnapshotsArgs", snaps,
		[]string{"-r", testRepo, "--password-file", testPW, "snapshots", "--json"})

	ls := LsArgs(testRepo, testPW, testID)
	assertArgs(t, "LsArgs", ls,
		[]string{"-r", testRepo, "--password-file", testPW, "ls", "--json", testID})
}

func TestMountArgs(t *testing.T) {
	const mnt = "/tmp/x/mnt"
	got := MountArgs(testRepo, testPW, mnt)
	assertArgs(t, "MountArgs", got,
		[]string{"-r", testRepo, "--password-file", testPW, "mount", "--path-template", "ids/%I", mnt})

	idx := -1
	for i, a := range got {
		if a == "--path-template" {
			idx = i
		}
	}
	if idx < 0 || idx+1 >= len(got) || got[idx+1] != "ids/%I" {
		t.Fatalf("MountArgs: --path-template and ids/%%I not adjacent separate elements: %q", got)
	}
	if got[len(got)-1] != mnt {
		t.Fatalf("MountArgs: mountpoint not last: %q", got)
	}
}

func TestArgsNeverContainPassword(t *testing.T) {
	pwFile, err := NewPasswordFile(t.TempDir())
	if err != nil {
		t.Fatalf("NewPasswordFile: %v", err)
	}
	raw, err := os.ReadFile(pwFile)
	if err != nil {
		t.Fatalf("read password file: %v", err)
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		t.Fatalf("password file %s is empty", pwFile)
	}

	builders := map[string]func() []string{
		"InitArgs":      func() []string { return InitArgs(testRepo, pwFile) },
		"BackupArgs":    func() []string { return BackupArgs(testRepo, pwFile, "/tmp/x/src") },
		"SnapshotsArgs": func() []string { return SnapshotsArgs(testRepo, pwFile) },
		"LsArgs":        func() []string { return LsArgs(testRepo, pwFile, testID) },
		"MountArgs":     func() []string { return MountArgs(testRepo, pwFile, "/tmp/x/mnt") },
	}
	for name, build := range builders {
		args := build()
		if len(args) == 0 {
			t.Fatalf("%s returned no args", name)
		}
		hasPath := false
		for _, a := range args {
			if a == pwFile {
				hasPath = true
			}
			if strings.Contains(a, password) {
				t.Errorf("%s: element %q contains the password value", name, a)
			}
		}
		if !hasPath {
			t.Errorf("%s: password-file path %q not present in %q", name, pwFile, args)
		}

		first, second := build(), build()
		if len(first) == 0 || len(second) == 0 {
			t.Fatalf("%s returned empty slice on repeat call", name)
		}
		orig := second[0]
		first[0] = "MUTATED"
		if second[0] != orig {
			t.Errorf("%s: two calls share a backing array (mutation leaked)", name)
		}
	}
}

func TestUnmountCommand(t *testing.T) {
	const mnt = "/tmp/x/mnt"
	cases := []struct {
		goos     string
		wantName string
		wantArgs []string
		wantErr  bool
	}{
		{goos: "darwin", wantName: "umount", wantArgs: []string{mnt}},
		{goos: "linux", wantName: "fusermount3", wantArgs: []string{"-u", mnt}},
		{goos: "windows", wantErr: true},
	}
	for _, tc := range cases {
		name, args, err := UnmountCommand(tc.goos, mnt)
		if tc.wantErr {
			if err == nil {
				t.Errorf("UnmountCommand(%q): want error, got %q %q", tc.goos, name, args)
			}
			continue
		}
		if err != nil {
			t.Errorf("UnmountCommand(%q): unexpected error %v", tc.goos, err)
			continue
		}
		if name != tc.wantName || !reflect.DeepEqual(args, tc.wantArgs) {
			t.Errorf("UnmountCommand(%q) = %q %q, want %q %q", tc.goos, name, args, tc.wantName, tc.wantArgs)
		}
	}
}
