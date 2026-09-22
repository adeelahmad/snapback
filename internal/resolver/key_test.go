package resolver

import (
	"regexp"
	"testing"
)

var directoryKeyGoldens = []struct {
	rootID, rel string
	want        string
}{
	{"home", "", "773723c174af82fb04fd1a1fef0948aa"},
	{"home", "docs", "326c6a7c2507a910d556977f1ed01902"},
	{"a", "bc", "ea1cc672b17a5c99d273503a298965ef"},
	{"ab", "c", "c150b536a0d7450f5d040d8dac8f6924"},
	{"home", "docs/\xff\xfe", "d922fbe507f175c3516e6cc48a966536"},
	{"home", "Docs", "7f9ebad3bf84cac3c601d76b8ff8da10"},
}

func TestDirectoryKeyGolden(t *testing.T) {
	for _, tt := range directoryKeyGoldens {
		if got := DirectoryKey(tt.rootID, tt.rel); got != tt.want {
			t.Errorf("DirectoryKey(%q, %q) = %q, want %q", tt.rootID, tt.rel, got, tt.want)
		}
	}
}

func TestDirectoryKeyShapeAndSeparation(t *testing.T) {
	const maxLinkTarget = 60
	hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
	seen := make(map[string]string)
	for _, tt := range directoryKeyGoldens {
		got := DirectoryKey(tt.rootID, tt.rel)
		if !hex32.MatchString(got) {
			t.Errorf("DirectoryKey(%q, %q) = %q, want 32 lowercase hex chars", tt.rootID, tt.rel, got)
		}
		if again := DirectoryKey(tt.rootID, tt.rel); again != got {
			t.Errorf("DirectoryKey(%q, %q) = %q then %q, want stable", tt.rootID, tt.rel, got, again)
		}
		pair := tt.rootID + "|" + tt.rel
		if prev, ok := seen[got]; ok {
			t.Errorf("DirectoryKey(%q) = %q, same as for %q, want distinct keys", pair, got, prev)
		}
		seen[got] = pair
		if link := "/h/roots/home/dirs/" + got; len(link) >= maxLinkTarget {
			t.Errorf("link target %q is %d bytes, want under %d", link, len(link), maxLinkTarget)
		}
	}
}
