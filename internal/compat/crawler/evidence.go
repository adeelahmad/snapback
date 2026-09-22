package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Row statuses. A skipped tool is never folded into a zero-hit pass.
const (
	StatusTested        Status = "tested"
	StatusNotTestedHere Status = "not-tested-here"
	StatusTimedOut      Status = "timed-out"
)

// SchemaVersion is the evidence schema version written into every Report.
const SchemaVersion = 1

// Row is the evidence for one crawler tool.
type Row struct {
	Tool     string         `json:"tool"`
	Argv     []string       `json:"argv"`
	Follows  bool           `json:"follows"`
	Status   Status         `json:"status"`
	Hits     *int           `json:"hits,omitempty"`
	Ops      map[string]int `json:"ops"`
	Version  string         `json:"version"`
	ExitCode int            `json:"exit_code"`
	Duration time.Duration  `json:"duration"`
	Reason   string         `json:"reason"`
}

// Report is the crawler evidence for one platform.
type Report struct {
	SchemaVersion       int    `json:"schema_version"`
	GOOS                string `json:"goos"`
	GOARCH              string `json:"goarch"`
	Seed                Shape  `json:"seed"`
	PositiveControlHits int    `json:"positive_control_hits"`
	Rows                []Row  `json:"rows"`
}

// Validate reports the first row that breaks the evidence rules.
func (r Report) Validate() error {
	for i, row := range r.Rows {
		switch row.Status {
		case StatusNotTestedHere:
			if row.Hits != nil {
				return fmt.Errorf("crawler: row %d (%s): %s row carries a hit count", i, row.Tool, row.Status)
			}
			if row.Reason == "" {
				return fmt.Errorf("crawler: row %d (%s): %s row lacks a reason", i, row.Tool, row.Status)
			}
		case StatusTested:
			if row.Hits == nil {
				return fmt.Errorf("crawler: row %d (%s): %s row has no hit count", i, row.Tool, row.Status)
			}
		case StatusTimedOut:
		default:
			return fmt.Errorf("crawler: row %d (%s): unknown status %q", i, row.Tool, row.Status)
		}
	}
	return nil
}

// NonFollowingViolations returns the tested non-following rows with hits or a
// timeout.
func NonFollowingViolations(r Report) []Row {
	var bad []Row
	for _, row := range r.Rows {
		if row.Follows {
			continue
		}
		hit := row.Status == StatusTested && row.Hits != nil && *row.Hits > 0
		if hit || row.Status == StatusTimedOut {
			bad = append(bad, row)
		}
	}
	return bad
}

// EvidenceFileName returns the evidence file name for goos.
func EvidenceFileName(goos string) (string, error) {
	if goos == "" {
		return "", errors.New("crawler: empty goos")
	}
	return "crawler-" + goos + ".json", nil
}

// WriteEvidence validates r and writes it as indented JSON to
// dir/crawler-<goos>.json.
func WriteEvidence(dir string, r Report) error {
	if dir == "" {
		return errors.New("crawler: empty evidence dir")
	}
	if err := r.Validate(); err != nil {
		return err
	}
	name, err := EvidenceFileName(r.GOOS)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("crawler: encode evidence: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0o644)
}
