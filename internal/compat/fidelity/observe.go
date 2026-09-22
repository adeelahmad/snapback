package fidelity

import (
	"io/fs"
	"os"
	"time"
)

// Observed is the metadata read from one path, plus times that are recorded
// but never asserted.
type Observed struct {
	Meta
	CTime     time.Time
	BirthTime *time.Time
}

// Observe reads the metadata of path without following its final component.
func Observe(path string) (Observed, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return Observed{}, err
	}
	o := Observed{Meta: Meta{Path: path, Size: fi.Size(), Mode: fi.Mode(), MTime: fi.ModTime()}}
	if fi.Mode()&fs.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return Observed{}, err
		}
		o.LinkTarget = target
	}
	o.CTime, o.BirthTime = statTimes(fi)
	return o, nil
}
