package doctor

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParseOSRelease(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantID     string
		wantIDLike []string
	}{
		{
			name:   "quoted id",
			body:   "NAME=\"Ubuntu\"\nID=\"ubuntu\"\nID_LIKE=\"debian\"\n",
			wantID: "ubuntu", wantIDLike: []string{"debian"},
		},
		{
			name:   "bare id and id_like list",
			body:   "ID=rocky\nID_LIKE=\"rhel centos fedora\"\nVERSION_ID=\"9.3\"\n",
			wantID: "rocky", wantIDLike: []string{"rhel", "centos", "fedora"},
		},
		{
			name:   "uppercase value is lowercased",
			body:   "ID=Arch\n",
			wantID: "arch", wantIDLike: nil,
		},
		{
			name:   "comments and blanks ignored",
			body:   "# a comment\n\nID='alpine'\n",
			wantID: "alpine", wantIDLike: nil,
		},
		{
			name:   "no id at all",
			body:   "NAME=\"Weird\"\n",
			wantID: "", wantIDLike: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotIDLike := parseOSRelease(strings.NewReader(tt.body))
			if gotID != tt.wantID || !slices.Equal(gotIDLike, tt.wantIDLike) {
				t.Errorf("parseOSRelease(%q) = %q, %v, want %q, %v",
					tt.body, gotID, gotIDLike, tt.wantID, tt.wantIDLike)
			}
		})
	}
}

func TestPackageCommand(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		idLike []string
		want   string
	}{
		{name: "debian", id: "debian", want: "sudo apt install fuse3"},
		{name: "ubuntu", id: "ubuntu", idLike: []string{"debian"}, want: "sudo apt install fuse3"},
		{name: "raspbian", id: "raspbian", want: "sudo apt install fuse3"},
		{name: "id_like debian", id: "linuxmint", idLike: []string{"ubuntu", "debian"}, want: "sudo apt install fuse3"},
		{name: "fedora", id: "fedora", want: "sudo dnf install fuse3"},
		{name: "rocky", id: "rocky", idLike: []string{"rhel", "centos", "fedora"}, want: "sudo dnf install fuse3"},
		{name: "alma", id: "almalinux", want: "sudo dnf install fuse3"},
		{name: "id_like rhel", id: "oracle", idLike: []string{"rhel"}, want: "sudo dnf install fuse3"},
		{name: "arch", id: "arch", want: "sudo pacman -S fuse3"},
		{name: "manjaro", id: "manjaro", want: "sudo pacman -S fuse3"},
		{name: "alpine", id: "alpine", want: "sudo apk add fuse3"},
		{name: "opensuse leap", id: "opensuse-leap", want: "sudo zypper install fuse3"},
		{name: "id_like suse", id: "sled", idLike: []string{"suse"}, want: "sudo zypper install fuse3"},
		{
			name: "unknown distribution", id: "plan9",
			want: "install the fuse3 package with your distribution's package manager",
		},
		{
			name: "empty id",
			want: "install the fuse3 package with your distribution's package manager",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := packageCommand(tt.id, tt.idLike, "fuse3")
			if got != tt.want {
				t.Errorf("packageCommand(%q, %v, %q) = %q, want %q", tt.id, tt.idLike, "fuse3", got, tt.want)
			}
		})
	}
}

func TestReadOSRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "os-release")
	if err := os.WriteFile(path, []byte("ID=debian\nID_LIKE=\"\"\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v, want nil", path, err)
	}

	gotID, gotIDLike := readOSRelease(path)
	if gotID != "debian" || len(gotIDLike) != 0 {
		t.Errorf("readOSRelease(%q) = %q, %v, want %q, []", path, gotID, gotIDLike, "debian")
	}

	missing := filepath.Join(dir, "absent")
	gotID, gotIDLike = readOSRelease(missing)
	if gotID != "" || len(gotIDLike) != 0 {
		t.Errorf("readOSRelease(%q) = %q, %v, want \"\", []", missing, gotID, gotIDLike)
	}
}
