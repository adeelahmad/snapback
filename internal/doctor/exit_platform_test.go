package doctor

import (
	"reflect"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

func TestApplyPlatform(t *testing.T) {
	const reason = "not applicable on macOS in v0.1"

	linuxOnly := []Check{
		{Name: "fuse_device", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "/dev/fuse is not available", Fix: "install the fuse3 package"},
		{Name: "fusermount3", Status: statusFail, Code: errcode.PrereqMissing,
			Detail: "fusermount3 not found", Fix: "install the fuse3 package"},
		{Name: "service_manager", Status: statusFail, Code: errcode.UnsupportedServiceManager,
			Detail: "no supported service manager detected", Fix: "run snapback run yourself"},
		{Name: "restic", Status: statusOK, Detail: "restic 0.18.0"},
	}
	skipped := []Check{
		{Name: "fuse_device", Status: statusSkip, Code: errcode.PrereqMissing,
			Detail: "/dev/fuse is not available", Fix: reason},
		{Name: "fusermount3", Status: statusSkip, Code: errcode.PrereqMissing,
			Detail: "fusermount3 not found", Fix: reason},
		{Name: "service_manager", Status: statusSkip, Code: errcode.UnsupportedServiceManager,
			Detail: "no supported service manager detected", Fix: reason},
		{Name: "restic", Status: statusOK, Detail: "restic 0.18.0"},
	}

	tests := []struct {
		name     string
		in       []Check
		goos     string
		strict   bool
		want     []Check
		wantExit int
	}{
		{
			name:     "darwin skips inapplicable checks",
			in:       linuxOnly,
			goos:     "darwin",
			want:     skipped,
			wantExit: 0,
		},
		{
			name:     "darwin strict keeps failures",
			in:       linuxOnly,
			goos:     "darwin",
			strict:   true,
			want:     linuxOnly,
			wantExit: 1,
		},
		{
			name:     "linux is untouched",
			in:       linuxOnly,
			goos:     "linux",
			want:     linuxOnly,
			wantExit: 1,
		},
		{
			name: "darwin keeps applicable failures",
			in: []Check{
				{Name: "fuse_device", Status: statusFail, Detail: "/dev/fuse is not available", Fix: "install"},
				{Name: "config", Status: statusFail, Detail: "no configuration", Fix: "fix the config"},
			},
			goos: "darwin",
			want: []Check{
				{Name: "fuse_device", Status: statusSkip, Detail: "/dev/fuse is not available", Fix: reason},
				{Name: "config", Status: statusFail, Detail: "no configuration", Fix: "fix the config"},
			},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyPlatform(tt.in, tt.goos, tt.strict)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyPlatform(%s, strict=%v) = %v, want %v", tt.goos, tt.strict, got, tt.want)
			}
			if code := exitCode(got); code != tt.wantExit {
				t.Errorf("exitCode(applyPlatform(%s, strict=%v)) = %d, want %d",
					tt.goos, tt.strict, code, tt.wantExit)
			}
		})
	}
}

func TestApplyPlatformDoesNotMutateInput(t *testing.T) {
	in := []Check{{Name: "fusermount3", Status: statusFail, Fix: "install the fuse3 package"}}
	applyPlatform(in, "darwin", false)
	want := Check{Name: "fusermount3", Status: statusFail, Fix: "install the fuse3 package"}
	if in[0] != want {
		t.Errorf("applyPlatform mutated its input = %v, want %v", in[0], want)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		in   []Check
		want int
	}{
		{name: "empty", in: nil, want: 0},
		{
			name: "skips and warnings do not fail",
			in: []Check{
				{Name: "fuse_device", Status: statusSkip},
				{Name: "daemon_socket", Status: statusWarn},
				{Name: "on_access", Status: statusUnavailable},
				{Name: "restic", Status: statusOK},
			},
			want: 0,
		},
		{
			name: "one failure fails",
			in:   []Check{{Name: "restic", Status: statusOK}, {Name: "config", Status: statusFail}},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCode(tt.in); got != tt.want {
				t.Errorf("exitCode(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
