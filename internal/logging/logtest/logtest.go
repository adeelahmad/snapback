// Package logtest captures the slog records code under test writes, so every
// package's logging tests decode them the same way.
package logtest

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// Record is one decoded record with its attributes in the order the handler
// wrote them. The level, message and timestamp are not attributes: Msg holds
// the message and the other two are dropped.
type Record struct {
	Msg   string
	Keys  []string
	Attrs map[string]any
}

// Log holds everything a captured logger has written.
type Log struct {
	t   testing.TB
	buf *bytes.Buffer
}

// Capture returns a JSON logger writing records at or above level, and the Log
// recording them.
func Capture(t testing.TB, level slog.Level) (*slog.Logger, *Log) {
	t.Helper()
	buf := &bytes.Buffer{}
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})
	return slog.New(h), &Log{t: t, buf: buf}
}

// Text returns the raw log output written so far.
func (l *Log) Text() string { return l.buf.String() }

// Records decodes every record written so far, numbers as float64.
func (l *Log) Records() []map[string]any {
	l.t.Helper()
	var out []map[string]any
	for _, line := range l.lines() {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			l.t.Fatalf("decode log line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// Ordered decodes every record written so far keeping attribute order, with
// numbers as json.Number.
func (l *Log) Ordered() []Record {
	l.t.Helper()
	var out []Record
	for _, line := range l.lines() {
		out = append(out, l.decodeOrdered(line))
	}
	return out
}

func (l *Log) decodeOrdered(line string) Record {
	l.t.Helper()
	dec := json.NewDecoder(strings.NewReader(line))
	dec.UseNumber()
	if _, err := dec.Token(); err != nil {
		l.t.Fatalf("decode %q: %v", line, err)
	}
	rec := Record{Attrs: map[string]any{}}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			l.t.Fatalf("decode key in %q: %v", line, err)
		}
		key, ok := tok.(string)
		if !ok {
			l.t.Fatalf("decode %q: key %v is not a string", line, tok)
		}
		var val any
		if err := dec.Decode(&val); err != nil {
			l.t.Fatalf("decode value for %q in %q: %v", key, line, err)
		}
		switch key {
		case slog.TimeKey, slog.LevelKey:
		case slog.MessageKey:
			rec.Msg, _ = val.(string)
		default:
			rec.Keys = append(rec.Keys, key)
			rec.Attrs[key] = val
		}
	}
	return rec
}

func (l *Log) lines() []string {
	text := strings.TrimSpace(l.buf.String())
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
