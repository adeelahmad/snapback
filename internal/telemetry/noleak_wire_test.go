// Package telemetry_test drives the full client -> otlp exporter -> HTTP wire
// pipeline with hostile inputs and proves that whatever bytes actually reach
// the collector can never carry a path, a URI, a hostname or an identifier,
// and that no request header leaks one either.
package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/errcode"
	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

// wireRequest is one HTTP request the recording server captured.
type wireRequest struct {
	body      []byte
	host      string
	header    http.Header
	userAgent string
}

// wireRecorder is an httptest.Server handler that records every request it
// receives, verbatim, for later inspection.
type wireRecorder struct {
	mu   sync.Mutex
	reqs []wireRequest
}

func (r *wireRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.mu.Lock()
	r.reqs = append(r.reqs, wireRequest{
		body:      body,
		host:      req.Host,
		header:    req.Header.Clone(),
		userAgent: req.Header.Get("User-Agent"),
	})
	r.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

// requests returns a snapshot of every request recorded so far.
func (r *wireRecorder) requests() []wireRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]wireRequest, len(r.reqs))
	copy(out, r.reqs)
	return out
}

// wireEventBuilder builds one schema event from a version string, using a
// fixed, valid value for every other, closed-list argument.
type wireEventBuilder struct {
	name  string
	build func(version string) (telemetry.Event, error)
}

// wireEventBuilders is all five S6-01 event constructors, each reduced to a
// version-only entry point so the hostile-input table below can drive every
// one of them the same way.
func wireEventBuilders(now time.Time) []wireEventBuilder {
	return []wireEventBuilder{
		{"SetupCompleted", func(v string) (telemetry.Event, error) {
			return telemetry.SetupCompleted(v, "ok", time.Minute, now)
		}},
		{"DaemonStarted", func(v string) (telemetry.Event, error) {
			return telemetry.DaemonStarted(v, now)
		}},
		{"MountReady", func(v string) (telemetry.Event, error) {
			return telemetry.MountReady(v, time.Minute, now)
		}},
		{"DoctorFailed", func(v string) (telemetry.Event, error) {
			return telemetry.DoctorFailed(v, "config", now)
		}},
		{"ErrorEvent", func(v string) (telemetry.Event, error) {
			return telemetry.ErrorEvent(v, errcode.InvalidConfig, now)
		}},
	}
}

// wireAllowedHeaders is the closed set of request headers the OTLP client in
// internal/telemetry/otlp/client.go sets or that Go's net/http transport adds
// on its own. Anything outside this set is unexpected and, on a real
// request, would be a place a future change could smuggle an identifier.
var wireAllowedHeaders = map[string]bool{
	"User-Agent":      true,
	"Content-Type":    true,
	"Content-Length":  true,
	"Accept-Encoding": true,
}

// TestNoleakWireEndToEnd drives a real [telemetry.Client] wired to a real
// [otlp.Client] exporter, pointed at a recording httptest.Server, through 200
// emits of all five schema events built from the hostile-input table. It then
// inspects every byte and header the server actually received.
//
// Most hostile versions are rejected before an Event is ever constructed
// (S6-04/T2's NewVersionAttr guard and S6-04/T1's rune checks), so those
// never reach Emit at all -- that is itself part of the guarantee this test
// pins: nothing hostile gets far enough to be queued for the wire. A small
// number of guaranteed-valid control emits (one per event type) ensure the
// server actually receives traffic, so the header and Host assertions below
// are not vacuous.
func TestNoleakWireEndToEnd(t *testing.T) {
	rec := &wireRecorder{}
	srv := httptest.NewServer(rec)
	defer srv.Close()

	ep, err := otlp.ParseEndpoint(srv.URL)
	if err != nil {
		t.Fatalf("ParseEndpoint(%q) = %v, want no error", srv.URL, err)
	}
	const clientVersion = "9.9.9"
	resource := otlp.Resource{
		Version:   clientVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		InstallID: strings.Repeat("f", 32),
	}
	exporter := otlp.NewClient(ep, resource, otlp.ClientOptions{
		Version: clientVersion,
		Timeout: 2 * time.Second,
	})

	client := telemetry.New(telemetry.Options{Enabled: true, Endpoint: srv.URL, Exporter: exporter})

	now := time.Now()
	builders := wireEventBuilders(now)
	ctx := context.Background()

	// Guarantee real wire traffic: one valid-version control emit per event
	// type, before the hostile-input barrage below.
	for _, b := range builders {
		ev, err := b.build(clientVersion)
		if err != nil {
			t.Fatalf("%s(%q, ...) = _, %v, want a valid control event", b.name, clientVersion, err)
		}
		client.Emit(ctx, ev)
	}

	hostiles := hostileInputs(t)
	for i := 0; i < 200; i++ {
		b := builders[i%len(builders)]
		h := hostiles[i%len(hostiles)]
		ev, err := b.build(h.value)
		if err != nil {
			// Rejected before it could ever be queued for the wire: the
			// property this test exists to pin.
			continue
		}
		client.Emit(ctx, ev)
	}

	if err := client.Close(ctx); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}

	reqs := rec.requests()
	if len(reqs) == 0 {
		t.Fatal("recording server saw zero requests, want at least the control emits")
	}

	epURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) = %v, want no error", srv.URL, err)
	}
	wantUserAgent := "snapback/" + clientVersion

	for i, req := range reqs {
		if findings := telemetry.ScanProhibited(req.body); len(findings) != 0 {
			t.Errorf("request %d body %s: ScanProhibited() = %+v, want zero findings", i, req.body, findings)
		}

		// Header check: Host and User-Agent must carry nothing but the
		// operator-configured endpoint and client version -- never a
		// machine hostname or any other identifier.
		if req.host != epURL.Host {
			t.Errorf("request %d Host = %q, want %q (the configured endpoint only)", i, req.host, epURL.Host)
		}
		if req.userAgent != wantUserAgent {
			t.Errorf("request %d User-Agent = %q, want %q", i, req.userAgent, wantUserAgent)
		}
		for key := range req.header {
			if !wireAllowedHeaders[key] {
				t.Errorf("request %d carries unexpected header %q: %v, want only %v", i, key, req.header[key], wireAllowedHeaders)
			}
		}
	}

	// TLS SNI: httptest.NewServer speaks plain HTTP over a loopback address
	// (127.0.0.1), so no TLS handshake happens here and there is no SNI to
	// read off the wire. otlp.ParseEndpoint (internal/telemetry/otlp/endpoint.go)
	// requires https for any non-loopback host, and the only host name it
	// ever uses -- for both the request URL and, in a real https deployment,
	// the resulting SNI -- is the operator-configured Endpoint string itself;
	// nothing derived from os.Hostname(), the install id or any event
	// attribute ever reaches ParseEndpoint, NewClient or Encode. The Host
	// assertion above is that guarantee's testable form with this transport.
}
