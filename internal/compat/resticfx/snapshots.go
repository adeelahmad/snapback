package resticfx

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

var fullIDPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Snapshot is one entry of `restic snapshots --json` output.
type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Paths    []string  `json:"paths"`
}

// ParseSnapshots decodes `restic snapshots --json` output, enforcing full 64-hex IDs.
func ParseSnapshots(data []byte) ([]Snapshot, error) {
	var snaps []Snapshot
	if err := json.Unmarshal(data, &snaps); err != nil {
		return nil, fmt.Errorf("decode snapshots json: %w", err)
	}
	for i, s := range snaps {
		if !fullIDPattern.MatchString(s.ID) {
			return nil, fmt.Errorf("snapshot %d: id %q is not a full 64-hex id", i, s.ID)
		}
	}
	return snaps, nil
}

// CheckObservedIDs passes only when entries contains fullID exactly and no truncated name.
func CheckObservedIDs(entries []string, fullID string) error {
	if len(entries) == 0 {
		return errors.New("no ids/ entries observed")
	}
	if !fullIDPattern.MatchString(fullID) {
		return fmt.Errorf("expected id %q is not a full 64-hex id", fullID)
	}
	if !slices.Contains(entries, fullID) {
		return fmt.Errorf("full id %s not found in ids/ entries %q", fullID, entries)
	}
	for _, e := range entries {
		if len(e) < len(fullID) && strings.HasPrefix(fullID, e) {
			return fmt.Errorf("ids/ entry %q is a truncated prefix of %s", e, fullID)
		}
	}
	return nil
}
