package setup

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestLinkRootsLinksEachRootOnce(t *testing.T) {
	roots := []string{"/a", "/b", "/c"}
	var calls []string
	link := func(_ context.Context, dir string) (bool, error) {
		calls = append(calls, dir)
		return dir != "/b", nil
	}

	got, err := LinkRoots(context.Background(), link, roots)
	if err != nil {
		t.Fatalf("LinkRoots(ctx, link, %v) error = %v, want nil", roots, err)
	}
	if !slices.Equal(calls, roots) {
		t.Errorf("LinkRoots(ctx, link, %v) called link with %v, want %v", roots, calls, roots)
	}
	if !slices.Equal(got, roots) {
		t.Errorf("LinkRoots(ctx, link, %v) = %v, want %v", roots, got, roots)
	}
}

func TestLinkRootsStopsAtFirstError(t *testing.T) {
	roots := []string{"/a", "/b", "/c"}
	boom := errors.New("boom")
	var calls []string
	link := func(_ context.Context, dir string) (bool, error) {
		calls = append(calls, dir)
		if dir == "/b" {
			return false, boom
		}
		return true, nil
	}

	got, err := LinkRoots(context.Background(), link, roots)
	if !errors.Is(err, boom) {
		t.Fatalf("LinkRoots(ctx, link, %v) error = %v, want %v", roots, err, boom)
	}
	if want := "/b"; !strings.Contains(err.Error(), want) {
		t.Errorf("LinkRoots(ctx, link, %v) error = %q, want it to name %q", roots, err, want)
	}
	if want := []string{"/a", "/b"}; !slices.Equal(calls, want) {
		t.Errorf("LinkRoots(ctx, link, %v) called link with %v, want %v", roots, calls, want)
	}
	if got != nil {
		t.Errorf("LinkRoots(ctx, link, %v) = %v, want nil", roots, got)
	}
}
