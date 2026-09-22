package seed

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// mk creates each root-relative directory path under root.
func mk(t *testing.T, root string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) = %v", rel, err)
		}
	}
}

// under reports whether dir equals base or lies below it, component-wise.
func under(dir, base string) bool {
	return dir == base || strings.HasPrefix(dir, base+string(filepath.Separator))
}

func TestPlanPathDepthLimited(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "a/b/c/d", "x")

	got, err := PlanPath(r, "", 2, nil)
	if err != nil {
		t.Fatalf("PlanPath(%q, \"\", 2, nil) error = %v, want nil", r, err)
	}
	want := []string{r, filepath.Join(r, "a"), filepath.Join(r, "a", "b"), filepath.Join(r, "x")}
	if !slices.Equal(got.Dirs, want) {
		t.Errorf("PlanPath(%q, \"\", 2, nil).Dirs = %q, want %q", r, got.Dirs, want)
	}
	if got.Count != 4 {
		t.Errorf("PlanPath(%q, \"\", 2, nil).Count = %d, want 4", r, got.Count)
	}
	if slices.Contains(got.Dirs, filepath.Join(r, "a", "b", "c")) {
		t.Errorf("PlanPath(%q, \"\", 2, nil).Dirs contains a/b/c beyond max depth", r)
	}
}

func TestPlanPathDepthZeroIsSeedOnly(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "p/q")
	want := []string{filepath.Join(r, "p")}

	for _, seed := range []string{"p", filepath.Join(r, "p")} {
		got, err := PlanPath(r, seed, 0, nil)
		if err != nil {
			t.Errorf("PlanPath(%q, %q, 0, nil) error = %v, want nil", r, seed, err)
			continue
		}
		if !slices.Equal(got.Dirs, want) {
			t.Errorf("PlanPath(%q, %q, 0, nil).Dirs = %q, want %q", r, seed, got.Dirs, want)
		}
		if got.Count != 1 {
			t.Errorf("PlanPath(%q, %q, 0, nil).Count = %d, want 1", r, seed, got.Count)
		}
	}
}

func TestPlanPathDefaultExclusions(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "src", ".git/objects", "node_modules/pkg", "target", "build", "dist",
		"__pycache__", ".cache/x", "Library/Caches/app", "Library/Prefs", "src/node_modules/y", "gitish")

	got, err := PlanPath(r, "", 5, nil)
	if err != nil {
		t.Fatalf("PlanPath(%q, \"\", 5, nil) error = %v, want nil", r, err)
	}
	for _, rel := range []string{"src", "Library", "Library/Prefs", "gitish"} {
		if want := filepath.Join(r, rel); !slices.Contains(got.Dirs, want) {
			t.Errorf("PlanPath(%q, \"\", 5, nil).Dirs = %q, want it to contain %q", r, got.Dirs, want)
		}
	}
	excluded := []string{".git", "node_modules", "src/node_modules", "target", "build", "dist",
		"__pycache__", ".cache", "Library/Caches"}
	for _, dir := range got.Dirs {
		for _, rel := range excluded {
			if under(dir, filepath.Join(r, rel)) {
				t.Errorf("PlanPath(%q, \"\", 5, nil).Dirs contains %q, want nothing under default exclusion %q", r, dir, rel)
			}
		}
	}
}

func TestPlanPathExtraExcludes(t *testing.T) {
	r := t.TempDir()
	mk(t, r, "keep", "skip/deep", "skipper", "state/cache")
	excludes := []string{"skip", filepath.Join(r, "state")}

	got, err := PlanPath(r, "", 5, excludes)
	if err != nil {
		t.Fatalf("PlanPath(%q, \"\", 5, %q) error = %v, want nil", r, excludes, err)
	}
	for _, rel := range []string{"keep", "skipper"} {
		if want := filepath.Join(r, rel); !slices.Contains(got.Dirs, want) {
			t.Errorf("PlanPath(%q, \"\", 5, %q).Dirs = %q, want it to contain %q", r, excludes, got.Dirs, want)
		}
	}
	for _, dir := range got.Dirs {
		for _, base := range []string{filepath.Join(r, "skip"), filepath.Join(r, "state")} {
			if under(dir, base) {
				t.Errorf("PlanPath(%q, \"\", 5, %q).Dirs contains %q, want nothing under %q", r, excludes, dir, base)
			}
		}
	}
}

func TestPlanPathNoSymlinkFollow(t *testing.T) {
	o := t.TempDir()
	mk(t, o, "leak")
	r := t.TempDir()
	mk(t, r, "real")
	if err := os.Symlink(o, filepath.Join(r, "out")); err != nil {
		t.Fatalf("Symlink(out) = %v", err)
	}
	if err := os.Symlink(filepath.Join(r, "real"), filepath.Join(r, "loop")); err != nil {
		t.Fatalf("Symlink(loop) = %v", err)
	}

	got, err := PlanPath(r, "", 5, nil)
	if err != nil {
		t.Fatalf("PlanPath(%q, \"\", 5, nil) error = %v, want nil", r, err)
	}
	want := []string{r, filepath.Join(r, "real")}
	if !slices.Equal(got.Dirs, want) {
		t.Errorf("PlanPath(%q, \"\", 5, nil).Dirs = %q, want %q", r, got.Dirs, want)
	}
	for _, dir := range got.Dirs {
		if under(dir, o) || strings.Contains(dir, "/out") || strings.Contains(dir, "/loop") {
			t.Errorf("PlanPath(%q, \"\", 5, nil).Dirs contains %q, want no symlinked path", r, dir)
		}
	}
}

func TestPlanPathSeedOutsideRootRejected(t *testing.T) {
	r := t.TempDir()
	o := t.TempDir()

	for _, seed := range []string{o, "../x"} {
		got, err := PlanPath(r, seed, 1, nil)
		if err == nil {
			t.Errorf("PlanPath(%q, %q, 1, nil) error = nil, want %s", r, seed, errcode.InvalidConfig)
		} else if code := errcode.Of(err); code != errcode.InvalidConfig {
			t.Errorf("errcode.Of(PlanPath(%q, %q, 1, nil)) = %s, want %s", r, seed, code, errcode.InvalidConfig)
		}
		if got.Count != 0 {
			t.Errorf("PlanPath(%q, %q, 1, nil).Count = %d, want 0", r, seed, got.Count)
		}
	}
}
