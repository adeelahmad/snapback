package setup

import (
	"strings"
	"testing"
)

func TestDefaultMountPoint(t *testing.T) {
	tests := []struct {
		name string
		goos string
		home string
		id   string
		want string
	}{
		{
			name: "linux uses /mnt and ignores the home directory",
			goos: "linux",
			home: "/home/ada",
			id:   "home-nas",
			want: "/mnt/home-nas",
		},
		{
			name: "an empty home on linux changes nothing",
			goos: "linux",
			home: "",
			id:   "home-nas",
			want: "/mnt/home-nas",
		},
		{
			name: "darwin keeps the mount under the given home",
			goos: "darwin",
			home: "/Users/ada",
			id:   "home-nas",
			want: "/Users/ada/Library/Application Support/snapback/mounts/home-nas",
		},
		{
			name: "an unknown operating system is treated like linux",
			goos: "plan9",
			home: "/home/ada",
			id:   "home-nas",
			want: "/mnt/home-nas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultMountPoint(tt.goos, tt.home, tt.id)
			if err != nil {
				t.Fatalf("DefaultMountPoint(%q, %q, %q) error = %v, want nil", tt.goos, tt.home, tt.id, err)
			}
			if got != tt.want {
				t.Errorf("DefaultMountPoint(%q, %q, %q) = %q, want %q", tt.goos, tt.home, tt.id, got, tt.want)
			}
		})
	}
}

func TestDefaultMountPointRejectsID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "a separator would escape the parent directory", id: "a/b"},
		{name: "a parent reference", id: ".."},
		{name: "the current directory", id: "."},
		{name: "a NUL byte", id: "a\x00b"},
		{name: "an empty id", id: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, goos := range []string{"linux", "darwin"} {
				got, err := DefaultMountPoint(goos, "/Users/ada", tt.id)
				if err == nil {
					t.Fatalf("DefaultMountPoint(%q, %q, %q) error = nil, want an error", goos, "/Users/ada", tt.id)
				}
				if got != "" {
					t.Errorf("DefaultMountPoint(%q, %q, %q) = %q, want %q on error", goos, "/Users/ada", tt.id, got, "")
				}
				if !strings.Contains(err.Error(), tt.id) {
					t.Errorf("DefaultMountPoint(%q, %q, %q) error = %q, want it to name the id %q", goos, "/Users/ada", tt.id, err, tt.id)
				}
			}
		})
	}
}
