package projection

import (
	"strings"
	"testing"
)

// Kind values mirror mount.KindDir and mount.KindSymlink; projection may not
// import mount, so the contract is pinned by value.
const (
	kindDir     uint8 = 1
	kindSymlink uint8 = 2
)

// fullID is a full 64-hex snapshot ID.
const fullID = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// fixtureSpec returns a fresh spec; root entries are declared out of byte order on purpose.
func fixtureSpec() Spec {
	return Spec{
		Dirs: []Dir{
			{Name: "snapshots", Links: []Link{{Name: fullID, Target: "../ids/" + fullID}}},
			{Name: "docs", Links: []Link{{Name: "readme-link", Target: "../target"}}},
			{Name: "empty"},
		},
		Links: []Link{
			{Name: "rel", Target: "a/b"},
			{Name: "up", Target: "../outside"},
			{Name: "abs", Target: "/var/tmp/x"},
			{Name: "latest", Target: "snapshots/" + fullID},
		},
	}
}

// fixturePaths lists every node of fixtureSpec as a slash-joined path.
func fixturePaths() []string {
	return []string{
		"snapshots", "snapshots/" + fullID,
		"docs", "docs/readme-link",
		"empty",
		"rel", "up", "abs", "latest",
	}
}

// buildFixture builds fixtureSpec and fails the test on error.
func buildFixture(t *testing.T) *Generation {
	t.Helper()
	g, err := Build(fixtureSpec())
	if err != nil {
		t.Fatalf("Build(fixtureSpec()) error = %v", err)
	}
	if g == nil {
		t.Fatal("Build(fixtureSpec()) returned nil generation")
	}
	return g
}

// lookupPath walks a slash-joined path from RootIno via Lookup.
func lookupPath(g *Generation, path string) (ino uint64, kind uint8, found bool) {
	ino = RootIno
	for _, name := range strings.Split(path, "/") {
		ino, kind, found = g.Lookup(ino, name)
		if !found {
			return 0, 0, false
		}
	}
	return ino, kind, true
}

// mustLookupPath is lookupPath that fails the test when the path is missing.
func mustLookupPath(t *testing.T, g *Generation, path string) uint64 {
	t.Helper()
	ino, _, found := lookupPath(g, path)
	if !found {
		t.Fatalf("lookupPath(%q) not found", path)
	}
	return ino
}
