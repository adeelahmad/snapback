package installer

import (
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
