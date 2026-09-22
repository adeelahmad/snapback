package setup

import (
	"context"
	"fmt"
)

// LinkRoots ensures the .snapshot link of every root, in order, through the
// injected link function, and returns the paths it linked — created ones and
// ones that already existed — so the caller can report them. It stops at the
// first root that cannot be linked, naming that root in the error.
func LinkRoots(ctx context.Context, link func(ctx context.Context, dir string) (created bool, err error), roots []string) ([]string, error) {
	linked := make([]string, 0, len(roots))
	for _, root := range roots {
		if _, err := link(ctx, root); err != nil {
			return nil, fmt.Errorf("link %s: %w", root, err)
		}
		linked = append(linked, root)
	}
	return linked, nil
}
