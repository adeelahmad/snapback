//go:build !linux

package seed

import "context"

// Run reports that directory watching needs Linux inotify.
func (w *Watcher) Run(ctx context.Context) error {
	panic("SUB-AGENT-TODO: return errcode.New(errcode.PrereqMissing, ...) naming inotify/Linux as the missing prerequisite")
}
