package crash

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// wantEventID and wantSentAt are the fixed, injected identity fields the
// golden tests use so Envelope's output is fully deterministic. Production
// callers get these from NewEventID and time.Now, but Envelope itself takes
// them as plain fields precisely so it stays a pure, golden-testable encoder.
const (
	wantEventID = "0123456789abcdef0123456789abcdef"
	wantSentAt  = "2024-01-15T10:30:00Z"
)

var fixedSentAt = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

// TestEnvelope_GoldenFormat pins the exact newline-delimited envelope bytes
// for one deterministic event: an envelope header with event_id/sent_at, an
// item header naming the event body's byte length, and the event body
// itself carrying platform/release/level/frames and the panic's type name.
func TestEnvelope_GoldenFormat(t *testing.T) {
	ev := PanicEvent{
		Panic: errors.New("boom"),
		Frames: []Frame{
			{Module: "github.com/adeelahmad/snapback/internal/mount", Function: "(*Catalog).Refresh"},
			{Module: "github.com/adeelahmad/snapback/internal/resolver", Function: "Resolve"},
		},
		Version: "0.1.0",
		EventID: wantEventID,
		SentAt:  fixedSentAt,
	}

	wantBody := `{"platform":"go","release":"snapback@0.1.0","level":"fatal","exception":{"values":[{"type":"*errors.errorString","stacktrace":{"frames":[{"module":"github.com/adeelahmad/snapback/internal/mount","function":"(*Catalog).Refresh"},{"module":"github.com/adeelahmad/snapback/internal/resolver","function":"Resolve"}]}}]}}`
	wantHeader := `{"event_id":"` + wantEventID + `","sent_at":"` + wantSentAt + `"}`
	wantItemHeader := fmt.Sprintf(`{"type":"event","length":%d}`, len(wantBody))
	want := wantHeader + "\n" + wantItemHeader + "\n" + wantBody + "\n"

	got := Envelope(ev)

	if string(got) != want {
		t.Fatalf("Envelope() =\n%q\nwant\n%q", got, want)
	}
}

// TestEnvelope_EmptyFramesGoldenFormat pins the same shape with zero frames,
// so GREEN cannot special-case "at least one frame" to pass the first test.
func TestEnvelope_EmptyFramesGoldenFormat(t *testing.T) {
	ev := PanicEvent{
		Panic:   errors.New("boom"),
		Frames:  nil,
		Version: "0.1.0",
		EventID: wantEventID,
		SentAt:  fixedSentAt,
	}

	wantBody := `{"platform":"go","release":"snapback@0.1.0","level":"fatal","exception":{"values":[{"type":"*errors.errorString","stacktrace":{"frames":[]}}]}}`
	wantHeader := `{"event_id":"` + wantEventID + `","sent_at":"` + wantSentAt + `"}`
	wantItemHeader := fmt.Sprintf(`{"type":"event","length":%d}`, len(wantBody))
	want := wantHeader + "\n" + wantItemHeader + "\n" + wantBody + "\n"

	got := Envelope(ev)

	if string(got) != want {
		t.Fatalf("Envelope() =\n%q\nwant\n%q", got, want)
	}
}

// TestEnvelope_NeverLeaksPanicMessage proves the hard privacy requirement:
// the panic's message text (and anything it interpolates) never appears in
// the encoded envelope, only the panic's reflect type name.
func TestEnvelope_NeverLeaksPanicMessage(t *testing.T) {
	const secretPath = "/Users/adeel/backups/production-db-credentials.txt"
	ev := PanicEvent{
		Panic:   fmt.Errorf("failed to open %s: permission denied", secretPath),
		Frames:  []Frame{{Module: "github.com/adeelahmad/snapback/internal/mount", Function: "(*Catalog).Refresh"}},
		Version: "0.1.0",
		EventID: wantEventID,
		SentAt:  fixedSentAt,
	}

	got := string(Envelope(ev))

	if strings.Contains(got, secretPath) {
		t.Errorf("Envelope() leaked the panic message's path: %s", got)
	}
	if strings.Contains(got, "permission denied") {
		t.Errorf("Envelope() leaked the panic message text: %s", got)
	}
	if !strings.Contains(got, `"type":"*errors.errorString"`) {
		t.Errorf("Envelope() = %s, want it to carry the panic type *errors.errorString", got)
	}
}

// TestNewEventID_FormatAndUniqueness pins NewEventID's shape: 32 lowercase
// hex characters, and two calls never collide.
func TestNewEventID_FormatAndUniqueness(t *testing.T) {
	id1, err := NewEventID()
	if err != nil {
		t.Fatalf("NewEventID() error = %v", err)
	}
	if len(id1) != 32 {
		t.Errorf("NewEventID() = %q (len %d), want 32 hex characters", id1, len(id1))
	}
	for _, r := range id1 {
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Errorf("NewEventID() = %q, want lowercase hex only", id1)
			break
		}
	}

	id2, err := NewEventID()
	if err != nil {
		t.Fatalf("NewEventID() error = %v", err)
	}
	if id1 == id2 {
		t.Errorf("NewEventID() returned the same id twice: %q", id1)
	}
}
