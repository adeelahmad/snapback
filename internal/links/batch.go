package links

import (
	"context"
	"errors"
	"sort"

	"golang.org/x/sys/unix"
)

// EnsureError reports the failure of one directory in an EnsureBatch call.
type EnsureError struct {
	Dir string
	Err error
}

func (e *EnsureError) Error() string { return e.Dir + ": " + e.Err.Error() }

func (e *EnsureError) Unwrap() error { return e.Err }

// EnsureBatch ensures the link for every dir with Ensure's per-directory
// guarantees, but stores the pending records of the whole batch in one write
// transaction and the owned records in another. Results are in input order,
// and the error joins one *EnsureError per failed directory.
func (e *Engine) EnsureBatch(ctx context.Context, dirs []string) ([]Result, error) {
	res := make([]Result, len(dirs))
	errs := make([]error, len(dirs))
	// A key repeated in the batch would wait on its own lock, so repeats are
	// ensured one by one after the batch releases its locks.
	for _, i := range e.ensureBatch(ctx, dirs, res, errs) {
		res[i], errs[i] = e.Ensure(ctx, dirs[i])
	}

	var joined []error
	for i, err := range errs {
		if err != nil {
			joined = append(joined, &EnsureError{Dir: dirs[i], Err: err})
		}
	}
	return res, errors.Join(joined...)
}

// pendingLink is a batch directory whose link is about to be created.
type pendingLink struct {
	i   int
	fd  int
	p   placement
	rec Record
}

// ensureBatch fills res and errs for the first directory of every key and
// returns the indexes of repeated keys, which it leaves untouched.
func (e *Engine) ensureBatch(ctx context.Context, dirs []string, res []Result, errs []error) (repeats []int) {
	places := make([]placement, len(dirs))
	seen := map[string]bool{}
	var keys []string
	var todo []int
	for i, dir := range dirs {
		if err := ctx.Err(); err != nil {
			errs[i] = err
			continue
		}
		p, err := place(e.pol, dir)
		if err != nil {
			errs[i] = err
			continue
		}
		if seen[p.key] {
			repeats = append(repeats, i)
			continue
		}
		seen[p.key] = true
		places[i] = p
		res[i] = Result{Key: p.key, Path: p.link}
		keys = append(keys, p.key)
		todo = append(todo, i)
	}
	// Locking in key order keeps concurrent batches from deadlocking.
	sort.Strings(keys)
	unlocks := make([]func(), len(keys))
	for n, k := range keys {
		unlocks[n] = e.lock(k)
	}
	defer func() {
		for _, unlock := range unlocks {
			unlock()
		}
	}()

	var pending []pendingLink
	defer func() {
		for _, pl := range pending {
			_ = unix.Close(pl.fd)
		}
	}()
	for _, i := range todo {
		fd, ok, err := e.inspect(ctx, places[i])
		if err != nil {
			errs[i] = err
			continue
		}
		if ok {
			pending = append(pending, pendingLink{i: i, fd: fd, p: places[i], rec: pendingRecord(places[i], dirs[i])})
		}
	}
	if len(pending) == 0 {
		return repeats
	}

	recs := make([]Record, len(pending))
	for n, pl := range pending {
		recs[n] = pl.rec
	}
	if err := e.reg.PutAll(recs); err != nil {
		for _, pl := range pending {
			errs[pl.i] = err
		}
		return repeats
	}

	var owned []Record
	var created []int
	for _, pl := range pending {
		ok, err := e.link(pl.fd, pl.p)
		if err != nil {
			errs[pl.i] = err
			continue
		}
		if ok {
			rec := pl.rec
			rec.State = StateOwned
			owned = append(owned, rec)
			created = append(created, pl.i)
		}
	}
	if len(owned) == 0 {
		return repeats
	}
	err := e.reg.PutAll(owned)
	for _, i := range created {
		if err != nil {
			errs[i] = err
			continue
		}
		res[i].Created = true
	}
	return repeats
}
