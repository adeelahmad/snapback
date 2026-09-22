package links

import (
	"errors"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/resolver"
)

var hexKey = regexp.MustCompile(`^[0-9a-f]{32}$`)

// newPlaceFixture returns a policy with one root "home" at r, history mount h
// and r/state excluded.
func newPlaceFixture(t *testing.T) (pol Policy, r, h string) {
	t.Helper()
	r = t.TempDir()
	h = t.TempDir()
	pol = Policy{
		LinkName:     ".snapshot",
		HistoryMount: h,
		Roots:        []resolver.RootSpec{{ID: "home", LocalPath: r}},
		Excluded:     []string{filepath.Join(r, "state")},
	}
	return pol, r, h
}

func TestPlaceComputesTarget(t *testing.T) {
	pol, r, h := newPlaceFixture(t)
	dir := filepath.Join(r, "docs", "proj")

	got, err := place(pol, dir)
	if err != nil {
		t.Fatalf("place(%q) error = %v, want nil", dir, err)
	}

	wantKey := resolver.DirectoryKey("home", "docs/proj")
	if got.rootID != "home" {
		t.Errorf("place(%q).rootID = %q, want %q", dir, got.rootID, "home")
	}
	if got.rel != "docs/proj" {
		t.Errorf("place(%q).rel = %q, want %q", dir, got.rel, "docs/proj")
	}
	if got.key != wantKey {
		t.Errorf("place(%q).key = %q, want %q", dir, got.key, wantKey)
	}
	if !hexKey.MatchString(got.key) {
		t.Errorf("place(%q).key = %q, want 32 lower-case hex characters", dir, got.key)
	}
	if want := h + "/roots/home/dirs/" + wantKey; got.target != want {
		t.Errorf("place(%q).target = %q, want %q", dir, got.target, want)
	}
	if want := filepath.Join(dir, ".snapshot"); got.link != want {
		t.Errorf("place(%q).link = %q, want %q", dir, got.link, want)
	}
	if limit := 60 + len(h); len(got.target) >= limit {
		t.Errorf("len(place(%q).target) = %d, want < %d", dir, len(got.target), limit)
	}
}

func TestPlaceRootItself(t *testing.T) {
	pol, r, _ := newPlaceFixture(t)

	got, err := place(pol, r)
	if err != nil {
		t.Fatalf("place(%q) error = %v, want nil", r, err)
	}
	if got.rel != "" {
		t.Errorf("place(%q).rel = %q, want %q", r, got.rel, "")
	}
	if want := filepath.Join(r, ".snapshot"); got.link != want {
		t.Errorf("place(%q).link = %q, want %q", r, got.link, want)
	}
}

func TestPlaceRefusesOutsideRoots(t *testing.T) {
	pol, _, _ := newPlaceFixture(t)
	dir := t.TempDir()

	_, err := place(pol, dir)
	if !errors.Is(err, ErrOutsideRoots) {
		t.Errorf("place(%q) error = %v, want ErrOutsideRoots", dir, err)
	}
	if got := errcode.Of(err); got != errcode.InvalidConfig {
		t.Errorf("errcode.Of(place(%q)) = %q, want %q", dir, got, errcode.InvalidConfig)
	}
}

func TestPlaceRefusesExcludedPaths(t *testing.T) {
	pol, r, h := newPlaceFixture(t)
	inner := filepath.Join(h, "inner")
	withInner := pol
	withInner.Roots = append([]resolver.RootSpec{{ID: "inner", LocalPath: inner}}, pol.Roots...)

	tests := []struct {
		name string
		pol  Policy
		dir  string
	}{
		{name: "excluded entry", pol: pol, dir: filepath.Join(r, "state")},
		{name: "under excluded entry", pol: pol, dir: filepath.Join(r, "state", "cache", "x")},
		{name: "history mount", pol: pol, dir: h},
		{name: "under history mount", pol: pol, dir: filepath.Join(h, "roots", "home")},
		{name: "root under history mount", pol: withInner, dir: filepath.Join(inner, "x")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := place(tt.pol, tt.dir)
			if !errors.Is(err, ErrExcluded) {
				t.Errorf("place(%q) error = %v, want ErrExcluded", tt.dir, err)
			}
		})
	}

	t.Run("prefix is not a component", func(t *testing.T) {
		dir := filepath.Join(r, "stateful")
		if _, err := place(pol, dir); err != nil {
			t.Errorf("place(%q) error = %v, want nil", dir, err)
		}
	})
}

func TestPlaceRefusesLinkNameComponent(t *testing.T) {
	pol, r, _ := newPlaceFixture(t)

	tests := []struct {
		name string
		dir  string
	}{
		{name: "link name itself", dir: filepath.Join(r, ".snapshot")},
		{name: "link name in the middle", dir: filepath.Join(r, "a", ".snapshot", "b")},
		{name: "under link name", dir: filepath.Join(r, ".snapshot", "roots")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := place(pol, tt.dir)
			if !errors.Is(err, ErrExcluded) {
				t.Errorf("place(%q) error = %v, want ErrExcluded", tt.dir, err)
			}
		})
	}

	t.Run("link name prefix is allowed", func(t *testing.T) {
		dir := filepath.Join(r, "a", ".snapshotx")
		if _, err := place(pol, dir); err != nil {
			t.Errorf("place(%q) error = %v, want nil", dir, err)
		}
	})
}
