package crash

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testVersion is the snapback release version used across this file's
// X-Sentry-Auth assertions.
const testVersion = "0.1.0"

// callReport invokes Report and turns its compile-shim panic into a
// t.Fatal, so a not-yet-implemented Report fails only the calling test
// instead of crashing the whole test binary and starving every later test
// in this package (e.g. the envelope and frames tests) of a chance to run.
func callReport(t *testing.T, ctx context.Context, opts Options, envelope []byte) (err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Report(...) panicked (not implemented yet): %v", r)
		}
	}()
	return Report(ctx, opts, envelope)
}

// dsnFor builds a Sentry/GlitchTip DSN-style envelope URL pointing at srv,
// carrying key as the URL's userinfo component: "http://<key>@host/api/1/envelope/".
func dsnFor(srv *httptest.Server, key string) string {
	return strings.Replace(srv.URL, "://", "://"+key+"@", 1) + "/api/1/envelope/"
}

// TestReport_PostsEnvelopeWithSentryAuthHeader pins the wire contract: a
// single POST of the exact envelope bytes to the userinfo-stripped path,
// carrying the standard Sentry envelope auth header derived from the
// endpoint's userinfo key and the configured version.
func TestReport_PostsEnvelopeWithSentryAuthHeader(t *testing.T) {
	var (
		calls   int
		gotAuth string
		gotPath string
		gotBody []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		gotAuth = r.Header.Get("X-Sentry-Auth")
		gotPath = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("server: read body: %v", err)
		}
		gotBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	envelope := []byte(`{"event_id":"abc"}` + "\n" + `{"type":"event","length":2}` + "\n" + `{}` + "\n")
	opts := Options{CrashReports: true, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	if err := callReport(t, context.Background(), opts, envelope); err != nil {
		t.Fatalf("Report(...) error = %v, want nil", err)
	}

	if calls != 1 {
		t.Fatalf("server received %d requests, want exactly 1", calls)
	}
	if gotPath != "/api/1/envelope/" {
		t.Errorf("request path = %q, want %q", gotPath, "/api/1/envelope/")
	}
	if !bytes.Equal(gotBody, envelope) {
		t.Errorf("request body = %q, want %q", gotBody, envelope)
	}
	wantAuth := "Sentry sentry_version=7, sentry_client=snapback/" + testVersion + ", sentry_key=testkey"
	if gotAuth != wantAuth {
		t.Errorf("X-Sentry-Auth = %q, want %q", gotAuth, wantAuth)
	}
}

// TestReport_CrashReportsDisabledSendsNothing pins D3: crash_reports=false
// sends nothing even when telemetry.enabled is true.
func TestReport_CrashReportsDisabledSendsNothing(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := Options{
		TelemetryEnabled: true,
		CrashReports:     false,
		Endpoint:         dsnFor(srv, "testkey"),
		Version:          testVersion,
	}

	if err := callReport(t, context.Background(), opts, []byte(`{}`)); err != nil {
		t.Fatalf("Report(...) error = %v, want nil", err)
	}
	if calls != 0 {
		t.Fatalf("server received %d requests, want 0 (crash_reports is false)", calls)
	}
}

// TestReport_EmptyEndpointSendsNothing pins D4: crash_reports=true with an
// empty crash_endpoint is inert, not an error.
func TestReport_EmptyEndpointSendsNothing(t *testing.T) {
	opts := Options{CrashReports: true, Endpoint: "", Version: testVersion}

	if err := callReport(t, context.Background(), opts, []byte(`{}`)); err != nil {
		t.Fatalf("Report(...) error = %v, want nil", err)
	}
}

// TestReport_4xxIsNotRetried pins that a terminal client error makes exactly
// one attempt: no retry loop for a 4xx response.
func TestReport_4xxIsNotRetried(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	opts := Options{CrashReports: true, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	err := callReport(t, context.Background(), opts, []byte(`{}`))
	if err == nil {
		t.Fatalf("Report(...) error = nil, want a non-nil error for a 400 response")
	}
	if calls != 1 {
		t.Fatalf("server received %d requests, want exactly 1 (a 4xx must not be retried)", calls)
	}
}

// TestReport_BoundedAtFiveSeconds pins D7: the call never waits on a hung
// server past its 5s bound.
func TestReport_BoundedAtFiveSeconds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(7 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := Options{CrashReports: true, Endpoint: dsnFor(srv, "testkey"), Version: testVersion}

	start := time.Now()
	_ = callReport(t, context.Background(), opts, []byte(`{}`))
	elapsed := time.Since(start)

	if elapsed > 6*time.Second {
		t.Fatalf("Report(...) took %s against a 7s-hung server, want it bounded at ~5s", elapsed)
	}
}

// TestReport_NeverPanicsOnNetworkFailure pins that Report reports network
// failures as errors, never as a panic, regardless of the failure mode.
func TestReport_NeverPanicsOnNetworkFailure(t *testing.T) {
	t.Run("ConnectionClosed", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		dsn := dsnFor(srv, "testkey")
		srv.Close() // the endpoint now refuses every connection

		opts := Options{CrashReports: true, Endpoint: dsn, Version: testVersion}

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Report(...) panicked on a closed connection (not implemented yet): %v", r)
				}
			}()
			return Report(context.Background(), opts, []byte(`{}`))
		}()
		if err == nil {
			t.Fatalf("Report(...) error = nil, want a non-nil error against a closed connection")
		}
	})

	t.Run("UnresolvableHost", func(t *testing.T) {
		opts := Options{
			CrashReports: true,
			Endpoint:     "http://testkey@no-such-host.invalid.snapback-test/api/1/envelope/",
			Version:      testVersion,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Report(...) panicked on an unresolvable host (not implemented yet): %v", r)
				}
			}()
			return Report(ctx, opts, []byte(`{}`))
		}()
		if err == nil {
			t.Fatalf("Report(...) error = nil, want a non-nil error against an unresolvable host")
		}
	})
}
