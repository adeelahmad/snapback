package restic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

// logRecord is one decoded JSON record from the test handler.
type logRecord map[string]any

// newLogRunner returns a LogRunner over fr at level, plus the buffer the
// records land in.
func newLogRunner(fr *fakeRunner, level slog.Level) (LogRunner, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level}))
	return LogRunner{Log: log, Runner: fr}, buf
}

// decodeRecords decodes every JSON line the handler wrote.
func decodeRecords(t *testing.T, buf *bytes.Buffer) []logRecord {
	t.Helper()
	var out []logRecord
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var rec logRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// str returns the record's field as a string, failing when it is absent.
func (r logRecord) str(t *testing.T, key string) string {
	t.Helper()
	v, ok := r[key]
	if !ok {
		t.Fatalf("record %v has no %q field", map[string]any(r), key)
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
	lr, buf := newLogRunner(fr, slog.LevelDebug)

	stdout, stderr, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv())
	if err != nil {
		t.Fatalf("LogRunner.Run() error = %v, want nil", err)
	}
	if string(stdout) != "[]" || string(stderr) != "warn" {
		t.Errorf("LogRunner.Run() = %q, %q, want %q, %q", stdout, stderr, "[]", "warn")
	}

	recs := decodeRecords(t, buf)
	if len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}

	before := recs[0]
	if got := before.str(t, "msg"); got != "restic exec" {
		t.Errorf("first record msg = %q, want %q", got, "restic exec")
	}
	if got := before.str(t, "level"); got != "DEBUG" {
		t.Errorf("first record level = %q, want %q", got, "DEBUG")
	}
	if got, want := before.str(t, "cmd"), CommandLine(logTestArgs(), logTestEnv()); got != want {
		t.Errorf("first record cmd = %q, want %q", got, want)
	}
	if got := before.str(t, "cmd"); !strings.Contains(got, "--password-file ***") {
		t.Errorf("first record cmd = %q, want it to contain %q", got, "--password-file ***")
	}
	if got := before.str(t, "op"); got != "snapshots" {
		t.Errorf("first record op = %q, want %q", got, "snapshots")
	}

	after := recs[1]
	if got := after.str(t, "msg"); got != "restic done" {
		t.Errorf("second record msg = %q, want %q", got, "restic done")
	}
	if got := after.str(t, "level"); got != "DEBUG" {
		t.Errorf("second record level = %q, want %q", got, "DEBUG")
	}
	if got := after.str(t, "op"); got != "snapshots" {
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
			lr, buf := newLogRunner(fr, tt.level)

			if _, _, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv()); err != nil {
				t.Fatalf("LogRunner.Run() error = %v, want nil", err)
			}
			if recs := decodeRecords(t, buf); len(recs) != tt.want {
				t.Errorf("LogRunner.Run() at %s emitted %d records, want %d: %v", tt.level, len(recs), tt.want, recs)
			}
		})
	}
}

func TestLogRunnerPassesErrorThrough(t *testing.T) {
	wantErr := errors.New("restic exited with status 1")
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {stderr: []byte("boom"), err: wantErr}}}
	lr, buf := newLogRunner(fr, slog.LevelDebug)

	_, stderr, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv())
	if !errors.Is(err, wantErr) {
		t.Fatalf("LogRunner.Run() error = %v, want %v", err, wantErr)
	}
	if string(stderr) != "boom" {
		t.Errorf("LogRunner.Run() stderr = %q, want %q", stderr, "boom")
	}

	recs := decodeRecords(t, buf)
	if len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}
	after := recs[1]
	if got := after.str(t, "level"); got != "ERROR" {
		t.Errorf("second record level = %q, want %q", got, "ERROR")
	}
	if got := after.str(t, "err"); !strings.Contains(got, wantErr.Error()) {
		t.Errorf("second record err = %q, want it to contain %q", got, wantErr.Error())
	}
}

func TestLogRunnerNeverLogsSecrets(t *testing.T) {
	fr := &fakeRunner{results: map[string]fakeResult{"snapshots": {err: errors.New("failed")}}}
	lr, buf := newLogRunner(fr, slog.LevelDebug)

	if _, _, err := lr.Run(context.Background(), "restic", logTestArgs(), logTestEnv()); err == nil {
		t.Fatalf("LogRunner.Run() error = nil, want an error")
	}
	if recs := decodeRecords(t, buf); len(recs) != 2 {
		t.Fatalf("LogRunner.Run() emitted %d records, want 2: %v", len(recs), recs)
	}

	for _, secret := range []string{passwordFilePath, "hunter2-plaintext", "abc123secretkey"} {
		if strings.Contains(buf.String(), secret) {
			t.Errorf("log output contains secret %q:\n%s", secret, buf.String())
		}
	}
	if !strings.Contains(buf.String(), "RESTIC_REPOSITORY=/srv/repo") {
		t.Errorf("log output dropped the non-secret env:\n%s", buf.String())
	}
}
