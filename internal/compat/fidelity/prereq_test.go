package fidelity

import (
	"errors"
	"testing"
)

type fakeEnv struct {
	env    map[string]string
	bins   map[string]bool
	exists map[string]bool
}

func (f fakeEnv) getenv(k string) string { return f.env[k] }

func (f fakeEnv) lookPath(name string) (string, error) {
	if f.bins[name] {
		return "/usr/local/bin/" + name, nil
	}
	return "", errors.New("executable file not found in $PATH")
}

func (f fakeEnv) exist(p string) bool { return f.exists[p] }

func allPresent() fakeEnv {
	return fakeEnv{
		env:  map[string]string{"SNAPBACK_FUSE_TESTS": "1"},
		bins: map[string]bool{"restic": true},
		exists: map[string]bool{
			"/dev/fuse":                       true,
			"/Library/Filesystems/macfuse.fs": true,
		},
	}
}

func TestMissingPrereqNamesEach(t *testing.T) {
	noEnv := allPresent()
	noEnv.env = map[string]string{}
	envZero := allPresent()
	envZero.env = map[string]string{"SNAPBACK_FUSE_TESTS": "0"}
	envYes := allPresent()
	envYes.env = map[string]string{"SNAPBACK_FUSE_TESTS": "yes"}
	noRestic := allPresent()
	noRestic.bins = map[string]bool{}
	noDevFuse := allPresent()
	noDevFuse.exists = map[string]bool{"/Library/Filesystems/macfuse.fs": true}
	noMacFUSE := allPresent()
	noMacFUSE.exists = map[string]bool{"/dev/fuse": true}

	tests := []struct {
		name string
		f    fakeEnv
		goos string
		want string
	}{
		{"env unset", noEnv, "linux", "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped"},
		{"env set to 0", envZero, "linux", "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped"},
		{"env set to yes", envYes, "linux", "SNAPBACK_FUSE_TESTS not set: fidelity tests skipped"},
		{"restic missing", noRestic, "darwin", "restic not on PATH: fidelity tests skipped"},
		{"linux without /dev/fuse", noDevFuse, "linux", "fuse3 device /dev/fuse not present"},
		{"darwin without macFUSE", noMacFUSE, "darwin", "macFUSE not installed: /Library/Filesystems/macfuse.fs missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MissingPrereq(tt.f.getenv, tt.f.lookPath, tt.f.exist, tt.goos)
			if got != tt.want {
				t.Errorf("MissingPrereq(%s, %q) = %q, want %q", tt.name, tt.goos, got, tt.want)
			}
		})
	}
}

func TestMissingPrereqEmptyWhenAllPresent(t *testing.T) {
	f := allPresent()
	for _, goos := range []string{"linux", "darwin"} {
		if got := MissingPrereq(f.getenv, f.lookPath, f.exist, goos); got != "" {
			t.Errorf("MissingPrereq(all present, %q) = %q, want \"\"", goos, got)
		}
	}
}
