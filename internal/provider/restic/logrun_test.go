package restic

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/logging/logtest"
)

// newLogRunner returns a LogRunner over fr at level, plus the Log the records
// land in.
func newLogRunner(t *testing.T, fr *fakeRunner, level slog.Level) (LogRunner, *logtest.Log) {
	t.Helper()
	log, lg := logtest.Capture(t, level)
	return LogRunner{Log: log, Runner: fr}, lg
}

// str returns the record's field as a string, failing when it is absent.
func str(t *testing.T, r map[string]any, key string) string {
	t.Helper()
	v, ok := r[key]
	if !ok {
		t.Fatalf("record %v has no %q field", r, key)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("record field %q = %v (%T), want a string", key, v, v)
	}
	return s
}

// logTestArgs is a restic invocation carrying a secret password file path.
const passwordFilePath = "/tmp/fake-secret-dir/restic.pass"

func logTestArgs() []string {
	return []string{"--password-file", passwordFilePath, "--json", "snapshots"}
}

func logTestEnv() []string {
	return []string{
		"PATH=/usr/bin",
		"RESTIC_REPOSITORY=/srv/repo",
		"RESTIC_PASSWORD=hunter2-plaintext",
		"AWS_SECRET_ACCESS_KEY=abc123secretkey",
	}
}

func TestLogRunnerLogsBeforeAndAfterAtDebug(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stdout: []byte("[]"), stderr: []byte("warn")}}}
	lr, lg := newLogRunner(t, fr, slog.LevelDebug)

	stdout, stderr, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv())
	if err != nil {
		t.Fatalf("LogRunner.Run() error = %v, want nil", err)
	}
	if string(stdout) != "[]" || string(stderr) != "warn" {
		t.Errorf("LogRunner.Run() = %q, %q, want %q, %q", stdout, stderr, "[]", "warn")
	}

	recs := lg.Records()
	if len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}

	before := recs[0]
	if got := str(t, before, "msg"); got != "restic exec" {
		t.Errorf("first record msg = %q, want %q", got, "restic exec")
	}
	if got := str(t, before, "level"); got != "DEBUG" {
		t.Errorf("first record level = %q, want %q", got, "DEBUG")
	}
	if got, want := str(t, before, "cmd"), CommandLine(logTestArgs(), logTestEnv()); got != want {
		t.Errorf("first record cmd = %q, want %q", got, want)
	}
	if got := str(t, before, "cmd"); !strings.Contains(got, "--password-file ***") {
		t.Errorf("first record cmd = %q, want it to contain %q", got, "--password-file ***")
	}
	if got := str(t, before, "op"); got != "snapshots" {
		t.Errorf("first record op = %q, want %q", got, "snapshots")
	}

	after := recs[1]
	if got := str(t, after, "msg"); got != "restic done" {
		t.Errorf("second record msg = %q, want %q", got, "restic done")
	}
	if got := str(t, after, "level"); got != "DEBUG" {
		t.Errorf("second record level = %q, want %q", got, "DEBUG")
	}
	if got := str(t, after, "op"); got != "snapshots" {
		t.Errorf("second record op = %q, want %q", got, "snapshots")
	}
	dur, ok := after["dur_ms"].(float64)
	if !ok {
		t.Fatalf("second record dur_ms = %v (%T), want a number", after["dur_ms"], after["dur_ms"])
	}
	if dur < 0 {
		t.Errorf("second record dur_ms = %v, want >= 0", dur)
	}
	exit, ok := after["exit"].(float64)
	if !ok {
		t.Fatalf("second record exit = %v (%T), want a number", after["exit"], after["exit"])
	}
	if exit != 0 {
		t.Errorf("second record exit = %v, want 0", exit)
	}
	nStderr, ok := after["stderr_bytes"].(float64)
	if !ok {
		t.Fatalf("second record stderr_bytes = %v (%T), want a number", after["stderr_bytes"], after["stderr_bytes"])
	}
	if nStderr != 4 {
		t.Errorf("second record stderr_bytes = %v, want 4", nStderr)
	}
	if _, ok := after["err"]; ok {
		t.Errorf("second record has err = %v on success, want no err field", after["err"])
	}
}

func TestLogRunnerEmitsNothingAtInfo(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		want  int
	}{
		{"debug", slog.LevelDebug, 2},
		{"info", slog.LevelInfo, 0},
		{"warn", slog.LevelWarn, 0},
		{"error", slog.LevelError, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stdout: []byte("[]")}}}
			lr, lg := newLogRunner(t, fr, tt.level)

			if _, _, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv()); err != nil {
				t.Fatalf("LogRunner.Run() error = %v, want nil", err)
			}
			if recs := lg.Records(); len(recs) != tt.want {
				t.Errorf("LogRunner.Run() at %s emitted %d records, want %d: %v", tt.level, len(recs), tt.want, recs)
			}
		})
	}
}

func TestLogRunnerPassesErrorThrough(t *testing.T) {
	wantErr := errors.New("restic exited with status 1")
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stderr: []byte("boom"), err: wantErr}}}
	lr, lg := newLogRunner(t, fr, slog.LevelDebug)

	_, stderr, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv())
	if !errors.Is(err, wantErr) {
		t.Fatalf("LogRunner.Run() error = %v, want %v", err, wantErr)
	}
	if string(stderr) != "boom" {
		t.Errorf("LogRunner.Run() stderr = %q, want %q", stderr, "boom")
	}

	recs := lg.Records()
	if len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}
	after := recs[1]
	if got := str(t, after, "level"); got != "ERROR" {
		t.Errorf("second record level = %q, want %q", got, "ERROR")
	}
	if got := str(t, after, "err"); !strings.Contains(got, wantErr.Error()) {
		t.Errorf("second record err = %q, want it to contain %q", got, wantErr.Error())
	}
}

func TestLogRunnerNeverLogsSecrets(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {err: errors.New("failed")}}}
	lr, lg := newLogRunner(t, fr, slog.LevelDebug)

	if _, _, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv()); err == nil {
		t.Fatalf("LogRunner.Run() error = nil, want an error")
	}
	if recs := lg.Records(); len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}

	for _, secret := range []string{passwordFilePath, "hunter2-plaintext", "abc123secretkey"} {
		if strings.Contains(lg.Text(), secret) {
			t.Errorf("log output contains secret %q:\n%s", secret, lg.Text())
		}
	}
	if !strings.Contains(lg.Text(), "RESTIC_REPOSITORY=/srv/repo") {
		t.Errorf("log output dropped the non-secret env:\n%s", lg.Text())
	}
}
