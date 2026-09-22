package links

import (
	"context"
	"errors"
	"path/filepath"

	"golang.org/x/sys/unix"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/rawpath"
)

// RepairEntry is one directory a Repair or RemoveManaged pass acted on.
type RepairEntry struct {
	Key  string
	Dir  rawpath.Path
	Code errcode.Code
}

// RepairReport lists what a Repair or RemoveManaged pass did.
type RepairReport struct {
	Completed, Repaired, Removed, Preserved []RepairEntry
}

// Repair finishes pending records and rewrites owned links whose expected
// target changed.
func (e *Engine) Repair(ctx context.Context) (RepairReport, error) {
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
		err := e.repairOne(rec, &rep)
		unlock()
		if err != nil {
			return rep, err
		}
	}
	return rep, nil
}

// repairOne resolves one record and appends the outcome to rep.
func (e *Engine) repairOne(rec Record, rep *RepairReport) error {
	fd, err := openDirChain(e.rootPath(rec.RootID), string(rec.Rel))
	if err != nil {
		return fsErr(err)
	}
	defer func() { _ = unix.Close(fd) }()

	name := e.pol.LinkName
	ent := RepairEntry{Key: rec.Key, Dir: rec.Dir}
	owned, err := linksTo(fd, name, rec.Target)
	absent := errors.Is(err, unix.ENOENT)
	if err != nil && !absent {
		return fsErr(err)
	}

	if rec.State == StatePending {
		switch {
		case absent:
			if err := symlinkAt(fd, name, rec.Target); err != nil {
				return fsErr(err)
			}
		case !owned:
			ent.Code = errcode.LinkConflict
			rep.Preserved = append(rep.Preserved, ent)
			return e.reg.Delete(rec.Key)
		}
		rec.State = StateOwned
		rep.Completed = append(rep.Completed, ent)
		return e.reg.Put(rec)
	}

	want := filepath.Join(e.pol.HistoryMount, "roots", rec.RootID, "dirs", rec.Key)
	if absent || rec.Target == want {
		return nil
	}
	if !owned {
		ent.Code = errcode.LinkConflict
		rep.Preserved = append(rep.Preserved, ent)
		return nil
	}
	if err := unlinkAt(fd, name); err != nil {
		return fsErr(err)
	}
	if err := symlinkAt(fd, name, want); err != nil {
		return fsErr(err)
	}
	rec.Target = want
	rep.Repaired = append(rep.Repaired, ent)
	return e.reg.Put(rec)
}
