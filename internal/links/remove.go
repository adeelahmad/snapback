package links

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const removeOp = "links.remove"

// Remove unlinks the .snapshot entry in dir when the registry proves Snapback
// owns it, then deletes the record.
func (e *Engine) Remove(ctx context.Context, dir string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p, err := place(e.pol, dir)
	if err != nil {
		return err
	}
	unlock := e.lock(p.key)
	defer unlock()
	rec, ok, err := e.reg.Get(p.key)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.New(errcode.LinkConflict, removeOp, fmt.Errorf("%q has no registry record", p.link))
	}
	return e.removeOwned(rec)
}

// removeOwned unlinks rec's link only when it is a symlink to rec's exact
// target, then deletes the record. Any other entry is left untouched.
func (e *Engine) removeOwned(rec Record) error {
	fd, err := openDirChain(e.rootPath(rec.RootID), string(rec.Rel))
	if err != nil {
		return fsErr(err)
	}
	defer func() { _ = unix.Close(fd) }()

	name := e.pol.LinkName
	owned, err := linksTo(fd, name, rec.Target)
	if errors.Is(err, unix.ENOENT) {
		return e.reg.Delete(rec.Key)
	}
	if err != nil {
		return fsErr(err)
	}
	if !owned {
		link := filepath.Join(string(rec.Dir), name)
		return errcode.New(errcode.LinkConflict, removeOp, fmt.Errorf("%q is not an owned link", link))
	}
	if err := unlinkAt(fd, name); err != nil {
		return fsErr(err)
	}
	return e.reg.Delete(rec.Key)
}

// linksTo reports whether name in fd is a symlink whose target is exactly
// target.
func linksTo(fd int, name, target string) (bool, error) {
	st, err := lstatAt(fd, name)
	if err != nil {
		return false, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFLNK {
		return false, nil
	}
	got, err := readlinkAt(fd, name)
	if err != nil {
		return false, err
	}
	return got == target, nil
}

// RemoveManaged applies the Remove ownership rule to every record.
func (e *Engine) RemoveManaged(ctx context.Context) (RepairReport, error) {
	var rep RepairReport
	recs, err := e.reg.List()
	if err != nil {
		return rep, err
	}
	for _, rec := range recs {
		if err := ctx.Err(); err != nil {
			return rep, err
		}
		unlock := e.lock(rec.Key)
		err := e.removeOwned(rec)
		unlock()
		ent := RepairEntry{Key: rec.Key, Dir: rec.Dir}
		switch {
		case err == nil:
			rep.Removed = append(rep.Removed, ent)
		case errcode.Of(err) == errcode.LinkConflict:
			ent.Code = errcode.LinkConflict
			rep.Preserved = append(rep.Preserved, ent)
		default:
			return rep, err
		}
	}
	return rep, nil
}

// List returns every registry record, sorted by key.
func (e *Engine) List() ([]Record, error) {
	return e.reg.List()
}
