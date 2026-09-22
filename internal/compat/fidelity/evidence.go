package fidelity

import (
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

// MarshalJSON encodes u as {value, claimed:false, note}.
func (u Unclaimed) MarshalJSON() ([]byte, error) {
	panic("SUB-AGENT-TODO: T4 encode {value: RFC3339Nano time or null when Value is nil, claimed: false, note: recorded-not-asserted, or not-exposed when Value is nil}")
}

// UnmarshalJSON decodes the {value, claimed, note} form into u.
func (u *Unclaimed) UnmarshalJSON(data []byte) error {
	panic("SUB-AGENT-TODO: T4 decode the object; Value is the parsed time, nil when value is null")
}

// WriteEvidence writes ev as indented JSON to dir/fidelity-<GOOS>.json and
// returns the path.
func WriteEvidence(dir string, ev Evidence) (string, error) {
	panic("SUB-AGENT-TODO: T4 error if dir does not exist; json.MarshalIndent; write filepath.Join(dir, fidelity-<runtime.GOOS>.json) with mode 0644; return the path")
}

// ReadEvidence reads the evidence JSON at path.
func ReadEvidence(path string) (Evidence, error) {
	panic("SUB-AGENT-TODO: T4 os.ReadFile then json.Unmarshal into Evidence; wrap errors with the path")
}
