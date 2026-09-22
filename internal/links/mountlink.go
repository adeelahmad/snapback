package links

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

const mountOp = "links.EnsureMountLink"

// mountRemoveOp names the operation that withdraws a mount link.
const mountRemoveOp = "links.RemoveMountLink"

// mountRootID marks registry records for repository mount points, which sit
// outside every configured root.
const mountRootID = "mount"

// mountKeyPrefix keeps mount keys stable and distinct from placement keys.
const mountKeyPrefix = "mount:"

// mountTmpSuffix names the transient symlink a repoint renames into place.
const mountTmpSuffix = ".snapback-new"

// mountAction is the work an existing entry at the link path requires.
type mountAction uint8

const (
	mountKeep mountAction = iota
	mountCreate
	mountRepoint
)

// EnsureMountLink creates dir with mode and publishes the managed link inside
// it pointing at target, repointing a stale managed link atomically. It never
// removes a regular file or directory found under the link name.
func (e *Engine) EnsureMountLink(dir, target string, mode fs.FileMode) (Result, error) {
	clean := filepath.Clean(dir)
	name := e.pol.LinkName
	res := Result{Key: mountKeyPrefix + clean, Path: filepath.Join(clean, name)}
	unlock := e.lock(res.Key)
	defer unlock()

	// MkdirAll applies the umask, so the mode is forced afterwards.
	if err := os.MkdirAll(clean, mode); err != nil {
		return res, fsErr(err)
	}
	if err := os.Chmod(clean, mode); err != nil {
		return res, fsErr(err)
	}

	fd, err := openDirChain(clean, "")
	if err != nil {
		return res, fsErr(err)
	}
	defer func() { _ = unix.Close(fd) }()

	act, err := mountActionFor(fd, name, target, res.Path)
	if err != nil {
		return res, err
	}
	rec := Record{
		Key:    res.Key,
		RootID: mountRootID,
		Rel:    rawpath.Path(clean),
		Dir:    rawpath.Path(clean),
		Target: target,
		State:  StateOwned,
	}
	if act == mountKeep {
		return res, e.reg.Put(rec)
	}

	rec.State = StatePending
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	if err := writeMountLink(fd, name, target, act); err != nil {
		if derr := e.reg.Delete(res.Key); derr != nil {
			return res, errors.Join(err, derr)
		}
		return res, err
	}
	rec.State = StateOwned
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	res.Created = act == mountCreate
	res.Repointed = act == mountRepoint
	return res, nil
}

// mountActionFor reports what the entry named name in fd requires, and fails
// with a link conflict naming link when it is not a symlink.
func mountActionFor(fd int, name, target, link string) (mountAction, error) {
	st, err := lstatAt(fd, name)
	if errors.Is(err, unix.ENOENT) {
		return mountCreate, nil
	}
	if err != nil {
		return mountKeep, fsErr(err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFLNK {
		return mountKeep, errcode.New(errcode.LinkConflict, mountOp, fmt.Errorf("%q exists and is not a managed link", link))
	}
	current, err := readlinkAt(fd, name)
	if err != nil {
		return mountKeep, fsErr(err)
	}
	if current == target {
		return mountKeep, nil
	}
	return mountRepoint, nil
}

// writeMountLink creates the symlink, or replaces a stale one by renaming a
// freshly made sibling over it so the name is never absent.
func writeMountLink(fd int, name, target string, act mountAction) error {
	if act == mountCreate {
		return fsErr(symlinkAt(fd, name, target))
	}
	tmp := name + mountTmpSuffix
	if err := unlinkAt(fd, tmp); err != nil && !errors.Is(err, unix.ENOENT) {
		return fsErr(err)
	}
	if err := symlinkAt(fd, tmp, target); err != nil {
		return fsErr(err)
	}
	if err := unix.Renameat(fd, tmp, fd, name); err != nil {
		_ = unlinkAt(fd, tmp)
		return fsErr(err)
	}
	return nil
}

// RemoveMountLink withdraws the managed link published in dir by
// EnsureMountLink and deletes its registry record, leaving dir in place.
func (e *Engine) RemoveMountLink(dir string) error {
	clean := filepath.Clean(dir)
	key := mountKeyPrefix + clean
	unlock := e.lock(key)
	defer unlock()

	rec, ok, err := e.reg.Get(key)
	if err != nil || !ok {
		return err
	}

	// Mount records sit outside every configured root, so the mount point
	// is opened directly rather than through rootPath.
	fd, err := openDirChain(clean, "")
	if err != nil {
		return fsErr(err)
	}
	defer func() { _ = unix.Close(fd) }()

	name := e.pol.LinkName
	owned, err := linksTo(fd, name, rec.Target)
	if errors.Is(err, unix.ENOENT) {
		return e.reg.Delete(key)
	}
	if err != nil {
		return fsErr(err)
	}
	if !owned {
		link := filepath.Join(clean, name)
		return errcode.New(errcode.LinkConflict, mountRemoveOp, fmt.Errorf("%q is not a managed link", link))
	}
	if err := unlinkAt(fd, name); err != nil {
		return fsErr(err)
	}
	return e.reg.Delete(key)
}
