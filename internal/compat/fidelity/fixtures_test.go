package fidelity

import (
	"io/fs"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestFixturesCoverDistinctMetadata(t *testing.T) {
	var regular []Meta
	for _, m := range Fixtures() {
		if m.Mode.IsRegular() {
			regular = append(regular, m)
		}
	}
	if len(regular) < 6 {
		t.Fatalf("Fixtures() has %d regular files, want >= 6", len(regular))
	}

	seen := make(map[time.Time]string)
	var subSecond, pre2000 bool
	y2k := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, m := range regular {
		key := m.MTime.UTC()
		if other, dup := seen[key]; dup {
			t.Errorf("Fixtures() %q and %q share mtime %v, want pairwise-distinct mtimes", other, m.Path, m.MTime)
		}
		seen[key] = m.Path
		if m.MTime.Nanosecond() != 0 {
			subSecond = true
		}
		if m.MTime.Before(y2k) {
			pre2000 = true
		}
	}
	if !subSecond {
		t.Errorf("Fixtures() has no mtime with non-zero Nanosecond(), want one sub-second value")
	}
	if !pre2000 {
		t.Errorf("Fixtures() has no mtime before %v, want one pre-2000 value", y2k)
	}

	var zero, small, large bool
	for _, m := range regular {
		switch {
		case m.Size == 0:
			zero = true
		case m.Size > 0 && m.Size < 64<<10:
			small = true
		case m.Size >= 2<<20:
			large = true
		}
	}
	sizeChecks := []struct {
		name string
		got  bool
	}{
		{"size 0", zero},
		{"size in (0, 64KiB)", small},
		{"size >= 2MiB", large},
	}
	for _, c := range sizeChecks {
		if !c.got {
			t.Errorf("Fixtures() has no regular file with %s, want one", c.name)
		}
	}

	modes := make(map[fs.FileMode]bool)
	for _, m := range regular {
		modes[m.Mode.Perm()] = true
	}
	for _, want := range []fs.FileMode{0o644, 0o600, 0o755, 0o444} {
		if !modes[want] {
			t.Errorf("Fixtures() has no regular file with mode %#o, want one", want)
		}
	}
}

func TestFixturesIncludeSymlinkSpaceUnicode(t *testing.T) {
	fixtures := Fixtures()
	paths := make(map[string]bool)
	for _, m := range fixtures {
		paths[m.Path] = true
	}

	var symlink, space, unicodeName, nested bool
	for _, m := range fixtures {
		if m.Mode&fs.ModeSymlink != 0 && m.LinkTarget == "small.txt" {
			if !paths[m.LinkTarget] {
				t.Errorf("Fixtures() symlink %q targets %q, which is not itself a fixture", m.Path, m.LinkTarget)
			}
			symlink = true
		}
		if strings.Contains(m.Path, " ") {
			space = true
		}
		if utf8.ValidString(m.Path) && hasNonASCII(m.Path) {
			unicodeName = true
		}
		if strings.Contains(m.Path, "/") {
			nested = true
		}
	}

	checks := []struct {
		name string
		got  bool
	}{
		{"symlink with LinkTarget small.txt", symlink},
		{"path containing a space", space},
		{"valid UTF-8 path with a non-ASCII rune", unicodeName},
		{"path containing /", nested},
	}
	for _, c := range checks {
		if !c.got {
			t.Errorf("Fixtures() has no %s, want one", c.name)
		}
	}
}

func hasNonASCII(s string) bool {
	for _, r := range s {
		if r >= utf8.RuneSelf {
			return true
		}
	}
	return false
}
