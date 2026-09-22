package setup

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fakeLookPath answers from found and reports os.ErrNotExist for anything else,
// the way exec.LookPath reports a binary that is not on PATH.
func fakeLookPath(found map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if p, ok := found[name]; ok {
			return p, nil
		}
		return "", os.ErrNotExist
	}
}

func TestApplyBinaries(t *testing.T) {
	relative := filepath.Join("bin", "restic")
	absolute, err := filepath.Abs(relative)
	if err != nil {
		t.Fatalf("filepath.Abs(%q) error = %v, want nil", relative, err)
	}

	tests := []struct {
		name        string
		lookPath    func(string) (string, error)
		wantPath    string
		wantRclone  bool
		wantReasons []string
	}{
		{
			name:       "both found",
			lookPath:   fakeLookPath(map[string]string{"restic": "/usr/local/bin/restic", "rclone": "/usr/local/bin/rclone"}),
			wantPath:   "/usr/local/bin/restic",
			wantRclone: true,
		},
		{
			name:       "relative restic path is resolved",
			lookPath:   fakeLookPath(map[string]string{"restic": relative}),
			wantPath:   absolute,
			wantRclone: false,
		},
		{
			name:        "restic missing names one reason",
			lookPath:    fakeLookPath(map[string]string{"rclone": "/usr/local/bin/rclone"}),
			wantPath:    "",
			wantRclone:  true,
			wantReasons: []string{"restic: not found on PATH"},
		},
		{
			name:        "rclone missing is silent",
			lookPath:    fakeLookPath(map[string]string{"restic": "/usr/bin/restic"}),
			wantPath:    "/usr/bin/restic",
			wantRclone:  false,
			wantReasons: nil,
		},
		{
			name:        "neither found",
			lookPath:    fakeLookPath(nil),
			wantPath:    "",
			wantRclone:  false,
			wantReasons: []string{"restic: not found on PATH"},
		},
		{
			name:        "nil seam finds nothing",
			lookPath:    nil,
			wantPath:    "",
			wantRclone:  false,
			wantReasons: []string{"restic: not found on PATH"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res Result
			applyBinaries(&res, test.lookPath)

			if got, want := res.ResticPath, test.wantPath; got != want {
				t.Errorf("applyBinaries(%s).ResticPath = %q, want %q", test.name, got, want)
			}
			if got, want := res.RcloneFound, test.wantRclone; got != want {
				t.Errorf("applyBinaries(%s).RcloneFound = %t, want %t", test.name, got, want)
			}
			if got, want := res.Reasons, test.wantReasons; !reflect.DeepEqual(got, want) {
				t.Errorf("applyBinaries(%s).Reasons = %v, want %v", test.name, got, want)
			}
		})
	}
}
