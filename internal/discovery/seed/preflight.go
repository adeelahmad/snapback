package seed

import (
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// Statfs is the inode view of a filesystem that Preflight budgets against.
type Statfs struct {
	Files, FreeFiles uint64
}

// StatfsOf reports the inode totals of the filesystem holding path.
func StatfsOf(path string) (Statfs, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return Statfs{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	return Statfs{Files: st.Files, FreeFiles: st.Ffree}, nil
}

// Preflight refuses a plan whose link count would exceed the inode budget or
// maxLinks, unless force is set.
func Preflight(p Plan, fsStat func(path string) (Statfs, error), threshold float64, maxLinks int, force bool) error {
	const op = "seed preflight"
	if force || len(p.Dirs) == 0 || p.Count == 0 {
		return nil
	}
	if p.Count > maxLinks {
		return errcode.New(errcode.InodeBudgetExceeded, op, fmt.Errorf("%d links exceed max_links_per_path %d", p.Count, maxLinks))
	}
	st, err := fsStat(p.Dirs[0])
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	used := float64(st.Files-st.FreeFiles+uint64(p.Count)) / float64(st.Files)
	if used > threshold {
		return errcode.New(errcode.InodeBudgetExceeded, op, fmt.Errorf("%d links would bring inode use to %.0f%%, above the %.0f%% threshold", p.Count, used*100, threshold*100))
	}
	return nil
}
