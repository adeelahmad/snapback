package version

import (
	"runtime"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		version, commit, target string
		want                    string
	}{
		{"dev", "none", "linux/amd64", "snapback dev (commit none, target linux/amd64)\n"},
		{"v1.2.3", "abc1234", "darwin/arm64", "snapback v1.2.3 (commit abc1234, target darwin/arm64)\n"},
		{"", "", "", "snapback  (commit , target )\n"},
	}
	for _, tt := range tests {
		got := Format(tt.version, tt.commit, tt.target)
		if got != tt.want {
			t.Errorf("Format(%q, %q, %q) = %q, want %q", tt.version, tt.commit, tt.target, got, tt.want)
		}
	}
}

func TestDefaults(t *testing.T) {
	if Version != "dev" {
		t.Errorf("Version = %q, want %q", Version, "dev")
	}
	if Commit != "none" {
		t.Errorf("Commit = %q, want %q", Commit, "none")
	}
	wantTarget := runtime.GOOS + "/" + runtime.GOARCH
	if Target != wantTarget {
		t.Errorf("Target = %q, want %q", Target, wantTarget)
	}
}

func TestStringUsesPackageVars(t *testing.T) {
	oldVersion, oldCommit, oldTarget := Version, Commit, Target
	t.Cleanup(func() { Version, Commit, Target = oldVersion, oldCommit, oldTarget })
	Version, Commit, Target = "v9.9.9", "deadbee", "linux/mipsle"

	got := String()
	if want := Format("v9.9.9", "deadbee", "linux/mipsle"); got != want {
		t.Errorf("String() = %q, want Format(...) = %q", got, want)
	}
	if want := "snapback v9.9.9 (commit deadbee, target linux/mipsle)\n"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
