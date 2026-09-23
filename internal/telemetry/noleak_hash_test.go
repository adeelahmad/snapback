// Package telemetry_test proves that hashing a hostile value does not make it
// safe to emit: the privacy page promises nothing hashed or in the clear, so
// even a SHA-256, SHA-1, MD5, FNV-1a or base64 encoding of a leaking value
// must never appear in an event's encoded payload. This guards against a
// future code path that tries to "anonymize" a value by hashing it instead of
// rejecting it, which [TestNoleakEvents] cannot catch because it only checks
// for the raw hostile value.
package telemetry_test

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"strings"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// hostileHashes returns every hash and encoding of value that a future code
// path might mistakenly emit in place of the raw value, believing a hash
// "anonymizes" it.
func hostileHashes(value string) []string {
	sha256Sum := sha256.Sum256([]byte(value))
	sha1Sum := sha1.Sum([]byte(value))
	md5Sum := md5.Sum([]byte(value))
	fnvHash := fnv.New32a()
	fnvHash.Write([]byte(value))
	return []string{
		hex.EncodeToString(sha256Sum[:]),
		hex.EncodeToString(sha1Sum[:]),
		hex.EncodeToString(md5Sum[:]),
		hex.EncodeToString(fnvHash.Sum(nil)),
		base64.StdEncoding.EncodeToString([]byte(value)),
	}
}

// assertNoHashOfHostileValue fails the test if the JSON encoding of ev.Attrs
// contains any hash or encoding of hostileValue. When err is non-nil, the
// constructor rejected hostileValue outright, so nothing was produced that
// could carry a hash of it.
func assertNoHashOfHostileValue(t *testing.T, ev telemetry.Event, err error, hostileValue string) {
	t.Helper()
	if err != nil {
		return
	}
	b, marshalErr := json.Marshal(ev.Attrs)
	if marshalErr != nil {
		t.Fatalf("json.Marshal(%+v) = %v, want no error", ev.Attrs, marshalErr)
	}
	payload := string(b)
	for _, h := range hostileHashes(hostileValue) {
		if strings.Contains(payload, h) {
			t.Fatalf("payload %s contains a hash/encoding of hostile value %q: %q", payload, hostileValue, h)
		}
	}
}

// TestNoleakHash drives every S6-01 event constructor over the hostile input
// table from [hostileInputs] and asserts that for each (constructor, hostile
// input) pair, the produced event's encoded payload contains no hash or
// encoding of that hostile input.
func TestNoleakHash(t *testing.T) {
	now := time.Now()

	t.Run("SetupCompleted_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.SetupCompleted(h.value, "ok", time.Minute, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	t.Run("DaemonStarted_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DaemonStarted(h.value, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	t.Run("MountReady_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.MountReady(h.value, time.Minute, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	// DoctorFailed takes (version, check string, now). Hostile inputs drive
	// version with a valid check, and separately drive check with a valid
	// version, mirroring TestNoleakEvents.
	t.Run("DoctorFailed_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DoctorFailed(h.value, "config", now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	t.Run("DoctorFailed_check", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.DoctorFailed("1.0.0", h.value, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	// ErrorEvent takes (version string, code errcode.Code, now). Hostile
	// inputs drive version with a valid code, and separately are cast to
	// errcode.Code and driven through code, mirroring TestNoleakEvents.
	t.Run("ErrorEvent_version", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				ev, err := telemetry.ErrorEvent(h.value, errcode.InvalidConfig, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})

	t.Run("ErrorEvent_code", func(t *testing.T) {
		for _, h := range hostileInputs(t) {
			t.Run(h.name, func(t *testing.T) {
				code := errcode.Code(h.value)
				ev, err := telemetry.ErrorEvent("1.0.0", code, now)
				assertNoHashOfHostileValue(t, ev, err, h.value)
			})
		}
	})
}
