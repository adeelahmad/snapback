//go:build darwin

package fidelity

import (
	"io/fs"
	"syscall"
	"time"
)

// statTimes returns the change time and birth time recorded for fi.
func statTimes(fi fs.FileInfo) (time.Time, *time.Time) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, nil
	}
	birth := time.Unix(st.Birthtimespec.Unix())
	return time.Unix(st.Ctimespec.Unix()), &birth
}
