package refresh

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/provider"
)

// logRecord is one decoded slog record: its message, its attribute keys in
// the order the handler wrote them, and the attribute values.
type logRecord struct {
	msg   string
	keys  []string
	attrs map[string]any
}

// jsonLog returns a debug-or-above JSON logger writing into buf, with the
// timestamp dropped so records compare byte-for-byte.
func jsonLog(level slog.Level) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(h), &buf
}

// decodeRecords decodes every JSON line in buf, keeping attribute order.
func decodeRecords(t *testing.T, buf *bytes.Buffer) []logRecord {
	t.Helper()
	var out []logRecord
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		if _, err := dec.Token(); err != nil {
			t.Fatalf("decode %q: %v", line, err)
		}
		rec := logRecord{attrs: map[string]any{}}
		for dec.More() {
			tok, err := dec.Token()
			if err != nil {
				t.Fatalf("decode key in %q: %v", line, err)
			}
			key, ok := tok.(string)
			if !ok {
				t.Fatalf("decode %q: key %v is not a string", line, tok)
			}
			var val any
			if err := dec.Decode(&val); err != nil {
				t.Fatalf("decode value for %q in %q: %v", key, line, err)
			}
			switch key {
			case slog.LevelKey:
			case slog.MessageKey:
				rec.msg, _ = val.(string)
			default:
				rec.keys = append(rec.keys, key)
				rec.attrs[key] = val
			}
		}
		out = append(out, rec)
	}
	return out
}

func messages(recs []logRecord) []string {
	out := make([]string, len(recs))
	for i, r := range recs {
		out[i] = r.msg
	}
	return out
}

// num reads an attribute the JSON decoder produced as a json.Number.
func num(t *testing.T, rec logRecord, key string) int64 {
	t.Helper()
	raw, ok := rec.attrs[key].(json.Number)
	if !ok {
		t.Fatalf("record %q attr %q = %#v, want a number", rec.msg, key, rec.attrs[key])
	}
	n, err := raw.Int64()
	if err != nil {
		t.Fatalf("record %q attr %q = %q: %v", rec.msg, key, raw, err)
	}
	return n
}

// logCycle runs one refresh cycle (Refresh then Prewarm) over the standard
// rig with one repository of three snapshots and exactly one prewarm, and
// returns the records the refresher emitted at level.
func logCycle(t *testing.T, level slog.Level) []logRecord {
	t.Helper()
	ids := []provider.SnapshotID{idA, idB, idC}
	rg := newRig(t, ids, ids, func(c *Config, _ map[string]provider.Lister) {
		c.PrewarmSnapshots = 1
		c.PrewarmConcurrency = 1
	})
	log, buf := jsonLog(level)
	rg.r.WithLog(log)

	ctx := context.Background()
	if _, err := rg.r.Refresh(ctx); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got := rg.r.Prewarm(ctx); len(got) != 1 {
		t.Fatalf("Prewarm warmed %d snapshots, want 1", len(got))
	}
	return decodeRecords(t, buf)
}

// TestLogRefreshCycleDebugRecords pins the exact debug records one refresh
// cycle emits: cycle start, one per repository, one per prewarm start and
// finish, and one for the eviction of the previous warm set.
func TestLogRefreshCycleDebugRecords(t *testing.T) {
	recs := logCycle(t, slog.LevelDebug)

	wantMsgs := []string{
		"refresh start",
		"refresh repo",
		"prewarm start",
		"prewarm done",
		"refresh evict",
	}
	if got := messages(recs); !slices.Equal(got, wantMsgs) {
		t.Fatalf("cycle messages = %q, want %q", got, wantMsgs)
	}

	wantKeys := [][]string{
		{"dirs", "repos"},
		{"repo", "snapshots", "visible"},
		{"snapshot"},
		{"snapshot", "dur_ms", "warm"},
		{"evicted", "warm"},
	}
	for i, rec := range recs {
		if !slices.Equal(rec.keys, wantKeys[i]) {
			t.Errorf("record %d (%q) attrs = %q, want %q in that order", i, rec.msg, rec.keys, wantKeys[i])
		}
	}

	start := recs[0]
	if got := num(t, start, "dirs"); got != 1 {
		t.Errorf("refresh start dirs = %d, want 1", got)
	}
	if got := num(t, start, "repos"); got != 1 {
		t.Errorf("refresh start repos = %d, want 1", got)
	}

	repo := recs[1]
	if got, want := repo.attrs["repo"], testRepo; got != want {
		t.Errorf("refresh repo repo = %v, want %q", got, want)
	}
	if got := num(t, repo, "snapshots"); got != 3 {
		t.Errorf("refresh repo snapshots = %d, want 3", got)
	}
	if got := num(t, repo, "visible"); got != 3 {
		t.Errorf("refresh repo visible = %d, want 3", got)
	}

	for _, i := range []int{2, 3} {
		if got, want := recs[i].attrs["snapshot"], string(idC); got != want {
			t.Errorf("record %d (%q) snapshot = %v, want the newest snapshot %q", i, recs[i].msg, got, want)
		}
	}
	if got := num(t, recs[3], "dur_ms"); got < 0 {
		t.Errorf("prewarm done dur_ms = %d, want >= 0", got)
	}
	if got, want := recs[3].attrs["warm"], true; got != want {
		t.Errorf("prewarm done warm = %v, want %v", got, want)
	}

	evict := recs[4]
	if got := num(t, evict, "evicted"); got != 0 {
		t.Errorf("refresh evict evicted = %d, want 0 on the first cycle", got)
	}
	if got := num(t, evict, "warm"); got != 1 {
		t.Errorf("refresh evict warm = %d, want 1", got)
	}
}

// TestLogRefreshInfoLevelSilent pins that a refresh cycle adds nothing at
// info: the one info-level "refresh" summary line is the daemon's, and the
// refresh package itself stays silent above debug.
func TestLogRefreshInfoLevelSilent(t *testing.T) {
	if recs := logCycle(t, slog.LevelInfo); len(recs) != 0 {
		t.Errorf("cycle at info emitted %d records (%q), want none", len(recs), messages(recs))
	}
}
