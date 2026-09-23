package crash

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

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
func Envelope(ev PanicEvent) []byte {
	body := eventBody(ev)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `{"event_id":%s,"sent_at":%s}`+"\n", jsonString(ev.EventID), jsonString(ev.SentAt.UTC().Format(time.RFC3339Nano)))
	fmt.Fprintf(&buf, `{"type":"event","length":%d}`+"\n", len(body))
	buf.Write(body)
	buf.WriteByte('\n')
	return buf.Bytes()
}

// eventBody builds the event body JSON: platform, release, level and the
// panic's reflect type name and frames. It never encodes ev.Panic's message
// or anything it wraps.
func eventBody(ev PanicEvent) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `{"platform":"go","release":%s,"level":"fatal","exception":{"values":[{"type":%s,"stacktrace":{"frames":[`,
		jsonString("snapback@"+ev.Version), jsonString(reflect.TypeOf(ev.Panic).String()))
	for i, frame := range ev.Frames {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, `{"module":%s,"function":%s}`, jsonString(frame.Module), jsonString(frame.Function))
	}
	buf.WriteString(`]}}]}}`)
	return buf.Bytes()
}

// jsonString returns s encoded as a JSON string literal, quotes included.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// NewEventID returns a fresh 32 lowercase hex character identifier for one
// crash report envelope, drawn from crypto/rand (the same shape as
// telemetry.InstallID's identifiers).
func NewEventID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("crash: generate event id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
