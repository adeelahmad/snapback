//go:build linux

package fidelity

import (
	"io/fs"
	"syscall"
	"time"
)

// statTimes returns the change time recorded for fi. Birth time is not
// exposed on linux without statx, which is outside the stdlib.
func statTimes(fi fs.FileInfo) (time.Time, *time.Time) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, nil
	}
	return time.Unix(st.Ctim.Unix()), nil
}
