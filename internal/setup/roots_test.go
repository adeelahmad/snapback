package setup

import (
	"errors"
	"reflect"
	"testing"
)

func TestApplyRoots(t *testing.T) {
	cwd := func(dir string) func() (string, error) {
		return func() (string, error) { return dir, nil }
	}
	failing := func() (string, error) { return "", errors.New("boom") }

	tests := []struct {
		name        string
		given       []string
		getwd       func() (string, error)
		excluded    []string
		tempDir     string
		wantRoots   []string
		wantReasons []string
	}{
		{
			name:      "given absolute paths are cleaned",
			given:     []string{"/work/project/", "/data/./photos"},
			getwd:     cwd("/work"),
			wantRoots: []string{"/work/project", "/data/photos"},
		},
		{
			name:      "given relative paths resolve against the working directory",
			given:     []string{"project", "../shared"},
			getwd:     cwd("/work/home"),
			wantRoots: []string{"/work/home/project", "/work/shared"},
		},
		{
			name:      "no given paths default to the working directory",
			getwd:     cwd("/work/project"),
			wantRoots: []string{"/work/project"},
		},
		{
			name:      "the filesystem root and the home root are allowed",
			given:     []string{"/", "/home/adeel"},
			getwd:     cwd("/work"),
			wantRoots: []string{"/", "/home/adeel"},
		},
		{
			name:        "a root under an excluded path is refused",
			given:       []string{"/work/project", "/var/lib/snapback/repo"},
			getwd:       cwd("/work"),
			excluded:    []string{"/var/lib/snapback/"},
			wantRoots:   []string{"/work/project"},
			wantReasons: []string{"root /var/lib/snapback/repo: inside /var/lib/snapback"},
		},
		{
			name:        "a root equal to an excluded path is refused",
			given:       []string{"/var/lib/snapback"},
			getwd:       cwd("/work"),
			excluded:    []string{"/var/lib/snapback"},
			wantReasons: []string{"root /var/lib/snapback: inside /var/lib/snapback"},
		},
		{
			name:        "a refused working directory leaves no roots",
			getwd:       cwd("/var/lib/snapback/state"),
			excluded:    []string{"/var/lib/snapback"},
			wantReasons: []string{"root /var/lib/snapback/state: inside /var/lib/snapback"},
		},
		{
			name:        "a root under the temporary directory is refused",
			given:       []string{"/work/project", "/scratch/tmp/run-1"},
			getwd:       cwd("/work"),
			tempDir:     "/scratch/tmp",
			wantRoots:   []string{"/work/project"},
			wantReasons: []string{"root /scratch/tmp/run-1: inside the temporary directory"},
		},
		{
			name:      "an empty temporary directory disables the temporary check",
			given:     []string{"/scratch/tmp/run-1"},
			getwd:     cwd("/work"),
			wantRoots: []string{"/scratch/tmp/run-1"},
		},
		{
			name:        "a working directory error leaves no roots",
			getwd:       failing,
			wantReasons: []string{"cwd: boom"},
		},
		{
			name:      "a sibling of an excluded path is kept",
			given:     []string{"/var/lib/snapback-data"},
			getwd:     cwd("/work"),
			excluded:  []string{"/var/lib/snapback"},
			wantRoots: []string{"/var/lib/snapback-data"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res Result
			applyRoots(&res, test.given, test.getwd, test.excluded, test.tempDir)

			if !reflect.DeepEqual(res.Roots, test.wantRoots) {
				t.Errorf("applyRoots(%v).Roots = %v, want %v", test.given, res.Roots, test.wantRoots)
			}
			if !reflect.DeepEqual(res.Reasons, test.wantReasons) {
				t.Errorf("applyRoots(%v).Reasons = %v, want %v", test.given, res.Reasons, test.wantReasons)
			}
		})
	}
}
