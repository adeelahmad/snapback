// agentic:shim

package resticfx

import "time"

// Snapshot is a compile shim for S2-03/T3; the scaffolder replaces it.
type Snapshot struct {
	ID       string
	Time     time.Time
	Hostname string
	Paths    []string
}

// ParseSnapshots is a deliberately wrong compile shim for S2-03/T3.
func ParseSnapshots(_ []byte) ([]Snapshot, error) {
	return nil, nil
}

// CheckObservedIDs is a deliberately wrong compile shim for S2-03/T3.
func CheckObservedIDs(_ []string, _ string) error {
	return nil
}
