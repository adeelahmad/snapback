package config

import (
	"bytes"
	"testing"
)

// parsedExampleWithEnv parses testdata/example.yaml and gives its repository
// a fake RCLONE_CONFIG and AWS_SECRET_ACCESS_KEY.
func parsedExampleWithEnv(t *testing.T) *Config {
	t.Helper()
	_, data := exampleYAML(t)
	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(example) = %v, want nil error", err)
	}
	if len(c.Repositories) == 0 {
		t.Fatal("Parse(example) has no repositories")
	}
	c.Repositories[0].Environment = map[string]string{
		"RCLONE_CONFIG":         "/p/rclone.conf",
		"AWS_SECRET_ACCESS_KEY": "abc123",
	}
	return c
}

func TestRedactMasksEnvironmentValues(t *testing.T) {
	c := parsedExampleWithEnv(t)

	r := Redact(c)
	if r == nil || len(r.Repositories) != 1 {
		t.Fatalf("Redact(c) = %+v, want a copy with 1 repository", r)
	}
	env := r.Repositories[0].Environment
	if len(env) != 2 {
		t.Errorf("Redact(c) environment = %v, want 2 keys", env)
	}
	for _, k := range []string{"RCLONE_CONFIG", "AWS_SECRET_ACCESS_KEY"} {
		if got, ok := env[k]; !ok || got != "***" {
			t.Errorf("Redact(c) environment[%q] = %q, %v, want %q, true", k, got, ok, "***")
		}
	}
	if got, want := r.Repositories[0].PasswordFile, c.Repositories[0].PasswordFile; got != want || got == "" {
		t.Errorf("Redact(c) PasswordFile = %q, want %q", got, want)
	}

	b, err := Marshal(r)
	if err != nil {
		t.Fatalf("Marshal(Redact(c)) = %v, want nil error", err)
	}
	if !bytes.Contains(b, []byte("RCLONE_CONFIG")) {
		t.Errorf("Marshal(Redact(c)) = %q, want it to contain RCLONE_CONFIG", b)
	}
	for _, secret := range []string{"abc123", "/p/rclone.conf"} {
		if bytes.Contains(b, []byte(secret)) {
			t.Errorf("Marshal(Redact(c)) = %q, want no %q", b, secret)
		}
	}
}

func TestRedactDoesNotMutateInput(t *testing.T) {
	c := parsedExampleWithEnv(t)
	if len(c.Roots) == 0 || len(c.Roots[0].SeedPaths) == 0 || len(c.Catalog.ReaderPolicy.DenyProcesses) == 0 {
		t.Fatal("Parse(example) lacks seed paths or deny processes")
	}
	seed := c.Roots[0].SeedPaths[0].Path
	deny := c.Catalog.ReaderPolicy.DenyProcesses[0]

	r := Redact(c)
	if r == nil || len(r.Roots) == 0 || len(r.Roots[0].SeedPaths) == 0 || len(r.Catalog.ReaderPolicy.DenyProcesses) == 0 {
		t.Fatalf("Redact(c) = %+v, want a full copy", r)
	}
	if got := c.Repositories[0].Environment["AWS_SECRET_ACCESS_KEY"]; got != "abc123" {
		t.Errorf("after Redact, c environment[AWS_SECRET_ACCESS_KEY] = %q, want %q", got, "abc123")
	}
	if got := c.Repositories[0].Environment["RCLONE_CONFIG"]; got != "/p/rclone.conf" {
		t.Errorf("after Redact, c environment[RCLONE_CONFIG] = %q, want %q", got, "/p/rclone.conf")
	}

	r.Roots[0].SeedPaths[0].Path = "changed"
	r.Catalog.ReaderPolicy.DenyProcesses[0] = "changed"
	if got := c.Roots[0].SeedPaths[0].Path; got != seed {
		t.Errorf("c.Roots[0].SeedPaths[0].Path = %q after mutating the copy, want %q", got, seed)
	}
	if got := c.Catalog.ReaderPolicy.DenyProcesses[0]; got != deny {
		t.Errorf("c.Catalog.ReaderPolicy.DenyProcesses[0] = %q after mutating the copy, want %q", got, deny)
	}
}

func TestRedactMasksTelemetryEndpoints(t *testing.T) {
	c := parsedExampleWithEnv(t)
	c.Telemetry.Enabled = true
	c.Telemetry.Endpoint = "http://127.0.0.1:61559"
	c.Telemetry.CrashReports = true
	c.Telemetry.CrashEndpoint = "http://127.0.0.1:61560"

	r := Redact(c)
	if r.Telemetry.Endpoint != "***" {
		t.Errorf("Redact(c).Telemetry.Endpoint = %q, want %q", r.Telemetry.Endpoint, "***")
	}
	if r.Telemetry.CrashEndpoint != "***" {
		t.Errorf("Redact(c).Telemetry.CrashEndpoint = %q, want %q", r.Telemetry.CrashEndpoint, "***")
	}
	if !r.Telemetry.Enabled || !r.Telemetry.CrashReports {
		t.Errorf("Redact(c).Telemetry = %+v, want enabled and crash_reports kept", r.Telemetry)
	}

	b, err := Marshal(r)
	if err != nil {
		t.Fatalf("Marshal(Redact(c)) = %v, want nil error", err)
	}
	for _, endpoint := range []string{c.Telemetry.Endpoint, c.Telemetry.CrashEndpoint} {
		if bytes.Contains(b, []byte(endpoint)) {
			t.Errorf("Marshal(Redact(c)) = %q, want no %q", b, endpoint)
		}
	}
	if !bytes.Contains(b, []byte("endpoint")) || !bytes.Contains(b, []byte("crash_endpoint")) {
		t.Errorf("Marshal(Redact(c)) = %q, want the telemetry keys kept", b)
	}
}

func TestRedactLeavesEmptyTelemetryEndpointsEmpty(t *testing.T) {
	c := parsedExampleWithEnv(t)

	r := Redact(c)
	if r.Telemetry.Endpoint != "" {
		t.Errorf("Redact(c).Telemetry.Endpoint = %q, want empty", r.Telemetry.Endpoint)
	}
	if r.Telemetry.CrashEndpoint != "" {
		t.Errorf("Redact(c).Telemetry.CrashEndpoint = %q, want empty", r.Telemetry.CrashEndpoint)
	}
}

func TestRedactNil(t *testing.T) {
	if got := Redact(nil); got != nil {
		t.Errorf("Redact(nil) = %+v, want nil", got)
	}
}
