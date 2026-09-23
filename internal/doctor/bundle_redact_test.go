package doctor

import (
	"strings"
	"testing"

	"github.com/adeelahmad/snapback/internal/config"
)

const (
	testRepoURI      = "s3:https://AKIAEXAMPLE:tok3n-s3cr3t@example.invalid/bucket/backups"
	testPasswordFile = "/home/adeel/.config/snapback/repo.pass"
	testEnvPassword  = "hunter2-very-secret"
)

func redactFixtureConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Repositories: []config.Repository{{
			ID:           "primary",
			Repository:   testRepoURI,
			PasswordFile: testPasswordFile,
			Environment:  map[string]string{"RESTIC_PASSWORD": testEnvPassword},
		}},
	}
}

func TestRedactForBundleRemovesRepositoryURIAndPasswordFile(t *testing.T) {
	doctorJSON := []byte(`{"repository":"` + testRepoURI + `","password_file":"` + testPasswordFile + `","host":"workstation"}`)
	daemonLog := []byte("opening " + testRepoURI + " with password file " + testPasswordFile + "\n")

	_, doctorOut, logOut, err := redactForBundle(redactFixtureConfig(), doctorJSON, daemonLog)
	if err != nil {
		t.Fatalf("redactForBundle(cfg, doctorJSON, daemonLog) returned error %v, want nil", err)
	}

	for _, tc := range []struct {
		name string
		got  string
	}{
		{"doctorOut", string(doctorOut)},
		{"logOut", string(logOut)},
	} {
		if strings.Contains(tc.got, testRepoURI) {
			t.Errorf("redactForBundle %s = %q, want no repository URI", tc.name, tc.got)
		}
		if strings.Contains(tc.got, testPasswordFile) {
			t.Errorf("redactForBundle %s = %q, want no password-file path", tc.name, tc.got)
		}
		if !strings.Contains(tc.got, "<redacted>") {
			t.Errorf("redactForBundle %s = %q, want it to contain \"<redacted>\"", tc.name, tc.got)
		}
	}

	// Hostnames stay: they are part of the doctor output the user consents to attach.
	if !strings.Contains(string(doctorOut), "workstation") {
		t.Errorf("redactForBundle doctorOut = %q, want the hostname kept", doctorOut)
	}
}

func TestRedactForBundleMasksConfigSecrets(t *testing.T) {
	configYAML, _, _, err := redactForBundle(redactFixtureConfig(), nil, nil)
	if err != nil {
		t.Fatalf("redactForBundle(cfg, nil, nil) returned error %v, want nil", err)
	}
	if strings.Contains(string(configYAML), testEnvPassword) {
		t.Errorf("redactForBundle configYAML = %q, want no password value", configYAML)
	}
	if !strings.Contains(string(configYAML), "***") {
		t.Errorf("redactForBundle configYAML = %q, want the masked marker %q", configYAML, "***")
	}
	if strings.Contains(string(configYAML), testRepoURI) {
		t.Errorf("redactForBundle configYAML = %q, want no repository URI", configYAML)
	}
	if strings.Contains(string(configYAML), testPasswordFile) {
		t.Errorf("redactForBundle configYAML = %q, want no password-file path", configYAML)
	}
	if !strings.Contains(string(configYAML), "primary") {
		t.Errorf("redactForBundle configYAML = %q, want the repository id kept", configYAML)
	}
}

func TestRedactForBundleMasksTelemetryEndpoints(t *testing.T) {
	cfg := redactFixtureConfig()
	cfg.Telemetry.Enabled = true
	cfg.Telemetry.Endpoint = "http://127.0.0.1:61559"
	cfg.Telemetry.CrashReports = true
	cfg.Telemetry.CrashEndpoint = "http://127.0.0.1:61560"

	configYAML, _, _, err := redactForBundle(cfg, nil, nil)
	if err != nil {
		t.Fatalf("redactForBundle(cfg, nil, nil) returned error %v, want nil", err)
	}
	for _, endpoint := range []string{cfg.Telemetry.Endpoint, cfg.Telemetry.CrashEndpoint} {
		if strings.Contains(string(configYAML), endpoint) {
			t.Errorf("redactForBundle configYAML = %q, want no %q", configYAML, endpoint)
		}
	}
	if !strings.Contains(string(configYAML), "***") {
		t.Errorf("redactForBundle configYAML = %q, want the masked marker %q", configYAML, "***")
	}
	if !strings.Contains(string(configYAML), "endpoint") || !strings.Contains(string(configYAML), "crash_endpoint") {
		t.Errorf("redactForBundle configYAML = %q, want the telemetry keys kept", configYAML)
	}
}

func TestRedactForBundleMasksPasswordValuePatterns(t *testing.T) {
	doctorJSON := []byte(`{"RESTIC_PASSWORD":"other-secret","password":"another-secret"}`)

	_, doctorOut, _, err := redactForBundle(redactFixtureConfig(), doctorJSON, nil)
	if err != nil {
		t.Fatalf("redactForBundle(cfg, doctorJSON, nil) returned error %v, want nil", err)
	}
	for _, want := range []string{"other-secret", "another-secret"} {
		if strings.Contains(string(doctorOut), want) {
			t.Errorf("redactForBundle doctorOut = %q, want no %q value", doctorOut, want)
		}
	}
}
