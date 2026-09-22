package setup

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDetectAbsentSeams(t *testing.T) {
	absentEnv := func(string) string { return "" }
	absentLookPath := func(string) (string, error) { return "", os.ErrNotExist }
	emptyString := func() (string, error) { return "", nil }

	tests := []struct {
		name string
		deps Deps
	}{
		{name: "nil seams", deps: Deps{}},
		{
			name: "empty seams",
			deps: Deps{
				Getenv:   absentEnv,
				LookPath: absentLookPath,
				Getwd:    emptyString,
				Hostname: emptyString,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			got, err := Detect(test.deps)
			if err != nil {
				t.Fatalf("Detect(%s) error = %v, want nil", test.name, err)
			}
			if !reflect.DeepEqual(got, Result{}) {
				t.Errorf("Detect(%s) = %+v, want %+v", test.name, got, Result{})
			}
			if got.Roots != nil {
				t.Errorf("Detect(%s).Roots = %v, want nil", test.name, got.Roots)
			}
			if got.Reasons != nil {
				t.Errorf("Detect(%s).Reasons = %v, want nil", test.name, got.Reasons)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("os.ReadDir(%q) error = %v, want nil", dir, err)
			}
			if len(entries) != 0 {
				t.Errorf("Detect(%s) wrote %d entries into %q, want 0", test.name, len(entries), dir)
			}
		})
	}
}

func TestPackageDoesNotExec(t *testing.T) {
	banned := `"os/` + `exec"`

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf(`os.ReadDir(".") error = %v, want nil`, err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("os.ReadFile(%q) error = %v, want nil", name, err)
		}
		if strings.Contains(string(src), banned) {
			t.Errorf("%s imports os/exec, want the package to run no subprocesses", name)
		}
	}
}

func TestDetectAssembledResult(t *testing.T) {
	wd := t.TempDir()
	sibling := filepath.Join(filepath.Dir(wd), "sibling")
	const (
		repo     = "sftp:backup@example.com:/srv/restic"
		password = "/etc/snapback/restic-password"
		restic   = "/opt/snapback/bin/restic"
		host     = "media-01"
	)

	env := map[string]string{
		"RESTIC_REPOSITORY":    repo,
		"RESTIC_PASSWORD_FILE": password,
	}
	deps := Deps{
		Getenv: func(name string) string { return env[name] },
		LookPath: func(name string) (string, error) {
			if name == "restic" {
				return restic, nil
			}
			return "", os.ErrNotExist
		},
		Getwd:    func() (string, error) { return wd, nil },
		Excluded: []string{sibling},
		Hostname: func() (string, error) { return host, nil },
	}

	got, err := Detect(deps)
	if err != nil {
		t.Fatalf("Detect(assembled) error = %v, want nil", err)
	}
	want := Result{
		RepoURI:        repo,
		CredentialFile: password,
		ResticPath:     restic,
		Roots:          []string{wd},
		Hostname:       host,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect(assembled) = %+v, want %+v", got, want)
	}
}
