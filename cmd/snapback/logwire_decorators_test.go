package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/mount"
)

// logRecord is one line of the daemon's JSON log.
type logRecord struct {
	Level    string `json:"level"`
	Msg      string `json:"msg"`
	Decision string `json:"decision"`
}

// decodeLog parses buf as one JSON log record per line.
func decodeLog(t *testing.T, buf *bytes.Buffer) []logRecord {
	t.Helper()
	var recs []logRecord
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var rec logRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("json.Unmarshal(log line %q) = %v, want nil error", line, err)
		}
		recs = append(recs, rec)
	}
	return recs
}

// hasMsg reports whether recs carry a record with msg at level.
func hasMsg(recs []logRecord, level, msg string) bool {
	for _, rec := range recs {
		if rec.Msg == msg && rec.Level == level {
			return true
		}
	}
	return false
}

// hasDeny reports whether recs carry an info-level deny decision.
func hasDeny(recs []logRecord) bool {
	for _, rec := range recs {
		if rec.Level == "INFO" && rec.Decision == "deny" {
			return true
		}
	}
	return false
}

// stubCatalog answers every lookup with a miss; it logs nothing itself, so
// any "catalog" record comes from the decorator.
type stubCatalog struct{}

func (stubCatalog) Lookup(uint64, string) (uint64, mount.Kind, bool) { return 0, 0, false }
func (stubCatalog) ReadDir(uint64) ([]string, bool)                  { return nil, false }
func (stubCatalog) Readlink(uint64) (string, bool)                   { return "", false }
func (stubCatalog) ReadFile(uint64) ([]byte, bool)                   { return nil, false }

// driveLogWiring builds the production wiring at level, drives one restic
// call, one catalog lookup, one reader decision and one refresh through it,
// and returns the records the logger emitted plus whether the reader was
// denied. The policy can only deny a reader it can name, which needs /proc,
// so the deny records are pinned only where the decision came back deny.
func driveLogWiring(t *testing.T, level slog.Level) ([]logRecord, bool) {
	t.Helper()
	tmp := shortTempDir(t)
	bin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) = %v", bin, err)
	}
	resticBin := filepath.Join(bin, "restic")
	if err := os.WriteFile(resticBin, []byte(fakeRestic), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%s) = %v", resticBin, err)
	}
	t.Setenv("PATH", bin)
	cfg := daemonDepsConfig(t, tmp, resticBin)
	cfg.Catalog.ReaderPolicy.DenyProcesses = []string{filepath.Base(os.Args[0])}
	ln := listenUnix(t, tmp)

	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level}))

	w, err := daemonBuilderWithLog(t.Context(), cfg, ln, log)
	if err != nil {
		t.Fatalf("daemonBuilderWithLog(ctx, cfg, ln, log) = %v, want nil error", err)
	}

	lister, ok := w.Listers["personal"]
	if !ok {
		t.Fatalf("daemonBuilderWithLog(...).Listers has no %q, want one per repository", "personal")
	}
	if _, err := lister.List(t.Context()); err != nil {
		t.Fatalf("Listers[%q].List(ctx) = %v, want nil error", "personal", err)
	}

	if w.Catalog == nil {
		t.Fatal("daemonBuilderWithLog(...).Catalog = nil, want the history view's catalog wrapper")
	}
	w.Catalog(stubCatalog{}).Lookup(mount.RootIno, "2024-01-01T00:00:00Z")

	if w.Gate == nil {
		t.Fatal("daemonBuilderWithLog(...).Gate = nil, want the reader-policy gate")
	}
	allowed := w.Gate.Allow(mount.Event{Op: mount.OpLookup, Path: "x", PID: uint32(os.Getpid())})

	if _, err := w.Deps.Refresher.Refresh(t.Context()); err != nil {
		t.Logf("Refresher.Refresh(ctx) = %v (tolerated; the records are what is pinned)", err)
	}
	return decodeLog(t, &buf), !allowed
}

// TestLogWireDecoratorsDebug pins that the production daemon wiring routes the
// S5-36 debug decorators at the daemon's own logger: a restic invocation, a
// catalog read, a reader-policy deny and a refresh each leave a record.
func TestLogWireDecoratorsDebug(t *testing.T) {
	recs, denied := driveLogWiring(t, slog.LevelDebug)

	if !hasMsg(recs, "DEBUG", "restic exec") {
		t.Errorf("debug log = %+v, want a DEBUG %q record from restic.LogRunner", recs, "restic exec")
	}
	if !hasMsg(recs, "DEBUG", "catalog") {
		t.Errorf("debug log = %+v, want a DEBUG %q record from mount.LogCatalog", recs, "catalog")
	}
	if !hasMsg(recs, "DEBUG", "reader policy decision") {
		t.Errorf("debug log = %+v, want a DEBUG %q record from readerpolicy.LogDecider", recs, "reader policy decision")
	}
	if denied && !hasDeny(recs) {
		t.Errorf("debug log = %+v, want an INFO decision=deny record from readerpolicy.LogDecider", recs)
	}
	if !hasMsg(recs, "DEBUG", "refresh start") {
		t.Errorf("debug log = %+v, want a DEBUG %q record from (*refresh.Refresher).WithLog", recs, "refresh start")
	}
}

// TestLogWireDecoratorsInfo pins that at info level the same wiring emits the
// deny record and none of the debug ones, so the decorators cost nothing when
// debug is off.
func TestLogWireDecoratorsInfo(t *testing.T) {
	recs, denied := driveLogWiring(t, slog.LevelInfo)

	if denied && !hasDeny(recs) {
		t.Errorf("info log = %+v, want an INFO decision=deny record from readerpolicy.LogDecider", recs)
	}
	for _, msg := range []string{"restic exec", "restic done", "catalog", "refresh start", "reader policy decision"} {
		if hasMsg(recs, "DEBUG", msg) {
			t.Errorf("info log = %+v, want no DEBUG %q record", recs, msg)
		}
	}
}
