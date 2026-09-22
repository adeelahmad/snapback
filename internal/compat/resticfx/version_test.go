package resticfx

import (
	"strings"
	"testing"
)

func TestParseResticVersion(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    string
		wantErr bool
	}{
		{name: "darwin 0.19.0", out: "restic 0.19.0 compiled with go1.26.4 on darwin/arm64", want: "0.19.0"},
		{name: "linux 0.18.1", out: "restic 0.18.1 compiled with go1.24.0 on linux/amd64", want: "0.18.1"},
		{name: "empty", out: "", wantErr: true},
		{name: "garbage", out: "garbage", wantErr: true},
		{name: "bare restic", out: "restic", wantErr: true},
		{name: "non-numeric version", out: "restic x.y.z", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResticVersion(tt.out)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseResticVersion(%q) = %q, nil; want error", tt.out, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseResticVersion(%q) error = %v; want %q", tt.out, err, tt.want)
			}
			if got != tt.want {
				t.Fatalf("ParseResticVersion(%q) = %q; want %q", tt.out, got, tt.want)
			}
		})
	}
}

func TestCheckPinnedVersion(t *testing.T) {
	if err := CheckPinnedVersion("restic 0.19.0 compiled with go1.26.4 on darwin/arm64"); err != nil {
		t.Fatalf("CheckPinnedVersion(0.19.0) = %v; want nil", err)
	}

	err := CheckPinnedVersion("restic 0.18.1 compiled with go1.24.0 on linux/amd64")
	if err == nil {
		t.Fatal("CheckPinnedVersion(0.18.1) = nil; want error")
	}
	for _, v := range []string{"0.18.1", "0.19.0"} {
		if !strings.Contains(err.Error(), v) {
			t.Errorf("CheckPinnedVersion(0.18.1) error %q does not mention %q", err.Error(), v)
		}
	}

	if err := CheckPinnedVersion("garbage"); err == nil {
		t.Fatal("CheckPinnedVersion(garbage) = nil; want error")
	}
}
