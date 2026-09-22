//go:build darwin

package fidelity

import (
	"io/fs"
	"time"
)

// statTimes returns the change time and birth time recorded for fi.
func statTimes(fi fs.FileInfo) (time.Time, *time.Time) {
	panic("SUB-AGENT-TODO: T3 darwin: fi.Sys().(*syscall.Stat_t); CTime from Ctimespec, BirthTime from Birthtimespec (both set)")
}
