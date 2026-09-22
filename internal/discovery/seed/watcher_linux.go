package seed

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// addWatch is the inotify_add_watch seam; tests replace it to inject failures.
var addWatch = unix.InotifyAddWatch

const (
	watchMask = unix.IN_CREATE | unix.IN_MOVED_TO | unix.IN_DELETE_SELF | unix.IN_ONLYDIR
	// pollTimeoutMs bounds how long Run waits before rechecking ctx.
	pollTimeoutMs = 100
	eventBufSize  = 64 * 1024
)

// inotifyLoop owns the inotify fd and the directory bookkeeping; only the Run
// goroutine uses it.
type inotifyLoop struct {
	w     *Watcher
	fd    int
	paths map[int32]string
	known map[string]bool
}

// Run watches every root with inotify until ctx is done.
func (w *Watcher) Run(ctx context.Context) error {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return errcode.New(errcode.PrereqMissing, "seed.Watcher.Run", fmt.Errorf("inotify_init1: %w", err))
	}
	defer func() { _ = unix.Close(fd) }()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.drain(ctx)
	}()
	defer wg.Wait()

	lp := &inotifyLoop{w: w, fd: fd, paths: map[int32]string{}, known: map[string]bool{}}
	lp.scanRoots()

	buf := make([]byte, eventBufSize)
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for ctx.Err() == nil {
		n, err := unix.Poll(fds, pollTimeoutMs)
		if err != nil && !errors.Is(err, unix.EINTR) {
			return fmt.Errorf("seed.Watcher.Run: poll inotify: %w", err)
		}
		if n <= 0 {
			continue
		}
		n, err = unix.Read(fd, buf)
		if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return fmt.Errorf("seed.Watcher.Run: read inotify: %w", err)
		}
		lp.handle(buf[:n])
	}
	return nil
}

// scanRoots watches every root and the directories below it. Known
// directories are skipped; roots themselves are never enqueued.
func (lp *inotifyLoop) scanRoots() {
	for _, r := range lp.w.roots {
		if !lp.known[r.Root] {
			lp.known[r.Root] = true
			lp.watch(r.Root)
		}
		lp.walk(r.Root)
	}
}

// add records dir as new: it is watched, enqueued for linking, and its
// existing subdirectories are added too.
func (lp *inotifyLoop) add(dir string) {
	if lp.known[dir] || !lp.w.covered(dir) {
		return
	}
	lp.known[dir] = true
	lp.watch(dir)
	lp.w.enqueue([]string{dir})
	lp.walk(dir)
}

func (lp *inotifyLoop) walk(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			lp.add(filepath.Join(dir, e.Name()))
		}
	}
}

func (lp *inotifyLoop) watch(dir string) {
	wd, err := addWatch(lp.fd, dir, watchMask)
	switch {
	case err == nil:
		lp.paths[int32(wd)] = dir
	case errors.Is(err, unix.ENOSPC):
		lp.w.markDegraded("inotify watch limit reached; raise fs.inotify.max_user_watches")
	case errors.Is(err, unix.ENOENT), errors.Is(err, unix.ENOTDIR):
		// The directory vanished before it could be watched.
	default:
		lp.w.markDegraded(fmt.Sprintf("inotify_add_watch %q: %v", dir, err))
	}
}

// handle processes a buffer of raw inotify events.
func (lp *inotifyLoop) handle(buf []byte) {
	for off := 0; off+unix.SizeofInotifyEvent <= len(buf); {
		wd := int32(binary.NativeEndian.Uint32(buf[off:]))
		mask := binary.NativeEndian.Uint32(buf[off+4:])
		nameLen := int(binary.NativeEndian.Uint32(buf[off+12:]))
		nameStart := off + unix.SizeofInotifyEvent
		off = nameStart + nameLen
		if off > len(buf) {
			return
		}
		switch {
		case mask&unix.IN_Q_OVERFLOW != 0:
			lp.scanRoots()
		case mask&unix.IN_IGNORED != 0:
			if dir, ok := lp.paths[wd]; ok {
				delete(lp.paths, wd)
				delete(lp.known, dir)
			}
		case mask&unix.IN_ISDIR != 0 && mask&(unix.IN_CREATE|unix.IN_MOVED_TO) != 0:
			parent, ok := lp.paths[wd]
			if !ok {
				continue
			}
			lp.add(filepath.Join(parent, unix.ByteSliceToString(buf[nameStart:off])))
		}
	}
}

func (w *Watcher) markDegraded(reason string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.degraded, w.reason = true, reason
}
