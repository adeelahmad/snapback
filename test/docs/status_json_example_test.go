package docs

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
	"github.com/adeelahmad/snapback/internal/status"
)

const (
	usageDoc         = "docs-site/usage.md"
	statusJSONAnchor = "status --json"
)

// firstJSONBlock returns the body of the first ```json fenced block that
// follows the first heading naming anchor.
func firstJSONBlock(t *testing.T, text, anchor string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	i := slices.IndexFunc(lines, func(l string) bool {
		return strings.HasPrefix(l, "#") && strings.Contains(l, anchor)
	})
	if i < 0 {
		t.Fatalf("%s has no heading naming %q", usageDoc, anchor)
	}
	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "```json" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "```" {
				return strings.Join(lines[i+1:j], "\n")
			}
		}
		t.Fatalf("%s has an unterminated ```json block after %q", usageDoc, anchor)
	}
	t.Fatalf("%s has no ```json block after the %q heading", usageDoc, anchor)
	return ""
}

// sortedKeys returns the object keys of b, sorted.
func sortedKeys(t *testing.T, what string, b []byte) []string {
	t.Helper()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json.Unmarshal(%s %s) error = %v", what, b, err)
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// populatedSnapshot mirrors internal/status's golden fixture: every field is
// set, so marshalling it emits each key including the optional ones.
func populatedSnapshot() status.Snapshot {
	at := time.Date(2026, 9, 22, 6, 0, 0, 0, time.UTC)
	id := provider.SnapshotID(strings.Repeat("a", 64))
	return status.Snapshot{
		State:         "degraded",
		Repos:         []status.Repo{{ID: "repoA", State: "failed", Code: errcode.RepoUnavailable}},
		LastRefresh:   at,
		Generation:    7,
		EligibleCount: map[string]int{"repoA": 3},
		Links:         2,
		Warm:          map[provider.SnapshotID]bool{id: true},
		Prewarm:       status.PrewarmSummary{Warm: 1, Cold: 2, Pending: 3, LastPrewarm: at},
		Pending:       []provider.SnapshotID{id},
		Discovery:     "running",
		Throttle:      []readerpolicy.ThrottleEvent{{PID: 42, Process: "mds", Rule: "deny", At: at}},
		WebURL:        "http://127.0.0.1:8080/",
		Recovery:      &status.RecoverySummary{Unmounted: []string{"/a"}, Foreign: []string{"/b"}},
	}
}

// TestUsageStatusJSONExampleMatchesWireKeys checks that the documented
// `status --json` example carries the ok/data envelope and that its data keys
// are exactly the keys a populated status.Snapshot marshals to.
func TestUsageStatusJSONExampleMatchesWireKeys(t *testing.T) {
	block := firstJSONBlock(t, readRepoFile(t, usageDoc), statusJSONAnchor)

	var envelope struct {
		OK   *bool           `json:"ok"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(block), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(%s example) error = %v", usageDoc, err)
	}
	if envelope.OK == nil || !*envelope.OK {
		t.Errorf("%s example ok = %v, want true", usageDoc, envelope.OK)
	}
	if len(envelope.Data) == 0 {
		t.Fatalf("%s example has no data object", usageDoc)
	}

	snap := populatedSnapshot()
	wire, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", snap, err)
	}

	got := sortedKeys(t, "example data", envelope.Data)
	want := sortedKeys(t, "status.Snapshot", wire)
	if !slices.Equal(got, want) {
		t.Errorf("%s example data keys = %v, want %v", usageDoc, got, want)
	}
}
