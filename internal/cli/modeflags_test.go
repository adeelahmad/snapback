package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/fsmode"
)

// modeFlagsUsage is the usage block the mode flag tests build their flag set
// from.
var modeFlagsUsage = Usage{
	Synopsis: "demo [flags]",
	Example:  "snapback demo -dir-mode 0750",
}

// credentialModeSentence is the sentence -h must state so that nobody reads
// these flags as a way to widen the credential store, the password files, the
// daemon lock, the pid file, the saved config or the diagnostic bundle.
const credentialModeSentence = "credential files stay 0600 and credential directories 0700 regardless"

// newModeFlags registers the shared mode flags on a fresh flag set and returns
// it together with the buffer its usage and errors are written to.
func newModeFlags(t *testing.T) (*bytes.Buffer, *flag.FlagSet, *ModeFlags) {
	t.Helper()
	var stderr bytes.Buffer
	fset := NewFlagSet(Env{Stdout: &bytes.Buffer{}, Stderr: &stderr}, modeFlagsUsage)
	return &stderr, fset, AddModeFlags(fset)
}

// resolveModeFlags parses args and resolves them over cfg, failing the test if
// either step reports an error.
func resolveModeFlags(t *testing.T, args []string, cfg config.Files) fsmode.Modes {
	t.Helper()
	_, fset, m := newModeFlags(t)
	if _, err := ParseWithUsage(fset, args); err != nil {
		t.Fatalf("ParseWithUsage(%q) err = %v, want nil", args, err)
	}
	got, err := m.Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve(%+v) after %q err = %v, want nil", cfg, args, err)
	}
	return got
}

func TestModeFlagsResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		cfg  config.Files
		want fsmode.Modes
	}{
		{
			name: "no flags and no config are the defaults",
			args: nil,
			cfg:  config.Files{},
			want: fsmode.Modes{Dir: 0o700, File: 0o600},
		},
		{
			name: "no flags keeps the explicit config modes",
			args: nil,
			cfg:  config.Files{DirMode: "0750", FileMode: "0640"},
			want: fsmode.Modes{Dir: 0o750, File: 0o640},
		},
		{
			name: "a flag umask replaces both config modes",
			args: []string{"--umask=027"},
			cfg:  config.Files{DirMode: "0700", FileMode: "0600"},
			want: fsmode.Modes{Dir: 0o750, File: 0o640},
		},
		{
			name: "a flag dir mode overrides only the config dir mode",
			args: []string{"--dir-mode=0755"},
			cfg:  config.Files{FileMode: "0640"},
			want: fsmode.Modes{Dir: 0o755, File: 0o640},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveModeFlags(t, tt.args, tt.cfg)
			if got != tt.want {
				t.Errorf("Resolve(%+v) after %q = %v, want %v",
					tt.cfg, tt.args, modesString(got), modesString(tt.want))
			}
		})
	}
}

func TestModeFlagsResolveRejectsUmaskWithExplicitMode(t *testing.T) {
	t.Parallel()

	args := []string{"--umask=027", "--dir-mode=0755"}
	_, fset, m := newModeFlags(t)
	if _, err := ParseWithUsage(fset, args); err != nil {
		t.Fatalf("ParseWithUsage(%q) err = %v, want nil", args, err)
	}

	got, err := m.Resolve(config.Files{})
	if err == nil {
		t.Fatalf("Resolve after %q = %v, nil, want a usage error", args, modesString(got))
	}
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Errorf("Resolve after %q err = %T, want *UsageError", args, err)
	}
	if !strings.Contains(err.Error(), "pick one") {
		t.Errorf("Resolve after %q err = %q, want it to say %q", args, err, "pick one")
	}
}

func TestModeFlagsResolveRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	args := []string{"--dir-mode=abc"}
	_, fset, m := newModeFlags(t)
	if _, err := ParseWithUsage(fset, args); err != nil {
		t.Fatalf("ParseWithUsage(%q) err = %v, want nil", args, err)
	}

	got, err := m.Resolve(config.Files{})
	if err == nil {
		t.Fatalf("Resolve after %q = %v, nil, want a usage error", args, modesString(got))
	}
	var ue *UsageError
	if !errors.As(err, &ue) {
		t.Errorf("Resolve after %q err = %T, want *UsageError", args, err)
	}
	for _, want := range []string{"-dir-mode", "abc"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Resolve after %q err = %q, want it to name %q", args, err, want)
		}
	}
}

func TestModeFlagsHelpDocumentsEveryFlag(t *testing.T) {
	t.Parallel()

	stderr, fset, _ := newModeFlags(t)
	help, err := ParseWithUsage(fset, []string{"-h"})
	if err != nil {
		t.Fatalf("ParseWithUsage(-h) err = %v, want nil", err)
	}
	if !help {
		t.Fatalf("ParseWithUsage(-h) help = false, want true")
	}

	got := stderr.String()
	for _, want := range []string{"-umask", "-dir-mode", "-file-mode", "0600", credentialModeSentence} {
		if !strings.Contains(got, want) {
			t.Errorf("-h stderr = %q, want it to contain %q", got, want)
		}
	}
}

// modesString renders a mode pair in the four-digit octal the config uses so
// that a failure reads like the setting it came from.
func modesString(m fsmode.Modes) string {
	return fmt.Sprintf("dir %04o, file %04o", m.Dir, m.File)
}
