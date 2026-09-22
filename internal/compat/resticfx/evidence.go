package resticfx

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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
	return getenv("SNAPBACK_EVIDENCE_DIR")
}

// WriteEvidence writes pathtemplate-<ev.GOOS>.json under dir.
func WriteEvidence(dir string, ev Evidence) (string, error) {
	if dir == "" {
		return "", errors.New("evidence dir is empty")
	}
	data, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode evidence: %w", err)
	}
	path := filepath.Join(dir, "pathtemplate-"+ev.GOOS+".json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write evidence: %w", err)
	}
	return path, nil
}

// ReadEvidence reads an evidence file.
func ReadEvidence(path string) (Evidence, error) {
	var ev Evidence
	data, err := os.ReadFile(path)
	if err != nil {
		return ev, fmt.Errorf("read evidence: %w", err)
	}
	if err := json.Unmarshal(data, &ev); err != nil {
		return ev, fmt.Errorf("decode evidence: %w", err)
	}
	return ev, nil
}
