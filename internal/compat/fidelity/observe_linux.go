//go:build linux

package fidelity

import (
	"io/fs"
	"time"
)

// statTimes returns the change time recorded for fi. Birth time is not
// exposed on linux without statx, which is outside the stdlib.
func statTimes(fi fs.FileInfo) (time.Time, *time.Time) {
	panic("SUB-AGENT-TODO: T3 linux: fi.Sys().(*syscall.Stat_t); CTime from Ctim; BirthTime nil (not exposed)")
}
