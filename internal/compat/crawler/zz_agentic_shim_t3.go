// agentic:shim

package crawler

import "time"

// Row statuses. A skipped tool is never folded into a zero-hit pass.
const (
	StatusTested        Status = "tested"
	StatusNotTestedHere Status = "not-tested-here"
	StatusTimedOut      Status = "timed-out"
)

// SchemaVersion is the evidence schema version written into every Report.
const SchemaVersion = 1

// Row is the evidence for one crawler tool. The JSON tags are part of the
// contract (snake_case: tool, argv, follows, status, hits, ops, version,
// exit_code, duration, reason; hits omitted or null when nil); the shim
// deliberately leaves them off.
type Row struct {
	Tool     string
	Argv     []string
	Follows  bool
	Status   Status
	Hits     *int
	Ops      map[string]int
	Version  string
	ExitCode int
	Duration time.Duration
	Reason   string
}

// Report is the crawler evidence for one platform. JSON keys:
// schema_version, goos, goarch, seed, positive_control_hits, rows.
type Report struct {
	SchemaVersion       int
	GOOS                string
	GOARCH              string
	Seed                Shape
	PositiveControlHits int
	Rows                []Row
}

// Validate reports the first row that breaks the evidence rules.
func (r Report) Validate() error { return nil }

// NonFollowingViolations returns the tested non-following rows with hits or a
// timeout.
func NonFollowingViolations(r Report) []Row { return r.Rows }

// EvidenceFileName returns the evidence file name for goos.
func EvidenceFileName(goos string) (string, error) { return "evidence.json", nil }

// WriteEvidence validates r and writes it as indented JSON to
// dir/crawler-<goos>.json.
func WriteEvidence(dir string, r Report) error { return nil }
