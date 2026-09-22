package mount

import (
	"log/slog"
	"slices"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging/logtest"
)

// logCatalogMsg is the message every catalog read record carries.
const logCatalogMsg = "catalog"

func newLogCatalog(t *testing.T, level slog.Level) (LogCatalog, *logtest.Log) {
	t.Helper()
	log, lg := logtest.Capture(t, level)
	return LogCatalog{Log: log, Catalog: newFakeCatalog()}, lg
}

func onlyRecord(t *testing.T, lg *logtest.Log) map[string]any {
	t.Helper()
	records := lg.Records()
	if len(records) != 1 {
		t.Fatalf("got %d log records, want exactly 1: %q", len(records), lg.Text())
	}
	rec := records[0]
	if got := rec["level"]; got != "DEBUG" {
		t.Errorf("level = %v, want DEBUG", got)
	}
	if got := rec["msg"]; got != logCatalogMsg {
		t.Errorf("msg = %v, want %q", got, logCatalogMsg)
	}
	return rec
}

func wantAttr(t *testing.T, rec map[string]any, key string, want any) {
	t.Helper()
	got, ok := rec[key]
	if !ok {
		t.Errorf("record has no attr %q: %v", key, rec)
		return
	}
	if got != want {
		t.Errorf("attr %q = %#v, want %#v", key, got, want)
	}
}

func TestLogCatalogLookupHit(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	ino, kind, found := cat.Lookup(RootIno, "docs")

	if ino != 2 || kind != KindDir || !found {
		t.Errorf("Lookup = (%d, %d, %v), want (2, %d, true)", ino, kind, found, KindDir)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "lookup")
	wantAttr(t, rec, "parent", float64(RootIno))
	wantAttr(t, rec, "name", "docs")
	wantAttr(t, rec, "found", true)
}

func TestLogCatalogLookupMiss(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	ino, kind, found := cat.Lookup(RootIno, "absent")

	if ino != 0 || kind != 0 || found {
		t.Errorf("Lookup = (%d, %d, %v), want (0, 0, false)", ino, kind, found)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "lookup")
	wantAttr(t, rec, "parent", float64(RootIno))
	wantAttr(t, rec, "name", "absent")
	wantAttr(t, rec, "found", false)
}

func TestLogCatalogReadDirCountsEntries(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	names, found := cat.ReadDir(RootIno)

	if !found || !slices.Equal(names, []string{"docs", "link"}) {
		t.Errorf("ReadDir = (%v, %v), want ([docs link], true)", names, found)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "readdir")
	wantAttr(t, rec, "ino", float64(RootIno))
	wantAttr(t, rec, "found", true)
	wantAttr(t, rec, "n", float64(len(names)))
}

func TestLogCatalogReadDirMiss(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	names, found := cat.ReadDir(99)

	if found || names != nil {
		t.Errorf("ReadDir = (%v, %v), want (nil, false)", names, found)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "readdir")
	wantAttr(t, rec, "ino", float64(99))
	wantAttr(t, rec, "found", false)
	wantAttr(t, rec, "n", float64(0))
}

func TestLogCatalogReadlink(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	target, found := cat.Readlink(3)

	if target != "../x" || !found {
		t.Errorf("Readlink = (%q, %v), want (\"../x\", true)", target, found)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "readlink")
	wantAttr(t, rec, "ino", float64(3))
	wantAttr(t, rec, "found", true)
}

func TestLogCatalogReadFileMiss(t *testing.T) {
	cat, lg := newLogCatalog(t, slog.LevelDebug)

	data, found := cat.ReadFile(2)

	if data != nil || found {
		t.Errorf("ReadFile = (%q, %v), want (nil, false)", data, found)
	}
	rec := onlyRecord(t, lg)
	wantAttr(t, rec, "op", "read")
	wantAttr(t, rec, "ino", float64(2))
	wantAttr(t, rec, "found", false)
}

// TestLogCatalogLevels pins that every read logs exactly one record at debug
// level and nothing at info level.
func TestLogCatalogLevels(t *testing.T) {
	exercise := func(cat LogCatalog) {
		cat.Lookup(RootIno, "docs")
		cat.ReadDir(RootIno)
		cat.Readlink(3)
		cat.ReadFile(2)
	}

	debugCat, debugLog := newLogCatalog(t, slog.LevelDebug)
	exercise(debugCat)
	if got := len(debugLog.Records()); got != 4 {
		t.Errorf("debug level: got %d records, want 4: %q", got, debugLog.Text())
	}

	infoCat, infoLog := newLogCatalog(t, slog.LevelInfo)
	exercise(infoCat)
	if got := len(infoLog.Records()); got != 0 {
		t.Errorf("info level: got %d records, want 0: %q", got, infoLog.Text())
	}
}
