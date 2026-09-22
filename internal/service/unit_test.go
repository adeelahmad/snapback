package service

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/errcode"
)

const unitMarker = "# Managed by snapback; do not edit."

var (
	userOpts = UnitOptions{
		Exe:    "/usr/local/bin/snapback",
		Config: "/home/u/.config/snapback/config.yaml",
		Scope:  "user",
	}
	hostileOpts = UnitOptions{
		Exe:    "/opt/my apps/snapback",
		Config: "/home/u/cfg 100%/$HOME \"q\" \\x.yaml",
		Scope:  "user",
	}
	systemOpts = UnitOptions{
		Exe:    "/usr/local/bin/snapback",
		Config: "/etc/snapback/config.yaml",
		Scope:  "system",
		User:   "alice",
	}
)

func readGolden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	if len(b) == 0 {
		t.Fatalf("golden %s is empty", name)
	}
	return string(b)
}

// systemdUnquote splits an ExecStart value the way systemd does for the
// quoting SystemdUnit emits, then undoes the %% and $$ specifier escapes.
func systemdUnquote(t *testing.T, s string) []string {
	t.Helper()
	var (
		args    []string
		cur     strings.Builder
		inQuote bool
		inWord  bool
	)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote && c == '\\' && i+1 < len(s):
			i++
			cur.WriteByte(s[i])
		case c == '"':
			inQuote = !inQuote
			inWord = true
		case !inQuote && c == ' ':
			if inWord {
				args = append(args, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteByte(c)
			inWord = true
		}
	}
	if inQuote {
		t.Fatalf("systemdUnquote(%q): unterminated quote", s)
	}
	if inWord {
		args = append(args, cur.String())
	}
	for i, a := range args {
		args[i] = strings.NewReplacer("%%", "%", "$$", "$").Replace(a)
	}
	return args
}

func TestSystemdUnitUserGolden(t *testing.T) {
	want := readGolden(t, "user.service.golden")

	got, err := SystemdUnit(userOpts)
	if err != nil {
		t.Fatalf("SystemdUnit(%+v) error = %v, want nil", userOpts, err)
	}
	if got != want {
		t.Errorf("SystemdUnit(%+v) = %q, want %q", userOpts, got, want)
	}
	for _, line := range []string{"Restart=on-failure", "RestartSec=5s", "TimeoutStopSec=30s", "WantedBy=default.target"} {
		if !slices.Contains(strings.Split(got, "\n"), line) {
			t.Errorf("SystemdUnit(%+v) lacks line %q", userOpts, line)
		}
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "User=") {
			t.Errorf("SystemdUnit(%+v) has %q, want no User= line in user scope", userOpts, line)
		}
	}
}

func TestSystemdUnitHostilePathsGolden(t *testing.T) {
	want := readGolden(t, "hostile.service.golden")

	got, err := SystemdUnit(hostileOpts)
	if err != nil {
		t.Fatalf("SystemdUnit(%+v) error = %v, want nil", hostileOpts, err)
	}
	if got != want {
		t.Errorf("SystemdUnit(%+v) = %q, want %q", hostileOpts, got, want)
	}

	var exec string
	found := false
	for _, line := range strings.Split(got, "\n") {
		if v, ok := strings.CutPrefix(line, "ExecStart="); ok {
			exec, found = v, true
			break
		}
	}
	if !found {
		t.Fatalf("SystemdUnit(%+v) has no ExecStart= line", hostileOpts)
	}
	gotArgs := systemdUnquote(t, exec)
	wantArgs := []string{hostileOpts.Exe, "run", "--config", hostileOpts.Config}
	if !slices.Equal(gotArgs, wantArgs) {
		t.Errorf("unquoted ExecStart %q = %q, want %q", exec, gotArgs, wantArgs)
	}
}

func TestSystemdUnitSystemScopeAndBadInput(t *testing.T) {
	want := readGolden(t, "system.service.golden")

	got, err := SystemdUnit(systemOpts)
	if err != nil {
		t.Fatalf("SystemdUnit(%+v) error = %v, want nil", systemOpts, err)
	}
	if got != want {
		t.Errorf("SystemdUnit(%+v) = %q, want %q", systemOpts, got, want)
	}
	lines := strings.Split(got, "\n")
	for _, line := range []string{"User=alice", "WantedBy=multi-user.target"} {
		if !slices.Contains(lines, line) {
			t.Errorf("SystemdUnit(%+v) lacks line %q", systemOpts, line)
		}
	}

	bad := []struct {
		name string
		opts UnitOptions
	}{
		{"relative exe", UnitOptions{Exe: "snapback", Config: userOpts.Config, Scope: "user"}},
		{"newline in config", UnitOptions{Exe: userOpts.Exe, Config: "/home/u/a\nb.yaml", Scope: "user"}},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SystemdUnit(tc.opts)
			if got := errcode.Of(err); got != errcode.InvalidConfig {
				t.Errorf("errcode.Of(SystemdUnit(%+v)) = %q, want %q (err %v)", tc.opts, got, errcode.InvalidConfig, err)
			}
		})
	}
}

func TestSystemdUnitHasNoMountHidingOptions(t *testing.T) {
	forbidden := []string{"PrivateTmp", "ProtectHome", "ProtectSystem", "PrivateMounts", "MountFlags", "ReadOnlyPaths", "InaccessiblePaths"}
	for _, opts := range []UnitOptions{userOpts, hostileOpts, systemOpts} {
		got, err := SystemdUnit(opts)
		if err != nil {
			t.Errorf("SystemdUnit(%+v) error = %v, want nil", opts, err)
			continue
		}
		if !strings.HasPrefix(got, unitMarker+"\n") {
			t.Errorf("SystemdUnit(%+v) = %q, want it to start with %q", opts, got, unitMarker)
		}
		for _, f := range forbidden {
			if strings.Contains(got, f) {
				t.Errorf("SystemdUnit(%+v) contains %q, which hides mounts", opts, f)
			}
		}
	}
}
