// Package telemetry_test drives every S6-01 event constructor with hostile
// inputs and proves neither the constructor nor its output can leak an
// identifier: it either rejects the input outright, or the JSON encoding of
// the event's Attrs it returns scans clean under [telemetry.ScanProhibited].
package telemetry_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// hostileInput is one adversarial value used in place of a version, check or
// code string across every constructor in this table.
type hostileInput struct {
	name  string
	value string
}

// hostileInputs is the fixed table of hostile values every constructor faces:
// a home path, a remote URI, the real machine hostname, a fake snapshot id, a
// plausible username and a CLI flag carrying a path.
func hostileInputs(t *testing.T) []hostileInput {
	t.Helper()
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname() = %v, want a hostname", err)
	}
	return []hostileInput{
		{"home_path", "/home/u/project"},
		{"uri", "sftp://backup.lan/repo"},
		{"real_hostname", hostname},
		{"snapshot_id", strings.Repeat("a1b2c3d4", 8)}, // 64 hex chars
		{"username", "j.doe"},
		{"password_flag", "--password-file /x"},
	}
}

// assertRejectedOrClean fails the test unless err is non-nil (the constructor
// rejected the hostile input) or ev.Attrs's JSON encoding scans clean.
func assertRejectedOrClean(t *testing.T, ev telemetry.Event, err error) {
	t.Helper()
	if err != nil {
		return
	}
	b, marshalErr := json.Marshal(ev.Attrs)
	if marshalErr != nil {
		t.Fatalf("json.Marshal(%+v) = %v, want no error", ev.Attrs, marshalErr)
	}
	if findings := telemetry.ScanProhibited(b); len(findings) != 0 {
		t.Fatalf("ScanProhibited(%s) = %+v, want zero findings", b, findings)
	}
}

// assertRejected fails the test unless err is non-nil: used where the hostile
// input is fed to a closed-list parameter (check, code) that must reject
// anything outside its fixed set.
func assertRejected(t *testing.T, ev telemetry.Event, err error, ctor, arg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s(%q) = %+v, nil, want an error rejecting the value", ctor, arg, ev)
	}
	assertRejectedOrClean(t, ev, err)
}

// TestNoleakEvents drives every S6-01 event constructor over the hostile
// input table and asserts that for each (constructor, hostile input) pair,
// the constructor either rejects the input or produces attrs that scan clean.
func TestNoleakEvents(t *testing.T) {
	now := time.Now()

	t.Run("SetupCompleted_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.SetupCompleted(h.value, "ok", time.Minute, now)
				assertRejectedOrClean(t, ev, err)
			})
		}
	})

	t.Run("DaemonStarted_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DaemonStarted(h.value, now)
				assertRejectedOrClean(t, ev, err)
			})
		}
	})

	t.Run("MountReady_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.MountReady(h.value, time.Minute, now)
				assertRejectedOrClean(t, ev, err)
			})
		}
	})

	// DoctorFailed takes (version, check string, now). Hostile inputs drive
	// version with a valid check, and separately drive check with a valid
	// version: check is a closed list, so every hostile value there must be
	// rejected outright.
	t.Run("DoctorFailed_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DoctorFailed(h.value, "config", now)
				assertRejectedOrClean(t, ev, err)
			})
		}
	})

	t.Run("DoctorFailed_check", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DoctorFailed("1.0.0", h.value, now)
				assertRejected(t, ev, err, "DoctorFailed check", h.value)
			})
		}
	})

	// ErrorEvent takes (version string, code errcode.Code, now). Hostile
	// inputs drive version with a valid code, and separately are cast to
	// errcode.Code and driven through code: code is a closed list, so every
	// hostile value there must be rejected outright.
	t.Run("ErrorEvent_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.ErrorEvent(h.value, errcode.InvalidConfig, now)
				assertRejectedOrClean(t, ev, err)
			})
		}
	})

	t.Run("ErrorEvent_code", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				code := errcode.Code(h.value)
				ev, err := telemetry.ErrorEvent("1.0.0", code, now)
				assertRejected(t, ev, err, "ErrorEvent code", string(code))
			})
		}
	})
}
