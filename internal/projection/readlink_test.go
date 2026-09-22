package projection

import (
	"fmt"
	"testing"
)

func TestReadlinkExactBytes(t *testing.T) {
	targets := []string{
		"../target",
		"/var/tmp/x",
		"..",
		"a//b/./c/",
		"x/../y",
		"trailing/",
		" spaced ",
		"Ünïcode/ß",
		"../ids/" + fullID,
	}
	var spec Spec
	for i, target := range targets {
		spec.Links = append(spec.Links, Link{Name: fmt.Sprintf("l%d", i), Target: target})
	}
	g, err := Build(spec)
	if err != nil {
		t.Fatalf("Build(spec) error = %v", err)
	}
	if g == nil {
		t.Fatal("Build(spec) returned nil generation")
	}

	for i, want := range targets {
		name := fmt.Sprintf("l%d", i)
		t.Run(fmt.Sprintf("%s=%q", name, want), func(t *testing.T) {
			ino, isDir, found := g.Lookup(RootIno, name)
			if !found || isDir {
				t.Fatalf("Lookup(RootIno, %q) = (%d, isDir=%v, found=%v), want a found symlink", name, ino, isDir, found)
			}
			got, ok := g.Readlink(ino)
			if !ok {
				t.Fatalf("Readlink(%d) found = false, want true", ino)
			}
			if got != want {
				t.Errorf("Readlink(%d) = %q (len %d), want %q (len %d) byte-for-byte", ino, got, len(got), want, len(want))
			}
		})
	}
}

func TestReadlinkNotFound(t *testing.T) {
	g := buildFixture(t)
	relIno, relIsDir, relFound := lookupPath(g, "rel")
	if !relFound || relIsDir {
		t.Fatalf("lookupPath(\"rel\") = (%d, isDir=%v, found=%v), want a found symlink", relIno, relIsDir, relFound)
	}
	docsIno, docsIsDir, docsFound := lookupPath(g, "docs")
	if !docsFound || !docsIsDir {
		t.Fatalf("lookupPath(\"docs\") = (%d, isDir=%v, found=%v), want a found directory", docsIno, docsIsDir, docsFound)
	}

	rows := []struct {
		label  string
		ino    uint64
		target string
		found  bool
	}{
		{"symlink control", relIno, "a/b", true},
		{"directory", docsIno, "", false},
		{"root", RootIno, "", false},
		{"zero", 0, "", false},
		{"unknown", 9999, "", false},
	}
	for _, tc := range rows {
		t.Run(tc.label, func(t *testing.T) {
			target, found := g.Readlink(tc.ino)
			if found != tc.found {
				t.Fatalf("Readlink(%d) found = %v, want %v", tc.ino, found, tc.found)
			}
			if target != tc.target {
				t.Errorf("Readlink(%d) target = %q, want %q", tc.ino, target, tc.target)
			}
		})
	}
}

func TestGenerationsIndependentTargets(t *testing.T) {
	spec := fixtureSpec()
	g1, err := Build(spec)
	if err != nil || g1 == nil {
		t.Fatalf("Build(spec) = (%v, %v), want a generation", g1, err)
	}
	relIdx := -1
	for i, l := range spec.Links {
		if l.Name == "rel" {
			relIdx = i
		}
	}
	if relIdx < 0 {
		t.Fatal("fixtureSpec has no root link \"rel\"")
	}
	spec.Links[relIdx].Target = "changed/target"
	g2, err := Build(spec)
	if err != nil || g2 == nil {
		t.Fatalf("Build(changed spec) = (%v, %v), want a generation", g2, err)
	}

	for _, tc := range []struct {
		label string
		g     *Generation
		want  string
	}{
		{"g1", g1, "a/b"},
		{"g2", g2, "changed/target"},
	} {
		ino, isDir, found := tc.g.Lookup(RootIno, "rel")
		if !found || isDir {
			t.Fatalf("%s: Lookup(RootIno, \"rel\") = (%d, isDir=%v, found=%v), want a found symlink", tc.label, ino, isDir, found)
		}
		got, ok := tc.g.Readlink(ino)
		if !ok {
			t.Fatalf("%s: Readlink(%d) found = false, want true", tc.label, ino)
		}
		if got != tc.want {
			t.Errorf("%s: Readlink(%d) = %q, want %q", tc.label, ino, got, tc.want)
		}
	}
}
