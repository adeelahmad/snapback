// agentic:shim
package resticfx

import (
	"errors"
	"os"
	"time"
)

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

func EvidenceDir(getenv func(string) string) string {
	return "shim-wrong-evidence-dir"
}

func WriteEvidence(dir string, ev Evidence) (string, error) {
	return "", errors.New("agentic shim: WriteEvidence not implemented")
}

func ReadEvidence(path string) (Evidence, error) {
	return Evidence{}, errors.New("agentic shim: ReadEvidence not implemented")
}

type Probe struct {
	Getenv           func(string) string
	LookPath         func(string) (string, error)
	Stat             func(string) (os.FileInfo, error)
	GOOS             string
	ResticVersionOut string
}

func MissingPrerequisite(p Probe) string {
	return "agentic shim: prerequisite missing"
}
