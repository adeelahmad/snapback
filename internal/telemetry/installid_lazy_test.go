package telemetry_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// TestInstallIDLazy_DisabledClientNeverTouchesInstallIDDir guards that a
// telemetry client which will never deliver anything -- either explicitly
// disabled, or enabled with no endpoint to deliver to -- never creates the
// install-id directory. Nothing wires InstallID into New, Emit or Close
// today, so this stands as a regression guard: it must keep passing as the
// client evolves, not start failing.
func TestInstallIDLazy_DisabledClientNeverTouchesInstallIDDir(t *testing.T) {
	cases := []struct {
		name string
		opts func(*telemetry.Fake) telemetry.Options
	}{
		{
			name: "explicitly disabled",
			opts: func(f *telemetry.Fake) telemetry.Options {
				return telemetry.Options{Enabled: false, Endpoint: "https://example.invalid", Exporter: f}
			},
		},
		{
			name: "enabled with empty endpoint",
			opts: func(f *telemetry.Fake) telemetry.Options {
				return telemetry.Options{Enabled: true, Endpoint: "", Exporter: f}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stateDir := t.TempDir()
			installDir := filepath.Join(stateDir, "telemetry")

			fake := &telemetry.Fake{}
			c := telemetry.New(tc.opts(fake))
			if c.Enabled() {
				t.Fatalf("client is enabled, want disabled for case %q", tc.name)
			}

			ctx := context.Background()
			for _, name := range telemetry.Names() {
				c.Emit(ctx, telemetry.Event{Name: name, Time: time.Now()})
			}
			if err := c.Close(ctx); err != nil {
				t.Fatalf("Close: %v", err)
			}

			if _, err := os.Stat(installDir); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("os.Stat(%q) = %v, want fs.ErrNotExist: a disabled client touched the install-id dir", installDir, err)
			}

			entries, err := os.ReadDir(stateDir)
			if err != nil {
				t.Fatalf("ReadDir(%q): %v", stateDir, err)
			}
			if len(entries) != 0 {
				t.Fatalf("stateDir has %d entries, want 0: %v", len(entries), entries)
			}
		})
	}
}
