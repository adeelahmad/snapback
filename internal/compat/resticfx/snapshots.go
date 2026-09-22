package resticfx

import "time"

// Snapshot is one entry of `restic snapshots --json` output.
type Snapshot struct {
	ID       string
	Time     time.Time
	Hostname string
	Paths    []string
}

// ParseSnapshots decodes `restic snapshots --json` output, enforcing full 64-hex IDs.
func ParseSnapshots(data []byte) ([]Snapshot, error) {
	panic("SUB-AGENT-TODO: json-decode snapshots; reject entries whose id is missing/short/non-hex (^[0-9a-f]{64}$); ignore short_id")
}

// CheckObservedIDs passes only when entries contains fullID exactly and no truncated name.
func CheckObservedIDs(entries []string, fullID string) error {
	panic("SUB-AGENT-TODO: error on empty entries or invalid fullID; require exact fullID; fail if any entry is a strict prefix of a full ID")
}
