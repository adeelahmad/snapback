package history

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/adeelahmad/snapback/internal/rawpath"
	"github.com/adeelahmad/snapback/internal/resolver"
)

const (
	infoName         = "info.json"
	stateOK          = "ok"
	stateUnavailable = "unavailable"
)

// info is the info.json document. It carries identifiers only, never
// repository locations or credentials.
type info struct {
	RootID    string         `json:"root_id"`
	Rel       rawpath.Path   `json:"rel"`
	Key       string         `json:"key"`
	RepoID    string         `json:"repo_id"`
	State     string         `json:"state"`
	Stale     bool           `json:"stale"`
	Snapshots []infoSnapshot `json:"snapshots"`
	Pending   []string       `json:"pending"`
}

type infoSnapshot struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
	Time  string `json:"time"`
}

// infoJSON renders info.json for d with the given state and linked snapshots.
func infoJSON(d Dir, state string, stale bool, linked []resolver.Eligible) ([]byte, error) {
	doc := info{
		RootID:    d.RootID,
		Rel:       rawpath.Path(d.Rel),
		Key:       d.Key,
		RepoID:    d.RepoID,
		State:     state,
		Stale:     stale,
		Snapshots: []infoSnapshot{},
		Pending:   []string{},
	}
	for _, e := range linked {
		s := infoSnapshot{ID: string(e.Snapshot.ID), Time: e.Snapshot.Time.UTC().Format(time.RFC3339)}
		for _, a := range d.Aliases.Aliases {
			if a.ID == e.Snapshot.ID {
				s.Alias = a.Name
				break
			}
		}
		doc.Snapshots = append(doc.Snapshots, s)
	}
	for _, id := range d.Pending {
		doc.Pending = append(doc.Pending, string(id))
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("history: encode info.json for %s: %w", d.Key, err)
	}
	return data, nil
}
