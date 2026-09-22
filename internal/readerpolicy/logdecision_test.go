package readerpolicy

import (
	"log/slog"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging/logtest"
)

// loggedPolicy wraps a Policy over cfg and procs in a LogDecider logging to log.
func loggedPolicy(cfg Config, procs map[uint32]string, log *slog.Logger) *LogDecider {
	return NewLogDecider(New(cfg, fakeProcs(procs), fixedNow(t0)), fakeProcs(procs), log)
}

// denyCfg denies "rg" with throttling off, so every deny has rule "deny".
var denyCfg = Config{Deny: []string{"rg"}}

// denyProcs names pid 1 as the denied reader and pid 3 as an allowed one.
var denyProcs = map[uint32]string{1: "rg", 3: "zsh"}

// wantAttrs checks rec carries the four decision attributes.
func wantAttrs(t *testing.T, rec map[string]any, decision, reason, process string, pid float64) {
	t.Helper()
	for _, a := range []struct {
		key  string
		want any
	}{
		{"decision", decision},
		{"reason", reason},
		{"process", process},
		{"pid", pid},
	} {
		got, ok := rec[a.key]
		if !ok {
			t.Errorf("record %v has no %q attribute", rec, a.key)
			continue
		}
		if got != a.want {
			t.Errorf("record %q = %#v, want %#v", a.key, got, a.want)
		}
	}
}

func TestLogDeciderAllowAtDebugEmitsOneDebugRecord(t *testing.T) {
	log, lg := logtest.Capture(t, slog.LevelDebug)
	p := loggedPolicy(denyCfg, denyProcs, log)

	if got := p.Allow(lookup(3, "/a")); !got {
		t.Fatalf("Allow(lookup pid 3 %q) = %v, want true", denyProcs[3], got)
	}

	recs := lg.Records()
	if len(recs) != 1 {
		t.Fatalf("allow at debug emitted %d records, want 1: %v", len(recs), recs)
	}
	if got := recs[0]["level"]; got != "DEBUG" {
		t.Errorf("allow record level = %v, want DEBUG", got)
	}
	wantAttrs(t, recs[0], "allow", "allowed", "zsh", 3)
}

func TestLogDeciderDenyAtDebugEmitsDebugAndInfo(t *testing.T) {
	log, lg := logtest.Capture(t, slog.LevelDebug)
	p := loggedPolicy(denyCfg, denyProcs, log)

	if got := p.Allow(lookup(1, "/a")); got {
		t.Fatalf("Allow(lookup pid 1 %q) = %v, want false", denyProcs[1], got)
	}

	recs := lg.Records()
	if len(recs) != 2 {
		t.Fatalf("deny at debug emitted %d records, want 2 (debug + info): %v", len(recs), recs)
	}
	for i, wantLevel := range []string{"DEBUG", "INFO"} {
		if got := recs[i]["level"]; got != wantLevel {
			t.Errorf("deny record %d level = %v, want %v", i, got, wantLevel)
		}
		wantAttrs(t, recs[i], "deny", "deny", "rg", 1)
	}
}

func TestLogDeciderDenyAtInfoEmitsOnlyTheInfoRecord(t *testing.T) {
	log, lg := logtest.Capture(t, slog.LevelInfo)
	p := loggedPolicy(denyCfg, denyProcs, log)

	if got := p.Allow(lookup(1, "/a")); got {
		t.Fatalf("Allow(lookup pid 1 %q) = %v, want false", denyProcs[1], got)
	}

	recs := lg.Records()
	if len(recs) != 1 {
		t.Fatalf("deny at info emitted %d records, want 1: %v", len(recs), recs)
	}
	if got := recs[0]["level"]; got != "INFO" {
		t.Errorf("deny record level = %v, want INFO", got)
	}
	wantAttrs(t, recs[0], "deny", "deny", "rg", 1)
}

func TestLogDeciderAllowAtInfoEmitsNothing(t *testing.T) {
	log, lg := logtest.Capture(t, slog.LevelInfo)
	p := loggedPolicy(denyCfg, denyProcs, log)

	if got := p.Allow(lookup(3, "/a")); !got {
		t.Fatalf("Allow(lookup pid 3 %q) = %v, want true", denyProcs[3], got)
	}

	if recs := lg.Records(); len(recs) != 0 {
		t.Errorf("allow at info emitted %d records, want 0: %v", len(recs), recs)
	}
}

func TestLogDeciderLeavesDecisionsUnchanged(t *testing.T) {
	cfg := Config{Deny: []string{"rg", "fd", "mdworker_shared_x"}}
	procs := map[uint32]string{
		1: "rg",
		2: "/usr/bin/fd",
		3: "zsh",
		4: "mdworker_shared",
		5: "rgx",
	}
	log, lg := logtest.Capture(t, slog.LevelDebug)
	p := loggedPolicy(cfg, procs, log)

	tests := []struct {
		pid  uint32
		why  string
		want bool
	}{
		{1, "exact name", false},
		{2, "basename match", false},
		{3, "not listed", true},
		{4, "15-byte comm truncation", false},
		{5, "no prefix match", true},
	}
	for _, tt := range tests {
		if got := p.Allow(lookup(tt.pid, "/a")); got != tt.want {
			t.Errorf("Allow(lookup pid %d %q, %s) = %v, want %v",
				tt.pid, procs[tt.pid], tt.why, got, tt.want)
		}
	}

	// Two allows log one record each; three denies log two each.
	if recs := lg.Records(); len(recs) != 8 {
		t.Errorf("table emitted %d records, want 8 (2 allows + 3 denies x 2): %v", len(recs), recs)
	}
}

func TestLogDeciderWithNilLoggerIsQuiet(t *testing.T) {
	p := loggedPolicy(denyCfg, denyProcs, nil)

	if got := p.Allow(lookup(1, "/a")); got {
		t.Errorf("Allow(lookup pid 1 %q) with a nil logger = %v, want false", denyProcs[1], got)
	}
	if got := len(p.Events()); got != 1 {
		t.Errorf("Events() after one deny = %d, want 1", got)
	}
}
