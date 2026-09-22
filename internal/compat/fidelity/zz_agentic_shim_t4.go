// agentic:shim

package fidelity

import (
	"errors"
	"io/fs"
	"time"
)

// Evidence is the per-platform fidelity record written as JSON.
type Evidence struct {
	Platform                  string         `json:"platform"`
	ResticVersion             string         `json:"restic_version"`
	SnapshotID                string         `json:"snapshot_id"`
	SnapshotAlias             string         `json:"snapshot_alias"`
	ResolvedInsideResticMount bool           `json:"resolved_inside_restic_mount"`
	MTimeToleranceNs          int64          `json:"mtime_tolerance_ns"`
	MTimePrecisionNs          int64          `json:"mtime_precision_ns"`
	FilesGenerated            int            `json:"files_generated"`
	FilesCompared             int            `json:"files_compared"`
	Pass                      bool           `json:"pass"`
	Files                     []FileEvidence `json:"files"`
}

// FileEvidence is the evidence for one compared file.
type FileEvidence struct {
	Path         string    `json:"path"`
	Expected     Attrs     `json:"expected"`
	Observed     Attrs     `json:"observed"`
	ResticLs     Attrs     `json:"restic_ls"`
	SizeOK       bool      `json:"size_ok"`
	ModeOK       bool      `json:"mode_ok"`
	MTimeOK      bool      `json:"mtime_ok"`
	MTimeDeltaNs int64     `json:"mtime_delta_ns"`
	CTime        Unclaimed `json:"ctime"`
	BirthTime    Unclaimed `json:"birth_time"`
}

// Attrs are the asserted attributes of one file.
type Attrs struct {
	Size  int64       `json:"size"`
	Mode  fs.FileMode `json:"mode"`
	MTime time.Time   `json:"mtime"`
}

// Unclaimed is a recorded timestamp that is never asserted. It encodes as
// {value, claimed:false, note}; a nil Value encodes as null.
type Unclaimed struct {
	Value *time.Time
}

// MarshalJSON is a compile shim with a deliberately wrong body.
func (u Unclaimed) MarshalJSON() ([]byte, error) {
	return []byte(`{"value":"shim","claimed":true,"note":""}`), nil
}

// UnmarshalJSON is a compile shim with a deliberately wrong body.
func (u *Unclaimed) UnmarshalJSON(_ []byte) error {
	u.Value = nil
	return nil
}

// WriteEvidence is a compile shim with a deliberately wrong body.
func WriteEvidence(_ string, _ Evidence) (string, error) {
	return "", errors.New("shim: not implemented")
}

// ReadEvidence is a compile shim with a deliberately wrong body.
func ReadEvidence(_ string) (Evidence, error) {
	return Evidence{}, nil
}

// MissingPrereq is a compile shim with a deliberately wrong body.
func MissingPrereq(_ func(string) string, _ func(string) (string, error), _ func(string) bool, _ string) string {
	return "shim"
}
