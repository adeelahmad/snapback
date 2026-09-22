package resolver

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const (
	resolverPkg = "github.com/adeelahmad/snapback/internal/resolver"
	rawpathPkg  = "github.com/adeelahmad/snapback/internal/rawpath"
	providerPkg = "github.com/adeelahmad/snapback/internal/provider"
	pathutilPkg = "github.com/adeelahmad/snapback/internal/pathutil"
)

// goList runs "go list" with args and returns the non-empty output lines.
func goList(t *testing.T, args ...string) []string {
	t.Helper()
	out, err := exec.Command("go", append([]string{"list"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("go list %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	var lines []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// isStdlib reports whether an import path belongs to the standard library.
func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

func TestDepsPureNoFilesystem(t *testing.T) {
	tests := []struct {
		pkg           string
		allowedNonStd []string
	}{
		{pkg: resolverPkg, allowedNonStd: []string{providerPkg, pathutilPkg}},
		{pkg: rawpathPkg},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			deps := goList(t, "-deps", tt.pkg)
			if len(deps) == 0 || !slices.Contains(deps, tt.pkg) {
				t.Fatalf("go list -deps %s = %q, want a non-empty list containing the package", tt.pkg, deps)
			}
			for _, d := range deps {
				for _, banned := range []string{"hanwen/go-fuse", "internal/mount", "golang.org/x/text"} {
					if strings.Contains(d, banned) {
						t.Errorf("go list -deps %s contains %q, want no %q dependency", tt.pkg, d, banned)
					}
				}
			}

			imports := goList(t, "-f", `{{join .Imports "\n"}}`, tt.pkg)
			for _, imp := range imports {
				switch imp {
				case "os", "os/exec", "io/fs", "syscall":
					t.Errorf("%s imports %q, want no filesystem or process import", tt.pkg, imp)
				}
				if !isStdlib(imp) && !slices.Contains(tt.allowedNonStd, imp) {
					t.Errorf("%s imports %q, want only stdlib plus %q", tt.pkg, imp, tt.allowedNonStd)
				}
			}
		})
	}
}
