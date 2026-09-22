package resticfx

import "time"

// Evidence is the path-template verification record written for CI.
type Evidence struct {
	GOOS          string    `json:"goos"`
	GOARCH        string    `json:"goarch"`
	ResticVersion string    `json:"restic_version"`
	RcloneVersion string    `json:"rclone_version,omitempty"`
	PathTemplate  string    `json:"path_template"`
	SnapshotID    string    `json:"snapshot_id"`
	ObservedDir   string    `json:"observed_dir"`
	IDsEntries    []string  `json:"ids_entries"`
	Result        string    `json:"result"`
	Reason        string    `json:"reason,omitempty"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// EvidenceDir returns the evidence directory from SNAPBACK_EVIDENCE_DIR.
func EvidenceDir(getenv func(string) string) string {
	panic("SUB-AGENT-TODO: return getenv of SNAPBACK_EVIDENCE_DIR")
}

// WriteEvidence writes pathtemplate-<ev.GOOS>.json under dir.
func WriteEvidence(dir string, ev Evidence) (string, error) {
	panic("SUB-AGENT-TODO: pretty-printed JSON + trailing newline, 0644, to dir/pathtemplate-<ev.GOOS>.json; return path")
}

// ReadEvidence reads an evidence file.
func ReadEvidence(path string) (Evidence, error) {
	panic("SUB-AGENT-TODO: read file and json-decode into Evidence")
}
