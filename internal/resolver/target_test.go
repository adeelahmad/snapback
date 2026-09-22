package resolver

import (
	"strings"
	"testing"
)

// targetSnapID stands in for the T4 fixture's id1 so T3 does not depend on
// fixture_test.go.
const targetSnapID = "1111111111111111111111111111111111111111111111111111111111111111"

const targetRoot = "/m/ids/" + targetSnapID

func TestHistoryTargetJoins(t *testing.T) {
	tests := []struct {
		root string
		tree string
		want string
	}{
		{targetRoot, "", targetRoot},
		{targetRoot, "/", targetRoot},
		{targetRoot, "home/alex/project", targetRoot + "/home/alex/project"},
		{targetRoot, "/home/alex/project/docs", targetRoot + "/home/alex/project/docs"},
		{targetRoot, "project/a b/Ünïcode", targetRoot + "/project/a b/Ünïcode"},
		{targetRoot, "x/\xff", targetRoot + "/x/\xff"},
		{targetRoot + "/", "a", targetRoot + "/a"},
	}
	for _, tt := range tests {
		got, err := HistoryTarget(tt.root, tt.tree)
		if err != nil {
			t.Errorf("HistoryTarget(%q, %q) returned error %v, want nil", tt.root, tt.tree, err)
			continue
		}
		if got != tt.want {
			t.Errorf("HistoryTarget(%q, %q) = %q, want %q", tt.root, tt.tree, got, tt.want)
		}
	}
}

func TestHistoryTargetRejectsEscape(t *testing.T) {
	// Control row first (M-002): a valid tree path must succeed.
	got, err := HistoryTarget(targetRoot, "a")
	if want := targetRoot + "/a"; err != nil || got != want {
		t.Fatalf("HistoryTarget(%q, %q) = %q, %v, want %q, nil", targetRoot, "a", got, err, want)
	}

	tests := []struct {
		root string
		tree string
	}{
		{targetRoot, "//etc/passwd"},
		{targetRoot, "../x"},
		{targetRoot, "a/../../x"},
		{targetRoot, "a/.."},
		{targetRoot, "./a"},
		{targetRoot, "a/./b"},
		{targetRoot, "a//b"},
		{targetRoot, "a/"},
		{targetRoot, "a\x00b"},
		{"", "a"},
		{"relative/root", "a"},
	}
	for _, tt := range tests {
		got, err := HistoryTarget(tt.root, tt.tree)
		if err == nil {
			t.Errorf("HistoryTarget(%q, %q) = %q, nil, want error", tt.root, tt.tree, got)
			continue
		}
		if got != "" {
			t.Errorf("HistoryTarget(%q, %q) = %q, want %q on error", tt.root, tt.tree, got, "")
		}
		if !strings.HasPrefix(err.Error(), "resolver:") {
			t.Errorf("HistoryTarget(%q, %q) error = %q, want prefix %q", tt.root, tt.tree, err, "resolver:")
		}
	}
}
