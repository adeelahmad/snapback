package status

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/provider"
)

func TestSummarizePrewarm(t *testing.T) {
	at := time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		results []provider.PrewarmResult
		pending int
		want    PrewarmSummary
	}{
		{
			name:    "none",
			results: nil,
			pending: 0,
			want:    PrewarmSummary{LastPrewarm: at},
		},
		{
			name: "mixed",
			results: []provider.PrewarmResult{
				{ID: "a", Warm: true},
				{ID: "b", Warm: true},
				{ID: "c", Err: errors.New("timeout")},
			},
			pending: 1,
			want:    PrewarmSummary{Warm: 2, Cold: 1, Pending: 1, LastPrewarm: at},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SummarizePrewarm(tt.results, tt.pending, at)
			if got != tt.want {
				t.Errorf("SummarizePrewarm(%v, %d, %v) = %+v, want %+v", tt.results, tt.pending, at, got, tt.want)
			}
		})
	}
}

func TestSnapshotJSONHasPrewarmSummary(t *testing.T) {
	at := time.Date(2026, 9, 22, 8, 0, 0, 0, time.UTC)
	s := Snapshot{
		State:   "ready",
		Warm:    map[provider.SnapshotID]bool{"a": true},
		Prewarm: PrewarmSummary{Warm: 2, Cold: 1, Pending: 3, LastPrewarm: at},
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal(Snapshot) = %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal(%s) = %v", b, err)
	}
	if _, ok := got["warm"]; !ok {
		t.Errorf("Snapshot JSON = %s, want the Warm map kept", b)
	}
	raw, ok := got["prewarm"]
	if !ok {
		t.Fatalf("Snapshot JSON = %s, want key %q", b, "prewarm")
	}
	var sum struct {
		Warm        *int       `json:"warm"`
		Cold        *int       `json:"cold"`
		Pending     *int       `json:"pending"`
		LastPrewarm *time.Time `json:"last_prewarm"`
	}
	if err := json.Unmarshal(raw, &sum); err != nil {
		t.Fatalf("json.Unmarshal(prewarm %s) = %v", raw, err)
	}
	if sum.Warm == nil || *sum.Warm != 2 || sum.Cold == nil || *sum.Cold != 1 ||
		sum.Pending == nil || *sum.Pending != 3 || sum.LastPrewarm == nil || !sum.LastPrewarm.Equal(at) {
		t.Errorf("prewarm JSON = %s, want warm=2 cold=1 pending=3 last_prewarm=%s", raw, at.Format(time.RFC3339))
	}
}
