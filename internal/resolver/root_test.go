package resolver

import (
	"errors"
	"strings"
	"testing"
)

func TestSelectRootLongestWins(t *testing.T) {
	fixtureRootsLocal := []RootSpec{
		{"home", "/home/alex/project"},
		{"nested", "/home/alex/project/docs"},
		{"other", "/srv"},
	}
	tests := []struct {
		dir  string
		want Match
	}{
		{"/home/alex/project", Match{"home", ""}},
		{"/home/alex/project/src/a", Match{"home", "src/a"}},
		{"/home/alex/project/docs", Match{"nested", ""}},
		{"/home/alex/project/docs/api/v1", Match{"nested", "api/v1"}},
		{"/home/alex/project/docs/../src/", Match{"home", "src"}},
		{"/srv/x", Match{"other", "x"}},
	}
	for _, tt := range tests {
		got, err := SelectRoot(fixtureRootsLocal, tt.dir)
		if err != nil || got != tt.want {
			t.Errorf("SelectRoot(roots, %q) = %+v, %v, want %+v, nil", tt.dir, got, err, tt.want)
		}
	}
}

func TestSelectRootRejects(t *testing.T) {
	project := []RootSpec{{"home", "/home/alex/project"}, {"alex", "/home/alex"}}
	ambiguous := []RootSpec{{"a", "/data"}, {"b", "/data/"}}
	tests := []struct {
		name    string
		roots   []RootSpec
		dir     string
		want    Match
		wantErr error
	}{
		{"control", []RootSpec{{"a", "/data"}}, "/data/x", Match{"a", "x"}, nil},
		{"component boundary", []RootSpec{{"alex", "/home/alex"}}, "/home/alexandra", Match{}, ErrOutsideRoots},
		{"outside", project, "/etc", Match{}, ErrOutsideRoots},
		{"relative", project, "relative/path", Match{}, ErrOutsideRoots},
		{"empty", project, "", Match{}, ErrOutsideRoots},
		{"equal cleaned roots", ambiguous, "/data/x", Match{}, ErrAmbiguousRoot},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectRoot(tt.roots, tt.dir)
			if tt.wantErr == nil {
				if err != nil || got != tt.want {
					t.Fatalf("SelectRoot(%v, %q) = %+v, %v, want %+v, nil", tt.roots, tt.dir, got, err, tt.want)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SelectRoot(%v, %q) error = %v, want %v", tt.roots, tt.dir, err, tt.wantErr)
			}
			if !strings.HasPrefix(err.Error(), "resolver:") {
				t.Errorf("SelectRoot(%v, %q) error = %q, want prefix %q", tt.roots, tt.dir, err, "resolver:")
			}
			if got != (Match{}) {
				t.Errorf("SelectRoot(%v, %q) = %+v on error, want zero Match", tt.roots, tt.dir, got)
			}
		})
	}
}

func TestSelectRootPreservesBytes(t *testing.T) {
	roots := []RootSpec{{"r", "/data"}}
	tests := []string{
		"Docs",
		"docs",
		"é",  // precomposed é
		"é", // decomposed é
		"a b",
		"\xff",
	}
	for _, rel := range tests {
		dir := "/data/" + rel
		got, err := SelectRoot(roots, dir)
		want := Match{"r", rel}
		if err != nil || got != want {
			t.Errorf("SelectRoot(roots, %q) = %#v, %v, want %#v, nil", dir, got, err, want)
		}
	}
}

func TestSelectRootSlashRoot(t *testing.T) {
	roots := []RootSpec{{"all", "/"}, {"home", "/home"}}
	tests := []struct {
		dir  string
		want Match
	}{
		{"/home/x", Match{"home", "x"}},
		{"/etc/x", Match{"all", "etc/x"}},
		{"/", Match{"all", ""}},
	}
	for _, tt := range tests {
		got, err := SelectRoot(roots, tt.dir)
		if err != nil || got != tt.want {
			t.Errorf("SelectRoot(roots, %q) = %+v, %v, want %+v, nil", tt.dir, got, err, tt.want)
		}
	}
}
