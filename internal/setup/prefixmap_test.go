package setup

import (
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

func TestDerivePrefixMap(t *testing.T) {
	tests := []struct {
		name         string
		local        string
		snaps        []ProbedSnapshot
		wantHostname string
		wantMappings []config.PrefixMapping
		wantReason   string
	}{
		{
			name:         "identical path needs no mapping",
			local:        "/Users/u/p",
			snaps:        []ProbedSnapshot{{Hostname: "mac", Paths: []string{"/Users/u/p"}}},
			wantHostname: "mac",
		},
		{
			name:         "linux source path maps onto the macOS local path",
			local:        "/Users/u/p",
			snaps:        []ProbedSnapshot{{Hostname: "workstation", Paths: []string{"/home/u/p"}}},
			wantHostname: "workstation",
			wantMappings: []config.PrefixMapping{
				{Hostname: "workstation", SourcePath: "/home/u/p", TreePrefix: "/home/u/p"},
			},
		},
		{
			name:  "majority host wins",
			local: "/Users/u/p",
			snaps: []ProbedSnapshot{
				{Hostname: "laptop", Paths: []string{"/home/u/p"}},
				{Hostname: "workstation", Paths: []string{"/srv/u/p"}},
				{Hostname: "workstation", Paths: []string{"/srv/u/p"}},
			},
			wantHostname: "workstation",
			wantMappings: []config.PrefixMapping{
				{Hostname: "workstation", SourcePath: "/srv/u/p", TreePrefix: "/srv/u/p"},
			},
		},
		{
			name:         "no snapshot path matches the local path",
			local:        "/Users/u/p",
			snaps:        []ProbedSnapshot{{Hostname: "mac", Paths: []string{"/var/log"}}},
			wantHostname: "mac",
			wantReason:   "no snapshot contains /Users/u/p; run: snapback snap /Users/u/p",
		},
		{
			name:       "no snapshots at all",
			local:      "/Users/u/p",
			wantReason: "repository has no snapshots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostname, mappings, reason := derivePrefixMap(tt.local, tt.snaps)
			if hostname != tt.wantHostname {
				t.Errorf("derivePrefixMap(%q, snaps) hostname = %q, want %q", tt.local, hostname, tt.wantHostname)
			}
			if reason != tt.wantReason {
				t.Errorf("derivePrefixMap(%q, snaps) reason = %q, want %q", tt.local, reason, tt.wantReason)
			}
			if len(mappings) != len(tt.wantMappings) {
				t.Fatalf("derivePrefixMap(%q, snaps) mappings = %v, want %v", tt.local, mappings, tt.wantMappings)
			}
			for i, got := range mappings {
				if got != tt.wantMappings[i] {
					t.Errorf("derivePrefixMap(%q, snaps) mappings[%d] = %v, want %v", tt.local, i, got, tt.wantMappings[i])
				}
			}
		})
	}
}
