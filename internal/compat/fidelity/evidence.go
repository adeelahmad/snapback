package fidelity

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	noteRecorded     = "FUSE-approximated; recorded, not claimed"
	noteNotExposed   = "not exposed; recorded, not claimed"
	evidenceFileMode = 0o644
)

type unclaimedJSON struct {
	Value   *time.Time `json:"value"`
	Claimed bool       `json:"claimed"`
	Note    string     `json:"note"`
}

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
	note := noteRecorded
	if u.Value == nil {
		note = noteNotExposed
	}
	return json.Marshal(unclaimedJSON{Value: u.Value, Note: note})
}

// UnmarshalJSON decodes the {value, claimed, note} form into u.
func (u *Unclaimed) UnmarshalJSON(data []byte) error {
	var v unclaimedJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	u.Value = v.Value
	return nil
}

// WriteEvidence writes ev as indented JSON to dir/fidelity-<GOOS>.json and
// returns the path.
func WriteEvidence(dir string, ev Evidence) (string, error) {
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("evidence dir: %w", err)
	}
	data, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode evidence: %w", err)
	}
	path := filepath.Join(dir, "fidelity-"+runtime.GOOS+".json")
	if err := os.WriteFile(path, data, evidenceFileMode); err != nil {
		return "", fmt.Errorf("write evidence %s: %w", path, err)
	}
	return path, nil
}

// ReadEvidence reads the evidence JSON at path.
func ReadEvidence(path string) (Evidence, error) {
	var ev Evidence
	data, err := os.ReadFile(path)
	if err != nil {
		return ev, fmt.Errorf("read evidence %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &ev); err != nil {
		return ev, fmt.Errorf("decode evidence %s: %w", path, err)
	}
	return ev, nil
}
