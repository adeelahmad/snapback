package cli

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/telemetry"
)

// toggleFixture is one on-disk config a toggle test runs enable/disable
// against, mirroring a real `snapback telemetry enable|disable` invocation:
// a validated config.yaml the command reads through Deps.LoadConfig and
// mutates through the same config-save path the web UI uses.
type toggleFixture struct {
	env  Env
	deps Deps
	errb *bytes.Buffer
	path string
}

// newToggleFixture writes a minimal, valid config carrying telemetryYAML
// (for example "telemetry:\n  endpoint: https://c:4318\n", or "" for no
// telemetry section at all) and returns a fixture ready to drive
// runTelemetryEnable/runTelemetryDisable against it.
func newToggleFixture(t *testing.T, telemetryYAML string) *toggleFixture {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "xs"))

	pw := filepath.Join(tmp, "password")
	if err := os.WriteFile(pw, []byte("s3cret\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", pw, err)
	}
	data := "version: 1\n" + telemetryYAML +
		"repositories:\n" +
		"  - id: personal\n" +
		"    repository: rclone:gdrive:Backups/restic\n" +
		"    restic_binary: /usr/bin/restic\n" +
		"    rclone_binary: /usr/bin/rclone\n" +
		"    password_file: " + pw + "\n" +
		"roots:\n" +
		"  - id: work\n" +
		"    local_path: " + filepath.Join(tmp, "work") + "\n" +
		"    repository_id: personal\n"

	path := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", path, err)
	}
	if _, err := config.Parse([]byte(data)); err != nil {
		t.Fatalf("config.Parse(fixture) = %v, want nil error (fixture is invalid)", err)
	}

	var out, errb bytes.Buffer
	return &toggleFixture{
		env: Env{
			Stdout:     &out,
			Stderr:     &errb,
			Getenv:     func(string) string { return "" },
			ConfigPath: path,
		},
		deps: Deps{
			LoadConfig: func(p string) (config.Config, error) {
				c, _, err := config.Load(p)
				if err != nil {
					return config.Config{}, err
				}
				return *c, nil
			},
		},
		errb: &errb,
		path: path,
	}
}

// load re-reads the fixture's config file from disk.
func (f *toggleFixture) load(t *testing.T) config.Config {
	t.Helper()
	cfg, err := f.deps.LoadConfig(f.path)
	if err != nil {
		t.Fatalf("LoadConfig(%q) = %v, want nil", f.path, err)
	}
	return cfg
}

// TestTelemetryEnableSavesThroughConfigSavePath pins that `enable` sets
// telemetry.enabled: true through the same config-save path the web UI
// uses: the write validates and round-trips, and it leaves every field the
// command never touches -- the fixture's repository and root -- intact.
func TestTelemetryEnableSavesThroughConfigSavePath(t *testing.T) {
	f := newToggleFixture(t, "telemetry:\n  endpoint: https://collector:4318\n")

	got := runTelemetryEnable(context.Background(), f.env, f.deps, false)

	if got != 0 {
		t.Fatalf("runTelemetryEnable() = %d, want 0 (stderr %q)", got, f.errb.String())
	}
	cfg := f.load(t)
	if !cfg.Telemetry.Enabled {
		t.Errorf("after enable, Telemetry.Enabled = false, want true")
	}
	if cfg.Telemetry.Endpoint != "https://collector:4318" {
		t.Errorf("after enable, Telemetry.Endpoint = %q, want unchanged %q", cfg.Telemetry.Endpoint, "https://collector:4318")
	}
	if len(cfg.Repositories) != 1 || cfg.Repositories[0].ID != "personal" {
		t.Errorf("after enable, Repositories = %+v, want the untouched fixture repository", cfg.Repositories)
	}
	if len(cfg.Roots) != 1 || cfg.Roots[0].ID != "work" {
		t.Errorf("after enable, Roots = %+v, want the untouched fixture root", cfg.Roots)
	}
}

// TestTelemetryEnableRefusesWithoutEndpoint pins that `enable` refuses with
// a clear error naming telemetry.endpoint and the privacy docs page when no
// endpoint is configured, and makes no partial write.
func TestTelemetryEnableRefusesWithoutEndpoint(t *testing.T) {
	f := newToggleFixture(t, "")

	got := runTelemetryEnable(context.Background(), f.env, f.deps, false)

	if got == 0 {
		t.Fatalf("runTelemetryEnable() without an endpoint = 0, want non-zero")
	}
	stderr := f.errb.String()
	if !strings.Contains(stderr, "telemetry.endpoint") {
		t.Errorf("runTelemetryEnable() stderr = %q, want it to name telemetry.endpoint", stderr)
	}
	if !strings.Contains(stderr, "https://snapback.run/privacy") {
		t.Errorf("runTelemetryEnable() stderr = %q, want it to point at the privacy docs page", stderr)
	}
	if cfg := f.load(t); cfg.Telemetry.Enabled {
		t.Errorf("after a refused enable, Telemetry.Enabled = true, want false (no partial write)")
	}
}

// TestTelemetryEnableIsIdempotent pins that calling `enable` on an
// already-enabled config is a no-op success, not an error.
func TestTelemetryEnableIsIdempotent(t *testing.T) {
	f := newToggleFixture(t, "telemetry:\n  enabled: true\n  endpoint: https://collector:4318\n")

	got := runTelemetryEnable(context.Background(), f.env, f.deps, false)

	if got != 0 {
		t.Fatalf("runTelemetryEnable() on an already-enabled config = %d, want 0 (stderr %q)", got, f.errb.String())
	}
	if f.errb.Len() != 0 {
		t.Errorf("runTelemetryEnable() on an already-enabled config stderr = %q, want empty", f.errb.String())
	}
	cfg := f.load(t)
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Endpoint != "https://collector:4318" {
		t.Errorf("after idempotent enable, Telemetry = %+v, want unchanged enabled config", cfg.Telemetry)
	}
}

// TestTelemetryDisableSavesAndForgetsInstallID pins that `disable` sets
// telemetry.enabled: false and calls telemetry.ForgetInstallID so
// <state_dir>/telemetry/install_id is gone afterwards.
func TestTelemetryDisableSavesAndForgetsInstallID(t *testing.T) {
	f := newToggleFixture(t, "telemetry:\n  enabled: true\n  endpoint: https://collector:4318\n")
	cfg := f.load(t)
	installDir := filepath.Join(cfg.StateDir, "telemetry")
	if _, err := telemetry.InstallID(installDir); err != nil {
		t.Fatalf("telemetry.InstallID(%q) = %v, want nil", installDir, err)
	}
	idPath := filepath.Join(installDir, "install_id")
	if _, err := os.Stat(idPath); err != nil {
		t.Fatalf("Stat(%q) = %v, want the install id to exist before disable", idPath, err)
	}

	got := runTelemetryDisable(context.Background(), f.env, f.deps, false)

	if got != 0 {
		t.Fatalf("runTelemetryDisable() = %d, want 0 (stderr %q)", got, f.errb.String())
	}
	if got := f.load(t); got.Telemetry.Enabled {
		t.Errorf("after disable, Telemetry.Enabled = true, want false")
	}
	if _, err := os.Stat(idPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(%q) after disable = %v, want fs.ErrNotExist", idPath, err)
	}
}

// TestTelemetryDisableIsIdempotent pins that calling `disable` twice on an
// already-off config succeeds both times without error.
func TestTelemetryDisableIsIdempotent(t *testing.T) {
	f := newToggleFixture(t, "")

	for i := 0; i < 2; i++ {
		got := runTelemetryDisable(context.Background(), f.env, f.deps, false)
		if got != 0 {
			t.Fatalf("runTelemetryDisable() call %d on an already-off config = %d, want 0 (stderr %q)", i+1, got, f.errb.String())
		}
	}
	if cfg := f.load(t); cfg.Telemetry.Enabled {
		t.Errorf("after idempotent disable, Telemetry.Enabled = true, want false")
	}
}

// countingRoundTripper is defined in telemetry_show_test.go.

// TestTelemetryToggleMakesNoNetworkCall pins that neither `enable` nor
// `disable` ever dials the network, even though enable's config carries a
// live-looking collector endpoint.
func TestTelemetryToggleMakesNoNetworkCall(t *testing.T) {
	f := newToggleFixture(t, "telemetry:\n  endpoint: https://collector:4318\n")
	rt := &countingRoundTripper{}
	prev := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = prev }()

	if got := runTelemetryEnable(context.Background(), f.env, f.deps, false); got != 0 {
		t.Fatalf("runTelemetryEnable() = %d, want 0 (stderr %q)", got, f.errb.String())
	}
	if got := runTelemetryDisable(context.Background(), f.env, f.deps, false); got != 0 {
		t.Fatalf("runTelemetryDisable() = %d, want 0 (stderr %q)", got, f.errb.String())
	}
	if rt.calls != 0 {
		t.Errorf("countingRoundTripper.calls = %d, want 0: enable/disable must never dial the network", rt.calls)
	}
}
