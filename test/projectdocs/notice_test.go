package projectdocs

import (
	"strings"
	"testing"
)

func TestNoticeListsGoStdlib(t *testing.T) {
	notice := readDoc(t, "NOTICE")
	for _, want := range []string{"Go standard library", "BSD-3-Clause"} {
		if !strings.Contains(notice, want) {
			t.Errorf("NOTICE missing %q", want)
		}
	}
}

func TestNoticeSaysListIsMaintained(t *testing.T) {
	notice := strings.ToLower(readDoc(t, "NOTICE"))
	if want := "updated as dependencies are added"; !strings.Contains(notice, want) {
		t.Errorf("NOTICE missing %q (case-insensitive)", want)
	}
}

func TestNoticeCoversGoModRequires(t *testing.T) {
	notice := readDoc(t, "NOTICE")
	for _, mod := range goModRequires(readDoc(t, "go.mod")) {
		if !strings.Contains(notice, mod) {
			t.Errorf("NOTICE missing go.mod dependency %q", mod)
		}
	}
}

// goModRequires returns the module paths of every require directive, in both
// single-line and block form.
func goModRequires(gomod string) []string {
	var mods []string
	inBlock := false
	for _, raw := range strings.Split(gomod, "\n") {
		line := strings.TrimSpace(raw)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case inBlock && line == ")":
			inBlock = false
		case inBlock:
			if f := strings.Fields(line); len(f) > 0 {
				mods = append(mods, f[0])
			}
		case line == "require (":
			inBlock = true
		case strings.HasPrefix(line, "require "):
			if f := strings.Fields(strings.TrimPrefix(line, "require ")); len(f) > 0 {
				mods = append(mods, f[0])
			}
		}
	}
	return mods
}
