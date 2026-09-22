package web

import (
	"bytes"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// TestNewLocalhostBindAcceptedWhenPolicyApproves pins the seam between Policy
// and New: Policy treats "localhost" as loopback, so New must too instead of
// re-deciding with its own net.ParseIP check.
func TestNewLocalhostBindAcceptedWhenPolicyApproves(t *testing.T) {
	const listen = "localhost:0"
	policy, err := Policy(listen, false, nil)
	if err != nil {
		t.Fatalf("Policy(%q, false, nil) error = %v, want a loopback approval", listen, err)
	}
	if policy.Warning != "" {
		t.Fatalf("Policy(%q, false, nil).Warning = %q, want empty for a loopback bind", listen, policy.Warning)
	}

	dir := t.TempDir()
	srv, err := New(Options{Listen: listen, StateDir: dir, Token: "tok", Stdout: &bytes.Buffer{}, Policy: policy})
	if srv != nil {
		t.Cleanup(func() { _ = srv.Close() })
	}
	if err != nil {
		t.Fatalf("New(Listen: %q, Policy: %+v) error = %v, want the policy-approved bind to be accepted", listen, policy, err)
	}

	host := hostOf(t, srv)
	addr, _, err := net.SplitHostPort(host)
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error = %v", host, err)
	}
	if ip := net.ParseIP(addr); ip == nil || !ip.IsLoopback() {
		t.Errorf("New(Listen: %q): listener host = %q, want a loopback IP", listen, addr)
	}
	if _, err := os.Stat(filepath.Join(dir, "web.url")); err != nil {
		t.Errorf("New(Listen: %q): web.url stat error = %v, want the URL published", listen, err)
	}
}

// TestNewLocalhostGuardStillRejectsWildcard keeps the loopback guard: an
// unapproved wildcard bind must still be refused once New defers to the policy.
func TestNewLocalhostGuardStillRejectsWildcard(t *testing.T) {
	const listen = "0.0.0.0:0"
	dir := t.TempDir()
	srv, err := New(Options{Listen: listen, StateDir: dir, Token: "tok", Stdout: &bytes.Buffer{}})
	if srv != nil {
		t.Cleanup(func() { _ = srv.Close() })
	}
	if got, want := errcode.Of(err), errcode.InvalidConfig; got != want {
		t.Errorf("errcode.Of(New(Listen: %q, Policy: {})) = %q, want %q", listen, got, want)
	}
	if err == nil || !strings.Contains(err.Error(), "not loopback") {
		t.Errorf("New(Listen: %q, Policy: {}) error = %v, want it to name the loopback refusal", listen, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web.url")); !os.IsNotExist(err) {
		t.Errorf("New(Listen: %q, Policy: {}): web.url stat error = %v, want not exist", listen, err)
	}
}

// TestNewLocalhostWildcardStaysApprovedWithWarning keeps the --allow-remote
// path: a warned, policy-approved wildcard bind still serves.
func TestNewLocalhostWildcardStaysApprovedWithWarning(t *testing.T) {
	const listen = "0.0.0.0:0"
	dir := t.TempDir()
	policy := BindPolicy{Bind: listen, Warning: "x"}
	srv, err := New(Options{Listen: listen, StateDir: dir, Token: "tok", Stdout: &bytes.Buffer{}, Policy: policy})
	if srv != nil {
		t.Cleanup(func() { _ = srv.Close() })
	}
	if err != nil {
		t.Fatalf("New(Listen: %q, Policy: %+v) error = %v, want the warned remote bind to be accepted", listen, policy, err)
	}
}

// TestBindFlagsLocalhostBindServes pins the same seam through the command:
// --bind localhost:0 is approved by the policy, so the run must serve.
func TestBindFlagsLocalhostBindServes(t *testing.T) {
	forceGOOS(t, "linux")
	const bind = "localhost:0"
	r := newCmdRun(t, map[string]string{})

	var printed string
	code := r.run(Command(), []string{"--bind", bind}, func(u string) { printed = u })

	if code != 0 {
		t.Errorf("Command().Run([--bind %s]) = %d, want 0 (stderr %q)", bind, code, r.stderr.String())
	}
	if !r.served() {
		t.Errorf("Command().Run([--bind %s]) published no URL (stdout %q), want web.url written", bind, r.stdout.String())
	}
	if printed == "" {
		t.Fatalf("Command().Run([--bind %s]) published URL = %q, want a http:// URL", bind, printed)
	}
	u, err := url.Parse(printed)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", printed, err)
	}
	if ip := net.ParseIP(u.Hostname()); ip == nil || !ip.IsLoopback() {
		t.Errorf("published URL with --bind %s = %q, want a loopback host", bind, printed)
	}
	if got := r.stderr.String(); strings.Contains(got, "reachable from other machines") {
		t.Errorf("stderr with --bind %s = %q, want no remote-bind warning", bind, got)
	}
}
