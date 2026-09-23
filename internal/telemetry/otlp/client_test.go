package otlp_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

func mustClientEndpoint(t *testing.T, rawURL string) otlp.Endpoint {
	t.Helper()
	ep, err := otlp.ParseEndpoint(rawURL)
	if err != nil {
		t.Fatalf("ParseEndpoint(%q): %v", rawURL, err)
	}
	return ep
}

// TestClient_Export_SuccessStatuses pins S6-03/T3: a 200 and a 202 are both
// success, with exactly one request sent.
func TestClient_Export_SuccessStatuses(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var reqCount int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&reqCount, 1)
				w.WriteHeader(status)
			}))
			defer srv.Close()

			c := otlp.NewClient(mustClientEndpoint(t, srv.URL), testResource(), otlp.ClientOptions{})
			if err := c.Export(context.Background(), testEvents()); err != nil {
				t.Fatalf("Export() with status %d: unexpected error: %v", status, err)
			}
			if got := atomic.LoadInt32(&reqCount); got != 1 {
				t.Fatalf("request count = %d, want 1", got)
			}
		})
	}
}

// TestClient_Export_4xxNotRetried pins S6-03/T3: a 4xx is a terminal error
// with exactly one request sent, no retry.
func TestClient_Export_4xxNotRetried(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := otlp.NewClient(mustClientEndpoint(t, srv.URL), testResource(), otlp.ClientOptions{Retries: 2, Backoff: 10 * time.Millisecond})
	if err := c.Export(context.Background(), testEvents()); err == nil {
		t.Fatal("Export() with a 4xx response: want error, got nil")
	}
	if got := atomic.LoadInt32(&reqCount); got != 1 {
		t.Fatalf("request count = %d, want 1 (a 4xx must not be retried)", got)
	}
}

// TestClient_Export_5xxRetriedThenFails pins S6-03/T3: a 5xx is retried at
// most Retries times, spaced by Backoff, then reported as an error.
func TestClient_Export_5xxRetriedThenFails(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	const retries = 2
	c := otlp.NewClient(mustClientEndpoint(t, srv.URL), testResource(), otlp.ClientOptions{Retries: retries, Backoff: 10 * time.Millisecond})
	if err := c.Export(context.Background(), testEvents()); err == nil {
		t.Fatal("Export() with a persistent 5xx response: want error, got nil")
	}
	if got, want := atomic.LoadInt32(&reqCount), int32(1+retries); got != want {
		t.Fatalf("request count = %d, want %d (1 initial attempt + %d retries)", got, want, retries)
	}
}

// TestClient_Export_TimeoutAbandonsHungRequest pins S6-03/T3: a hung
// handler is abandoned at Timeout rather than blocking forever.
func TestClient_Export_TimeoutAbandonsHungRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := otlp.NewClient(mustClientEndpoint(t, srv.URL), testResource(), otlp.ClientOptions{Timeout: 100 * time.Millisecond})
	start := time.Now()
	err := c.Export(context.Background(), testEvents())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Export() against a hung handler: want error, got nil")
	}
	if elapsed > time.Second {
		t.Fatalf("Export() took %s, want it abandoned near the 100ms timeout", elapsed)
	}
}

// TestClient_Export_DoesNotFollowRedirect pins S6-03/T3: a 302 is not
// followed; the handler is hit exactly once and Export reports an error.
func TestClient_Export_DoesNotFollowRedirect(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	defer srv.Close()

	c := otlp.NewClient(mustClientEndpoint(t, srv.URL), testResource(), otlp.ClientOptions{})
	if err := c.Export(context.Background(), testEvents()); err == nil {
		t.Fatal("Export() against a 302 response: want error, got nil")
	}
	if got := atomic.LoadInt32(&reqCount); got != 1 {
		t.Fatalf("request count = %d, want 1 (a redirect must not be followed)", got)
	}
}

// TestClient_Export_RequestShape pins S6-03/T3: Export POSTs the T1 body to
// /v1/metrics with Content-Type: application/json and a User-Agent of
// exactly "snapback/<Version>", and never sends back a Cookie header even
// after the server has set one.
func TestClient_Export_RequestShape(t *testing.T) {
	type captured struct {
		body        []byte
		contentType string
		userAgent   string
		cookie      string
		method      string
		path        string
	}
	var reqs []captured
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		reqs = append(reqs, captured{
			body:        body,
			contentType: r.Header.Get("Content-Type"),
			userAgent:   r.Header.Get("User-Agent"),
			cookie:      r.Header.Get("Cookie"),
			method:      r.Method,
			path:        r.URL.Path,
		})
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "leak"})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	const version = "9.9.9"
	res := testResource()
	events := testEvents()
	wantBody, err := otlp.Encode(events, res)
	if err != nil {
		t.Fatalf("Encode(): unexpected error: %v", err)
	}

	c := otlp.NewClient(mustClientEndpoint(t, srv.URL), res, otlp.ClientOptions{Version: version})
	if err := c.Export(context.Background(), events); err != nil {
		t.Fatalf("first Export(): unexpected error: %v", err)
	}
	if err := c.Export(context.Background(), events); err != nil {
		t.Fatalf("second Export(): unexpected error: %v", err)
	}

	if len(reqs) != 2 {
		t.Fatalf("server received %d requests, want 2", len(reqs))
	}

	first := reqs[0]
	if first.method != http.MethodPost {
		t.Errorf("method = %q, want %q", first.method, http.MethodPost)
	}
	if first.path != "/v1/metrics" {
		t.Errorf("path = %q, want /v1/metrics", first.path)
	}
	if first.contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", first.contentType)
	}
	wantUA := "snapback/" + version
	if first.userAgent != wantUA {
		t.Errorf("User-Agent = %q, want %q", first.userAgent, wantUA)
	}
	if string(first.body) != string(wantBody) {
		t.Errorf("body = %s, want %s", first.body, wantBody)
	}

	if second := reqs[1]; second.cookie != "" {
		t.Errorf("Cookie header on second request = %q, want none", second.cookie)
	}
}
