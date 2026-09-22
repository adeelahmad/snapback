package status

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/provider"
	"github.com/adeelahmad/snapback/internal/readerpolicy"
)

// snapshotKeys is the public wire shape of a status snapshot: every key is
// snake_case, appears exactly once and no other key is emitted.
var snapshotKeys = []string{
	"discovery",
	"eligible_count",
	"generation",
	"last_refresh",
	"links",
	"pending",
	"prewarm",
	"recovery",
	"repos",
	"state",
	"throttle",
	"warm",
	"web_url",
}

// populatedSnapshot returns a snapshot with every field set, so marshalling it
// exercises each key including the optional ones.
func populatedSnapshot() Snapshot {
	at := time.Date(2026, 9, 22, 6, 0, 0, 0, time.UTC)
	id := provider.SnapshotID(strings.Repeat("a", 64))
	return Snapshot{
		State:         "degraded",
		Repos:         []Repo{{ID: "repoA", State: "failed", Code: errcode.RepoUnavailable}},
		LastRefresh:   at,
		Generation:    7,
		EligibleCount: map[string]int{"repoA": 3},
		Links:         2,
		Warm:          map[provider.SnapshotID]bool{id: true},
		Prewarm:       PrewarmSummary{Warm: 1, Cold: 2, Pending: 3, LastPrewarm: at},
		Pending:       []provider.SnapshotID{id},
		Discovery:     "running",
		Throttle:      []readerpolicy.ThrottleEvent{{PID: 42, Process: "mds", Rule: "deny", At: at}},
		WebURL:        "http://127.0.0.1:8080/",
		Recovery:      &RecoverySummary{Unmounted: []string{"/a"}, Foreign: []string{"/b"}},
	}
}

// objectKeys decodes b as a JSON object and returns its keys in the order the
// document lists them, so duplicates survive for the caller to detect.
func objectKeys(t *testing.T, b []byte) []string {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(string(b)))
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("json.Decoder.Token(%s) error = %v", b, err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		t.Fatalf("json.Decoder.Token(%s) = %v, want '{'", b, tok)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("json.Decoder.Token(%s) error = %v", b, err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("json.Decoder.Token(%s) = %v, want a key", b, tok)
		}
		keys = append(keys, key)
		if err := skipValue(dec); err != nil {
			t.Fatalf("skipValue(%s) error = %v", b, err)
		}
	}
	return keys
}

// skipValue consumes the next value, descending into objects and arrays.
func skipValue(dec *json.Decoder) error {
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := tok.(json.Delim); ok {
			switch delim {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
		if depth == 0 {
			return nil
		}
	}
}

// checkKeys asserts the object in b has exactly want, with no duplicates.
func checkKeys(t *testing.T, what string, b []byte, want []string) {
	t.Helper()
	got := objectKeys(t, b)
	sorted := slices.Clone(got)
	slices.Sort(sorted)
	if len(slices.Compact(slices.Clone(sorted))) != len(sorted) {
		t.Errorf("%s keys = %v, want no duplicate key", what, got)
	}
	if !slices.Equal(sorted, want) {
		t.Errorf("%s keys = %v, want %v", what, sorted, want)
	}
}

func TestSnapshotJSONKeysAreSnakeCase(t *testing.T) {
	snap := populatedSnapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", snap, err)
	}
	checkKeys(t, "json.Marshal(snapshot)", b, snapshotKeys)

	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", b, err)
	}
	for _, mixed := range []string{"State", "Repos", "LastRefresh", "Generation", "EligibleCount", "Links", "Warm", "Pending", "Discovery", "Throttle", "WebURL", "Recovery"} {
		if _, ok := top[mixed]; ok {
			t.Errorf("json.Marshal(snapshot) = %s, want no key %q", b, mixed)
		}
	}
}

func TestSnapshotNestedJSONKeysAreSnakeCase(t *testing.T) {
	snap := populatedSnapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal(%+v) error = %v", snap, err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", b, err)
	}

	nested := []struct {
		key  string
		want []string
	}{
		{"prewarm", []string{"cold", "last_prewarm", "pending", "warm"}},
		{"recovery", []string{"foreign", "unmounted"}},
	}
	for _, tt := range nested {
		raw, ok := top[tt.key]
		if !ok {
			t.Errorf("json.Marshal(snapshot) = %s, want key %q", b, tt.key)
			continue
		}
		checkKeys(t, tt.key, raw, tt.want)
	}

	arrays := []struct {
		key  string
		want []string
	}{
		{"repos", []string{"code", "id", "state"}},
		{"throttle", []string{"at", "pid", "process", "rule"}},
	}
	for _, tt := range arrays {
		raw, ok := top[tt.key]
		if !ok {
			t.Errorf("json.Marshal(snapshot) = %s, want key %q", b, tt.key)
			continue
		}
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			t.Fatalf("json.Unmarshal(%s %s) error = %v", tt.key, raw, err)
		}
		if len(items) != 1 {
			t.Fatalf("%s JSON = %s, want 1 entry", tt.key, raw)
		}
		checkKeys(t, tt.key+"[0]", items[0], tt.want)
	}
}
