package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFuseFixText(t *testing.T) {
	const (
		macFUSE = "install macFUSE yourself; snapback doctor never installs it"
		generic = "install the fuse3 package with your distribution's package manager"
	)

	dir := t.TempDir()
	write := func(name, body string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile(%q) = %v, want nil", path, err)
		}
		return path
	}

	debian := write("debian", "NAME=\"Debian GNU/Linux\"\nID=debian\n")
	fedora := write("fedora", "ID=fedora\nID_LIKE=\"rhel centos\"\n")
	unknown := write("unknown", "ID=plan9\n")
	missing := filepath.Join(dir, "absent")

	tests := []struct {
		name string
		goos string
		path string
		want string
	}{
		{name: "linux debian", goos: "linux", path: debian, want: "sudo apt install fuse3"},
		{name: "linux fedora", goos: "linux", path: fedora, want: "sudo dnf install fuse3"},
		{name: "linux unknown", goos: "linux", path: unknown, want: generic},
		{name: "linux missing os-release", goos: "linux", path: missing, want: generic},
		{name: "darwin ignores os-release", goos: "darwin", path: debian, want: macFUSE},
		{name: "other platform", goos: "windows", path: debian, want: generic},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := fuseFixText(test.goos, test.path); got != test.want {
				t.Errorf("fuseFixText(%q, %q) = %q, want %q", test.goos, test.path, got, test.want)
			}
		})
	}
}
