package seed

import (
	"context"

	"golang.org/x/sys/unix"
)

// addWatch is the inotify_add_watch seam; tests replace it to inject failures.
var addWatch = unix.InotifyAddWatch

// Run watches every root with inotify until ctx is done.
func (w *Watcher) Run(ctx context.Context) error {
	panic("SUB-AGENT-TODO: InotifyInit1(IN_CLOEXEC|IN_NONBLOCK); addWatch each non-excluded dir under every root (IN_CREATE|IN_MOVED_TO|IN_DELETE_SELF|IN_ONLYDIR); on IN_ISDIR create/move enqueue the dir and watch it recursively; addWatch failure marks Degraded and continues; poll the fd and return nil when ctx is done, closing the fd")
}
