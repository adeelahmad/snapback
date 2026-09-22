package projection

import (
	"os/exec"
	"strings"
	"testing"
)

const projectionPkg = "github.com/adeelahmad/snapback/internal/projection"

func TestDepsStdlibOnly(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", projectionPkg).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps %s failed: %v\n%s", projectionPkg, err, out)
	}

	var deps []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			deps = append(deps, line)
		}
	}
	if len(deps) == 0 {
		t.Fatal("go list -deps returned no packages")
	}

	foundSelf := false
	for _, dep := range deps {
		if strings.Contains(dep, "hanwen/go-fuse") || strings.Contains(dep, "internal/mount") {
			t.Errorf("projection must not depend on %q", dep)
		}
		if dep == projectionPkg {
			foundSelf = true
			continue
		}
		first, _, _ := strings.Cut(dep, "/")
		if strings.Contains(first, ".") {
			t.Errorf("projection depends on non-standard-library package %q", dep)
		}
	}
	if !foundSelf {
		t.Errorf("go list -deps output does not contain %s:\n%s", projectionPkg, out)
	}
}
