package projectdocs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// The self-hosted collector example: a docs page plus the two files an
// operator copies to stand up their own OpenTelemetry collector and
// GlitchTip instance. Snapback ships neither service nor a default endpoint.
const (
	telemetryCollectorDoc = "docs-site/telemetry-collector.md"
	telemetryComposeFile  = "docs/examples/telemetry/docker-compose.yml"
	telemetryOtelConfig   = "docs/examples/telemetry/otel-collector.yaml"
)

// composeService is the subset of a docker-compose service block this test
// cares about.
type composeService struct {
	Image       string   `yaml:"image"`
	Ports       []string `yaml:"ports"`
	EnvFile     any      `yaml:"env_file"`
	Environment any      `yaml:"environment"`
}

// composeDoc is the subset of docker-compose.yml this test cares about.
type composeDoc struct {
	Services map[string]composeService `yaml:"services"`
}

func loadTelemetryCompose(t *testing.T) composeDoc {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), telemetryComposeFile))
	if err != nil {
		t.Fatalf("read %s: %v", telemetryComposeFile, err)
	}
	var cf composeDoc
	if err := yaml.Unmarshal(data, &cf); err != nil {
		t.Fatalf("parse %s: %v", telemetryComposeFile, err)
	}
	return cf
}

func telemetryComposeServiceNames(cf composeDoc) []string {
	names := make([]string, 0, len(cf.Services))
	for name := range cf.Services {
		names = append(names, name)
	}
	return names
}

func TestTelemetryComposeHasExactlyTwoServices(t *testing.T) {
	cf := loadTelemetryCompose(t)
	if len(cf.Services) != 2 {
		t.Fatalf("%s defines %d services %v, want exactly 2 (otel-collector, glitchtip)",
			telemetryComposeFile, len(cf.Services), telemetryComposeServiceNames(cf))
	}
	for _, name := range []string{"otel-collector", "glitchtip"} {
		if _, ok := cf.Services[name]; !ok {
			t.Errorf("%s has no %q service, has %v", telemetryComposeFile, name, telemetryComposeServiceNames(cf))
		}
	}
}

func TestTelemetryComposeOtelCollectorImageIsTheContribDistribution(t *testing.T) {
	cf := loadTelemetryCompose(t)
	svc, ok := cf.Services["otel-collector"]
	if !ok {
		t.Fatalf("%s has no otel-collector service", telemetryComposeFile)
	}
	if !strings.HasPrefix(svc.Image, "otel/opentelemetry-collector-contrib:") {
		t.Errorf("%s otel-collector image = %q, want the otel/opentelemetry-collector-contrib distribution",
			telemetryComposeFile, svc.Image)
	}
}

func TestTelemetryComposeImagesArePinnedNotLatest(t *testing.T) {
	cf := loadTelemetryCompose(t)
	for name, svc := range cf.Services {
		if svc.Image == "" {
			t.Errorf("%s service %q has no image", telemetryComposeFile, name)
			continue
		}
		tag := ""
		if i := strings.LastIndex(svc.Image, ":"); i >= 0 {
			tag = svc.Image[i+1:]
		}
		if tag == "" || tag == "latest" {
			t.Errorf("%s service %q image %q is not pinned to a released tag",
				telemetryComposeFile, name, svc.Image)
		}
	}
}

func TestTelemetryComposeBindsEveryPublishedPortToLoopback(t *testing.T) {
	cf := loadTelemetryCompose(t)
	published := 0
	for name, svc := range cf.Services {
		for _, p := range svc.Ports {
			published++
			if !strings.HasPrefix(p, "127.0.0.1:") {
				t.Errorf("%s service %q publishes port %q, want it bound to 127.0.0.1 only",
					telemetryComposeFile, name, p)
			}
		}
	}
	if published == 0 {
		t.Fatalf("%s publishes no ports; otel-collector and glitchtip each need one reachable from the host",
			telemetryComposeFile)
	}
}

func TestTelemetryComposeGlitchtipUsesAnEnvFileNotPlainSecrets(t *testing.T) {
	cf := loadTelemetryCompose(t)
	svc, ok := cf.Services["glitchtip"]
	if !ok {
		t.Fatalf("%s has no glitchtip service", telemetryComposeFile)
	}
	if svc.EnvFile == nil {
		t.Errorf("%s glitchtip service has no env_file reference, so secrets have nowhere to live off-disk",
			telemetryComposeFile)
	}
}

// plainSecretKeys are environment keys that must never carry a literal value
// in the compose file itself; a real secret belongs in the referenced .env.
var plainSecretKeys = []string{"SECRET_KEY", "PASSWORD", "GLITCHTIP_DB_PASSWORD"}

func TestTelemetryComposeSetsNoDefaultCredentialsInPlainText(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), telemetryComposeFile))
	if err != nil {
		t.Fatalf("read %s: %v", telemetryComposeFile, err)
	}
	for i, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		for _, key := range plainSecretKeys {
			if !strings.HasPrefix(trimmed, key+":") {
				continue
			}
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, key+":"))
			if value != "" && !strings.Contains(value, "${") {
				t.Errorf("%s:%d: %s has a literal value %q, want a ${VAR} reference into .env",
					telemetryComposeFile, i+1, key, value)
			}
		}
	}
}

// otelReceiverOTLP is the subset of the otlp receiver block this test cares
// about: which protocols it exposes.
type otelReceiverOTLP struct {
	Protocols map[string]any `yaml:"protocols"`
}

// otelCollectorDoc is the subset of otel-collector.yaml this test cares about.
type otelCollectorDoc struct {
	Receivers struct {
		OTLP otelReceiverOTLP `yaml:"otlp"`
	} `yaml:"receivers"`
}

func loadOtelCollectorConfig(t *testing.T) otelCollectorDoc {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), telemetryOtelConfig))
	if err != nil {
		t.Fatalf("read %s: %v", telemetryOtelConfig, err)
	}
	var cfg otelCollectorDoc
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse %s: %v", telemetryOtelConfig, err)
	}
	return cfg
}

func TestOtelCollectorConfigHasAnHTTPOtlpReceiver(t *testing.T) {
	cfg := loadOtelCollectorConfig(t)
	if _, ok := cfg.Receivers.OTLP.Protocols["http"]; !ok {
		t.Errorf("%s otlp receiver has no http protocol block", telemetryOtelConfig)
	}
}

func TestOtelCollectorConfigHasNoGrpcProtocol(t *testing.T) {
	cfg := loadOtelCollectorConfig(t)
	if _, ok := cfg.Receivers.OTLP.Protocols["grpc"]; ok {
		t.Errorf("%s otlp receiver declares a grpc protocol block; only http is supported here", telemetryOtelConfig)
	}
}

func TestOtelCollectorConfigForwardsToGlitchtip(t *testing.T) {
	doc := strings.ToLower(readDoc(t, telemetryOtelConfig))
	if !strings.Contains(doc, "glitchtip") {
		t.Errorf("%s never names glitchtip in its exporter configuration", telemetryOtelConfig)
	}
}

func TestTelemetryCollectorPageExists(t *testing.T) {
	if doc := readDoc(t, telemetryCollectorDoc); strings.TrimSpace(doc) == "" {
		t.Errorf("%s is empty, want a page explaining the self-hosted collector", telemetryCollectorDoc)
	}
}

func TestTelemetryCollectorPageSaysSnapbackShipsNoCollector(t *testing.T) {
	doc := readDoc(t, telemetryCollectorDoc)
	if !lineWithAll(doc, "snapback", "no", "collector") {
		t.Errorf("%s has no sentence saying Snapback ships no collector", telemetryCollectorDoc)
	}
}

func TestTelemetryCollectorPageSaysNoDefaultEndpoint(t *testing.T) {
	doc := readDoc(t, telemetryCollectorDoc)
	if !lineWithAll(doc, "no default", "endpoint") {
		t.Errorf("%s has no sentence saying Snapback ships no default endpoint", telemetryCollectorDoc)
	}
}

func TestTelemetryCollectorPageNamesTheTwoComposeServices(t *testing.T) {
	doc := readDoc(t, telemetryCollectorDoc)
	for _, name := range []string{"otel-collector", "glitchtip"} {
		if !strings.Contains(doc, name) {
			t.Errorf("%s never names the %q compose service", telemetryCollectorDoc, name)
		}
	}
}

func TestTelemetryCollectorPageLinksTheExampleFiles(t *testing.T) {
	doc := readDoc(t, telemetryCollectorDoc)
	for _, ref := range []string{"docker-compose.yml", "otel-collector.yaml"} {
		if !strings.Contains(doc, ref) {
			t.Errorf("%s never references %q", telemetryCollectorDoc, ref)
		}
	}
}
