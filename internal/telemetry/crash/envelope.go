package crash

import "time"

// PanicEvent is everything Envelope needs to build one crash report. The
// recovered panic value is carried as-is, but Envelope encodes only its
// reflect type name — the panic message text, and anything it wraps, is
// never retained or transmitted (same privacy boundary as S6-04's
// leak-prevention work).
type PanicEvent struct {
	// Panic is the value recover() returned. Envelope encodes only
	// reflect.TypeOf(Panic).String(), e.g. "*errors.errorString" — never
	// Panic's message or any value it wraps.
	Panic any
	// Frames are the snapback-only stack frames from Frames(stack).
	Frames []Frame
	// Version is the snapback release version, e.g. "0.1.0". Envelope
	// formats the event's release as "snapback@<Version>".
	Version string
	// EventID is 32 lowercase hex characters identifying this envelope.
	EventID string
	// SentAt is when the envelope was built.
	SentAt time.Time
}

// Envelope encodes ev as a newline-delimited Sentry envelope of the shape
// GlitchTip's envelope endpoint accepts:
//
//	{"event_id":"<32 lowercase hex>","sent_at":"<RFC3339>"}
//	{"type":"event","length":<N>}
//	<N bytes of event body JSON>
//
// The event body carries platform "go", release "snapback@<Version>", level
// "fatal", ev.Frames, and ev.Panic's type name only — never its message.
//
// SUB-AGENT-TODO(S6-08/T2 GREEN): implement the encoding above. This shim
// returns nil so callers compile and every test fails by assertion.
func Envelope(ev PanicEvent) []byte {
	return nil
}

// NewEventID returns a fresh 32 lowercase hex character identifier for one
// crash report envelope, drawn from crypto/rand (the same shape as
// telemetry.InstallID's identifiers).
//
// SUB-AGENT-TODO(S6-08/T2 GREEN): implement the generation above. This shim
// returns a zero value so callers compile and every test fails by assertion.
func NewEventID() (string, error) {
	return "", nil
}
