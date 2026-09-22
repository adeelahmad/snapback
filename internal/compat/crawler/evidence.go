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

// Row is the evidence for one crawler tool.
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

// Report is the crawler evidence for one platform.
type Report struct {
	SchemaVersion       int
	GOOS                string
	GOARCH              string
	Seed                Shape
	PositiveControlHits int
	Rows                []Row
}

// Validate reports the first row that breaks the evidence rules.
func (r Report) Validate() error {
	panic("SUB-AGENT-TODO: reject a not-tested-here row that carries a hit count or lacks a reason, and a tested row with no hit count (a skipped tool must never be folded into a zero-hit pass); return the first violation, nil otherwise")
}

// NonFollowingViolations returns the tested non-following rows with hits or a
// timeout.
func NonFollowingViolations(r Report) []Row {
	panic("SUB-AGENT-TODO: return the non-following rows that are tested with hits > 0 or have status timed-out; ignore following and not-tested-here rows")
}

// EvidenceFileName returns the evidence file name for goos.
func EvidenceFileName(goos string) (string, error) {
	panic("SUB-AGENT-TODO: return crawler-<goos>.json; reject an empty goos with an error")
}

// WriteEvidence validates r and writes it as indented JSON to
// dir/crawler-<goos>.json.
func WriteEvidence(dir string, r Report) error {
	panic("SUB-AGENT-TODO: error on empty dir; r.Validate(); EvidenceFileName(r.GOOS); write indented JSON with a trailing newline to dir/crawler-<goos>.json (add snake_case JSON tags to Row and Report per plan.md T3: tool, argv, follows, status, hits omitempty, ops, version, exit_code, duration, reason; schema_version, goos, goarch, seed, positive_control_hits, rows)")
}
