package links

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

const ensureOp = "links.ensure"

// Ensure creates the link for dir if absent, or proves an existing one is owned.
func (e *Engine) Ensure(ctx context.Context, dir string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	p, err := place(e.pol, dir)
	if err != nil {
		return Result{}, err
	}
	res := Result{Key: p.key, Path: p.link}
	unlock := e.lock(p.key)
	defer unlock()

	fd, pending, err := e.inspect(ctx, p)
	if err != nil || !pending {
		return res, err
	}
	defer func() { _ = unix.Close(fd) }()

	rec := pendingRecord(p, dir)
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	created, err := e.link(fd, p)
	if err != nil || !created {
		return res, err
	}
	rec.State = StateOwned
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	res.Created = true
	return res, nil
}

// inspect runs Ensure's pre-checks for p with its key locked. When a new link
// is needed it reports pending and returns the open directory fd, which the
// caller closes. Otherwise no fd is left open and err is nil only when the
// existing entry is an owned link.
func (e *Engine) inspect(ctx context.Context, p placement) (fd int, pending bool, err error) {
	fd, err = openDirChain(e.rootPath(p.rootID), p.rel)
	if err != nil {
		return -1, false, fsErr(err)
	}
	keep := false
	defer func() {
		if !keep {
			_ = unix.Close(fd)
		}
	}()

	name := e.pol.LinkName
	other, found, err := caseFoldSibling(fd, name)
	if err != nil {
		return -1, false, fsErr(err)
	}
	if found {
		return -1, false, errcode.New(errcode.LinkConflict, ensureOp, fmt.Errorf("%q has case-equivalent sibling %q", p.link, other))
	}

	if _, err := lstatAt(fd, name); err == nil {
		return -1, false, e.proveOwned(fd, p)
	} else if !errors.Is(err, unix.ENOENT) {
		return -1, false, fsErr(err)
	}

	if err := ctx.Err(); err != nil {
		return -1, false, err
	}
	keep = true
	return fd, true, nil
}

// pendingRecord returns the pending registry record for p's link in dir.
func pendingRecord(p placement, dir string) Record {
	return Record{
		Key:    p.key,
		RootID: p.rootID,
		Rel:    rawpath.Path(p.rel),
		Dir:    rawpath.Path(filepath.Clean(dir)),
		Target: p.target,
		State:  StatePending,
	}
}

// link creates p's symlink in fd once its pending record is stored. On
// failure it deletes that record; if an entry appeared meanwhile, it reports
// created false and whether that entry is an owned link.
func (e *Engine) link(fd int, p placement) (created bool, err error) {
	err = symlinkAt(fd, e.pol.LinkName, p.target)
	if err == nil {
		return true, nil
	}
	if derr := e.reg.Delete(p.key); derr != nil {
		return false, errors.Join(fsErr(err), derr)
	}
	if errors.Is(err, unix.EEXIST) {
		return false, e.proveOwned(fd, p)
	}
	return false, fsErr(err)
}

// proveOwned returns nil only when the existing entry is a symlink to the
// exact expected target and the registry holds an owned record for it.
func (e *Engine) proveOwned(fd int, p placement) error {
	conflict := errcode.New(errcode.LinkConflict, ensureOp, fmt.Errorf("%q is not an owned link", p.link))
	st, err := lstatAt(fd, e.pol.LinkName)
	if err != nil {
		return fsErr(err)
	}
	if st.Mode&unix.S_IFMT != unix.S_IFLNK {
		return conflict
	}
	target, err := readlinkAt(fd, e.pol.LinkName)
	if err != nil {
		return fsErr(err)
	}
	rec, ok, err := e.reg.Get(p.key)
	if err != nil {
		return err
	}
	if !ok || rec.State != StateOwned || rec.Target != p.target || target != p.target {
		return conflict
	}
	return nil
}

func (e *Engine) rootPath(id string) string {
	for _, r := range e.pol.Roots {
		if r.ID == id {
			return r.LocalPath
		}
	}
	return ""
}

// fsErr maps permission and read-only filesystem errors to permission_denied.
func fsErr(err error) error {
	if errors.Is(err, unix.EACCES) || errors.Is(err, unix.EPERM) || errors.Is(err, unix.EROFS) {
		return errcode.New(errcode.PermissionDenied, ensureOp, err)
	}
	return err
}
