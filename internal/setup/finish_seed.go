package setup

import "context"

// LinkRoots is filled in by the GREEN step.
func LinkRoots(ctx context.Context, link func(ctx context.Context, dir string) (created bool, err error), roots []string) ([]string, error) {
	return nil, nil
}
