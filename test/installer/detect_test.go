package installer

import (
	"os"
	"strings"
	"testing"
)

func TestAssetNameAllTargets(t *testing.T) {
	cases := []struct {
		osName, arch, asset string
	}{
		{"Linux", "x86_64", "snapback_linux_amd64.tar.gz"},
		{"Linux", "aarch64", "snapback_linux_arm64.tar.gz"},
		{"Darwin", "x86_64", "snapback_darwin_amd64.tar.gz"},
		{"Darwin", "arm64", "snapback_darwin_arm64.tar.gz"},
		{"Linux", "armv7l", "snapback_linux_arm_unverified.tar.gz"},
		{"Linux", "mips", "snapback_linux_mips_unverified.tar.gz"},
		{"Linux", "mipsel", "snapback_linux_mipsle_unverified.tar.gz"},
	}
	for _, tc := range cases {
		t.Run(tc.osName+"/"+tc.arch, func(t *testing.T) {
			stdout, stderr, code := runInstaller(t, dryRunEnv(tc.osName, tc.arch))
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}
			if !strings.Contains(stdout, tc.asset) {
				t.Errorf("stdout does not contain %q; stdout=%q", tc.asset, stdout)
			}
		})
	}
}

func TestArchAliases(t *testing.T) {
	cases := []struct {
		arch, want string
	}{
		{"amd64", "linux_amd64"},
		{"arm64", "linux_arm64"},
		{"armv6l", "linux_arm_unverified"},
	}
	for _, tc := range cases {
		t.Run(tc.arch, func(t *testing.T) {
			stdout, stderr, code := runInstaller(t, dryRunEnv("Linux", tc.arch))
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}
			want := "snapback_" + tc.want + ".tar.gz"
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout does not contain %q; stdout=%q", want, stdout)
			}
		})
	}
}

func TestUnsupportedPlatformFails(t *testing.T) {
	cases := []struct {
		osName, arch string
	}{
		{"Windows_NT", "x86_64"},
		{"MINGW64_NT", "x86_64"},
		{"FreeBSD", "amd64"},
		{"Linux", "riscv64"},
		{"Linux", "i686"},
		{"Linux", "s390x"},
		{"Darwin", "mips"},
	}
	for _, tc := range cases {
		t.Run(tc.osName+"/"+tc.arch, func(t *testing.T) {
			installDir := t.TempDir()
			env := dryRunEnv(tc.osName, tc.arch)
			env["SNAPBACK_INSTALL_DIR"] = installDir

			_, stderr, code := runInstaller(t, env)

			if code == 0 {
				t.Errorf("exit code = 0, want non-zero")
			}
			if !strings.Contains(stderr, "unsupported") {
				t.Errorf("stderr does not contain %q; stderr=%q", "unsupported", stderr)
			}
			entries, err := os.ReadDir(installDir)
			if err != nil {
				t.Fatalf("read install dir: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("install dir has %d entries, want 0", len(entries))
			}
		})
	}
}

func TestUnverifiedWarning(t *testing.T) {
	cases := []struct {
		osName, arch string
		wantWarning  bool
	}{
		{"Linux", "armv7l", true},
		{"Linux", "mips", true},
		{"Linux", "mipsel", true},
		{"Linux", "x86_64", false},
		{"Darwin", "arm64", false},
	}
	for _, tc := range cases {
		t.Run(tc.osName+"/"+tc.arch, func(t *testing.T) {
			stdout, stderr, code := runInstaller(t, dryRunEnv(tc.osName, tc.arch))
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
			}
			combined := stdout + stderr
			if got := strings.Contains(combined, "unverified"); got != tc.wantWarning {
				t.Errorf("output contains %q = %v, want %v; output=%q", "unverified", got, tc.wantWarning, combined)
			}
		})
	}
}
