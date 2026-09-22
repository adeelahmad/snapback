package resolver

import (
	"path"
	"strings"
	"testing"
)

func FuzzHistoryTarget(f *testing.F) {
	for _, seed := range []string{"", "/", "a", "/a/b", "//a", "../a", "a/../..", "a\x00", "\xff/a"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, tree string) {
		got, err := HistoryTarget(targetRoot, tree)
		if err != nil {
			if got != "" {
				t.Errorf("HistoryTarget(%q, %q) = %q, %v, want %q on error", targetRoot, tree, got, err, "")
			}
			return
		}
		if !targetUnderRoot(got, targetRoot) {
			t.Errorf("HistoryTarget(%q, %q) = %q, want the root or a path under %q", targetRoot, tree, got, targetRoot+"/")
		}
		if cleaned := path.Clean(got); !targetUnderRoot(cleaned, targetRoot) {
			t.Errorf("path.Clean(HistoryTarget(%q, %q)) = %q, want the root or a path under %q", targetRoot, tree, cleaned, targetRoot+"/")
		}
	})
}

func targetUnderRoot(p, root string) bool {
	return p == root || strings.HasPrefix(p, root+"/")
}
