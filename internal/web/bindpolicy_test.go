package web

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

// remoteWarning is the exact line a remote bind must carry. It is pinned here
// so the wording cannot drift without a test change.
const remoteWarning = "warning: snapback web is reachable from other machines on %s; the session token is the only protection"

func warningFor(bind string) string {
	return strings.Replace(remoteWarning, "%s", bind, 1)
}

// wantInvalidConfig asserts err is an InvalidConfig-class *errcode.Error whose
// message mentions want.
func wantInvalidConfig(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Policy error = nil, want an InvalidConfig error naming %q", want)
	}
	var ce *errcode.Error
	if !errors.As(err, &ce) {
		t.Fatalf("Policy error = %v (%T), want *errcode.Error", err, err)
	}
	if ce.Code != errcode.InvalidConfig {
		t.Errorf("Policy error code = %q, want %q", ce.Code, errcode.InvalidConfig)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("Policy error = %q, want it to contain %q", err.Error(), want)
	}
}

func TestPolicyBind(t *testing.T) {
	tests := []struct {
		name        string
		bind        string
		allowRemote bool
		wantBind    string
		wantWarning string
	}{
		{name: "loopback ipv4", bind: "127.0.0.1:7373", wantBind: "127.0.0.1:7373"},
		{name: "loopback ipv6", bind: "[::1]:7373", wantBind: "[::1]:7373"},
		{name: "loopback name", bind: "localhost:7373", wantBind: "localhost:7373"},
		{name: "empty is the listen default", bind: "", wantBind: "127.0.0.1:0"},
		{
			name:        "loopback with allow-remote stays quiet",
			bind:        "127.0.0.1:7373",
			allowRemote: true,
			wantBind:    "127.0.0.1:7373",
		},
		{
			name:        "wildcard with allow-remote",
			bind:        "0.0.0.0:7373",
			allowRemote: true,
			wantBind:    "0.0.0.0:7373",
			wantWarning: warningFor("0.0.0.0:7373"),
		},
		{
			name:        "lan address with allow-remote",
			bind:        "192.168.1.5:7373",
			allowRemote: true,
			wantBind:    "192.168.1.5:7373",
			wantWarning: warningFor("192.168.1.5:7373"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Policy(tt.bind, tt.allowRemote, nil)
			if err != nil {
				t.Fatalf("Policy(%q, %t, nil) error = %v, want nil", tt.bind, tt.allowRemote, err)
			}
			if got.Bind != tt.wantBind {
				t.Errorf("Bind = %q, want %q", got.Bind, tt.wantBind)
			}
			if got.Warning != tt.wantWarning {
				t.Errorf("Warning = %q, want %q", got.Warning, tt.wantWarning)
			}
		})
	}
}

func TestPolicyRemoteBindNeedsFlag(t *testing.T) {
	tests := []struct {
		name string
		bind string
	}{
		{name: "wildcard ipv4", bind: "0.0.0.0:7373"},
		{name: "lan address", bind: "192.168.1.5:7373"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Policy(tt.bind, false, nil)
			wantInvalidConfig(t, err, "--allow-remote")
			if got.Bind != "" {
				t.Errorf("Bind = %q, want %q on a refused bind", got.Bind, "")
			}
		})
	}
}

func TestPolicyAllowOrigins(t *testing.T) {
	origins := []string{"https://x.example", "http://10.0.0.2:8080"}
	got, err := Policy("127.0.0.1:7373", false, origins)
	if err != nil {
		t.Fatalf("Policy(..., %v) error = %v, want nil", origins, err)
	}
	if !slices.Equal(got.AllowOrigins, origins) {
		t.Errorf("AllowOrigins = %v, want %v", got.AllowOrigins, origins)
	}
}

func TestPolicyRejectsBadOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		// want is the bad entry as the message must quote it (%q).
		want string
	}{
		{name: "no scheme", origin: "x.example", want: `"x.example"`},
		{name: "has a path", origin: "https://x.example/path", want: `"https://x.example/path"`},
		{name: "empty entry", origin: "", want: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Policy("127.0.0.1:7373", false, []string{tt.origin})
			wantInvalidConfig(t, err, tt.want)
			if err != nil && !strings.Contains(err.Error(), "--allow-origin") {
				t.Errorf("Policy error = %q, want it to contain %q", err.Error(), "--allow-origin")
			}
		})
	}
}
