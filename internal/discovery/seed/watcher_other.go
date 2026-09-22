//go:build !linux

package seed

import (
	"context"
	"errors"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Run reports that directory watching needs Linux inotify.
func (w *Watcher) Run(ctx context.Context) error {
	return errcode.New(errcode.PrereqMissing, "seed.Watcher.Run", errors.New("directory watching needs inotify, which is only available on linux"))
}
