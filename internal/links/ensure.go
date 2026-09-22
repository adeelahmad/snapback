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

	fd, err := openDirChain(e.rootPath(p.rootID), p.rel)
	if err != nil {
		return res, fsErr(err)
	}
	defer func() { _ = unix.Close(fd) }()

	name := e.pol.LinkName
	other, found, err := caseFoldSibling(fd, name)
	if err != nil {
		return res, fsErr(err)
	}
	if found {
		return res, errcode.New(errcode.LinkConflict, ensureOp, fmt.Errorf("%q has case-equivalent sibling %q", p.link, other))
	}

	if _, err := lstatAt(fd, name); err == nil {
		return res, e.proveOwned(fd, p)
	} else if !errors.Is(err, unix.ENOENT) {
		return res, fsErr(err)
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}
	rec := Record{
		Key:    p.key,
		RootID: p.rootID,
		Rel:    rawpath.Path(p.rel),
		Dir:    rawpath.Path(filepath.Clean(dir)),
		Target: p.target,
		State:  StatePending,
	}
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	if err := symlinkAt(fd, name, p.target); err != nil {
		if derr := e.reg.Delete(p.key); derr != nil {
			return res, errors.Join(fsErr(err), derr)
		}
		if errors.Is(err, unix.EEXIST) {
			return res, e.proveOwned(fd, p)
		}
		return res, fsErr(err)
	}
	rec.State = StateOwned
	if err := e.reg.Put(rec); err != nil {
		return res, err
	}
	res.Created = true
	return res, nil
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
