package resticfx

import (
	"os"
	"strings"
	"testing"
)

type t6ProbeSpec struct {
	env     map[string]string
	onPath  map[string]bool
	exists  map[string]bool
	goos    string
	version string
}

func t6FullSpec(goos string) t6ProbeSpec {
	return t6ProbeSpec{
		env:     map[string]string{"SNAPBACK_FUSE_TESTS": "1"},
		onPath:  map[string]bool{"restic": true, "fusermount3": true},
		exists:  map[string]bool{"/dev/fuse": true, "/Library/Filesystems/macfuse.fs": true},
		goos:    goos,
		version: "restic 0.19.0 compiled with go1.26.4 on " + goos + "/arm64",
	}
}

func (s t6ProbeSpec) probe(t *testing.T) Probe {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatalf("stat temp dir: %v", err)
	}
	return Probe{
		Getenv: func(k string) string { return s.env[k] },
		LookPath: func(name string) (string, error) {
			if s.onPath[name] {
				return "/usr/local/bin/" + name, nil
			}
			return "", os.ErrNotExist
		},
		Stat: func(path string) (os.FileInfo, error) {
			if s.exists[path] {
				return info, nil
			}
			return nil, os.ErrNotExist
		},
		GOOS:             s.goos,
		ResticVersionOut: s.version,
	}
}

func TestMissingPrerequisite(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*t6ProbeSpec)
		goos   string
		want   []string
	}{
		{name: "all present linux", goos: "linux", mutate: func(*t6ProbeSpec) {}},
		{name: "all present darwin", goos: "darwin", mutate: func(*t6ProbeSpec) {}},
		{name: "fuse tests env unset", goos: "linux", mutate: func(s *t6ProbeSpec) { s.env = map[string]string{} }, want: []string{"SNAPBACK_FUSE_TESTS"}},
		{name: "restic not on PATH", goos: "linux", mutate: func(s *t6ProbeSpec) { s.onPath = map[string]bool{"fusermount3": true} }, want: []string{"restic"}},
		{name: "linux without /dev/fuse", goos: "linux", mutate: func(s *t6ProbeSpec) { s.exists = map[string]bool{"/Library/Filesystems/macfuse.fs": true} }, want: []string{"/dev/fuse"}},
		{name: "linux without fusermount3", goos: "linux", mutate: func(s *t6ProbeSpec) { s.onPath = map[string]bool{"restic": true} }, want: []string{"fusermount3"}},
		{name: "darwin without macFUSE", goos: "darwin", mutate: func(s *t6ProbeSpec) { s.exists = map[string]bool{"/dev/fuse": true} }, want: []string{"macfuse"}},
		{name: "wrong restic version", goos: "linux", mutate: func(s *t6ProbeSpec) { s.version = "restic 0.18.1 compiled with go1.24.0 on linux/amd64" }, want: []string{"0.18.1", "0.19.0"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := t6FullSpec(tc.goos)
			tc.mutate(&spec)
			got := MissingPrerequisite(spec.probe(t))
			if len(tc.want) == 0 {
				if got != "" {
					t.Errorf("MissingPrerequisite = %q, want empty", got)
				}
				return
			}
			if got == "" {
				t.Fatalf("MissingPrerequisite = empty, want message naming %v", tc.want)
			}
			for _, w := range tc.want {
				if !strings.Contains(strings.ToLower(got), strings.ToLower(w)) {
					t.Errorf("MissingPrerequisite = %q, want it to name %q", got, w)
				}
			}
		})
	}
}
