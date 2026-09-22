// agentic:shim

package seed

import (
	"context"
	"errors"
)

// addWatch is the inotify_add_watch seam; the shim body always fails.
var addWatch = func(int, string, uint32) (int, error) {
	return -1, errors.New("agentic shim: addWatch not implemented")
}

// Run is a compile shim that returns at once without watching anything.
func (w *Watcher) Run(context.Context) error {
	_, _ = addWatch(-1, "", 0)
	return nil
}
