package otlp_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/adeelahmad/snapback/internal/telemetry"
	"github.com/adeelahmad/snapback/internal/telemetry/otlp"
)

// scopeName is the instrumentation scope every Snapback metric is reported under.
const scopeName = "github.com/adeelahmad/snapback"

// closedAttrKeys is the only set of attribute keys that may appear anywhere in an
// encoded payload: the four resource keys plus the closed event attribute key set.
var closedAttrKeys = map[string]bool{
	"version":    true,
	"os":         true,
	"arch":       true,
	"install_id": true,
	"check":      true,
	"code":       true,
	"duration":   true,
	"outcome":    true,
}

func testResource() otlp.Resource {
	return otlp.Resource{
		Version:   "1.4.1",
		OS:        "linux",
		Arch:      "arm64",
		InstallID: "0123456789abcdef0123456789abcdef",
	}
}

func testEvents() []telemetry.Event {
	return []telemetry.Event{
		{
			Name: "snapback.command",
			Attrs: []telemetry.Attr{
				{Key: "outcome", Value: "ok"},
				{Key: "duration", Value: "<1s"},
			},
			Time: time.Unix(1700000000, 123456789).UTC(),
		},
		{
			Name: "snapback.doctor",
			Attrs: []telemetry.Attr{
				{Key: "check", Value: "mount"},
				{Key: "code", Value: "0"},
			},
			Time: time.Unix(1700000060, 987654321).UTC(),
		},
	}
}

func encodeToMap(t *testing.T, events []telemetry.Event, res otlp.Resource) map[string]any {
	t.Helper()
	b, err := otlp.Encode(events, res)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Encode produced invalid JSON %q: %v", b, err)
	}
	return got
}

func mapAt(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()
	v, ok := parent[key]
	if !ok {
		t.Fatalf("missing object %q, have keys %v", key, sortedKeys(parent))
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("%q is %T, want a JSON object", key, v)
	}
	return m
}

func sliceAt(t *testing.T, parent map[string]any, key string) []any {
	t.Helper()
	v, ok := parent[key]
	if !ok {
		t.Fatalf("missing array %q, have keys %v", key, sortedKeys(parent))
	}
	s, ok := v.([]any)
	if !ok {
		t.Fatalf("%q is %T, want a JSON array", key, v)
	}
	return s
}

func elemAt(t *testing.T, s []any, i int, what string) map[string]any {
	t.Helper()
	if len(s) <= i {
		t.Fatalf("%s has %d entries, want more than %d", what, len(s), i)
	}
	m, ok := s[i].(map[string]any)
	if !ok {
		t.Fatalf("%s[%d] is %T, want a JSON object", what, i, s[i])
	}
	return m
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// attrPairs flattens an OTLP attribute array into key -> stringValue, failing when
// an entry is not the {"key":…,"value":{"stringValue":…}} shape.
func attrPairs(t *testing.T, attrs []any, what string) map[string]string {
	t.Helper()
	out := make(map[string]string, len(attrs))
	for i, raw := range attrs {
		entry, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s[%d] is %T, want a JSON object", what, i, raw)
		}
		key, ok := entry["key"].(string)
		if !ok {
			t.Fatalf("%s[%d] has key %v (%T), want a string", what, i, entry["key"], entry["key"])
		}
		value := mapAt(t, entry, "value")
		sv, ok := value["stringValue"].(string)
		if !ok {
			t.Fatalf("%s[%d] value is %v, want {\"stringValue\": string}", what, i, value)
		}
		out[key] = sv
	}
	return out
}

// scopeMetricsOf walks down to the single scopeMetrics entry, asserting the
// resourceMetrics/scopeMetrics cardinality along the way.
func scopeMetricsOf(t *testing.T, req map[string]any) map[string]any {
	t.Helper()
	resourceMetrics := sliceAt(t, req, "resourceMetrics")
	if len(resourceMetrics) != 1 {
		t.Fatalf("resourceMetrics has %d entries, want exactly 1", len(resourceMetrics))
	}
	rm := elemAt(t, resourceMetrics, 0, "resourceMetrics")
	scopeMetrics := sliceAt(t, rm, "scopeMetrics")
	if len(scopeMetrics) != 1 {
		t.Fatalf("scopeMetrics has %d entries, want exactly 1", len(scopeMetrics))
	}
	return elemAt(t, scopeMetrics, 0, "scopeMetrics")
}

func TestEncodeCarriesResourceAttributesAsStringValues(t *testing.T) {
	res := testResource()
	req := encodeToMap(t, testEvents(), res)

	resourceMetrics := sliceAt(t, req, "resourceMetrics")
	if len(resourceMetrics) != 1 {
		t.Fatalf("resourceMetrics has %d entries, want exactly 1", len(resourceMetrics))
	}
	rm := elemAt(t, resourceMetrics, 0, "resourceMetrics")
	got := attrPairs(t, sliceAt(t, mapAt(t, rm, "resource"), "attributes"), "resource.attributes")

	want := map[string]string{
		"version":    res.Version,
		"os":         res.OS,
		"arch":       res.Arch,
		"install_id": res.InstallID,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resource.attributes = %v, want %v", got, want)
	}
}

func TestEncodeUsesTheSnapbackScope(t *testing.T) {
	sm := scopeMetricsOf(t, encodeToMap(t, testEvents(), testResource()))
	scope := mapAt(t, sm, "scope")
	if got := scope["name"]; got != scopeName {
		t.Errorf("scope.name = %v, want %q", got, scopeName)
	}
}

func TestEncodeMakesOneMonotonicDeltaSumPerEvent(t *testing.T) {
	events := testEvents()
	sm := scopeMetricsOf(t, encodeToMap(t, events, testResource()))

	metrics := sliceAt(t, sm, "metrics")
	if len(metrics) != len(events) {
		t.Fatalf("metrics has %d entries, want %d (one per event)", len(metrics), len(events))
	}

	for i, ev := range events {
		metric := elemAt(t, metrics, i, "metrics")
		if got := metric["name"]; got != ev.Name {
			t.Errorf("metrics[%d].name = %v, want %q", i, got, ev.Name)
		}

		sum := mapAt(t, metric, "sum")
		if got := sum["isMonotonic"]; got != true {
			t.Errorf("metrics[%d].sum.isMonotonic = %v, want true", i, got)
		}
		if got := sum["aggregationTemporality"]; got != float64(2) {
			t.Errorf("metrics[%d].sum.aggregationTemporality = %v, want 2 (DELTA)", i, got)
		}

		points := sliceAt(t, sum, "dataPoints")
		if len(points) != 1 {
			t.Fatalf("metrics[%d].sum.dataPoints has %d entries, want exactly 1", i, len(points))
		}
		point := elemAt(t, points, 0, "dataPoints")

		if got := point["asInt"]; got != "1" {
			t.Errorf("metrics[%d] dataPoint.asInt = %#v, want the string \"1\"", i, got)
		}
		wantNano := strconv.FormatInt(ev.Time.UnixNano(), 10)
		if got := point["timeUnixNano"]; got != wantNano {
			t.Errorf("metrics[%d] dataPoint.timeUnixNano = %#v, want the string %q", i, got, wantNano)
		}

		wantAttrs := make(map[string]string, len(ev.Attrs))
		for _, a := range ev.Attrs {
			wantAttrs[a.Key] = a.Value
		}
		gotAttrs := attrPairs(t, sliceAt(t, point, "attributes"), "dataPoint.attributes")
		if !reflect.DeepEqual(gotAttrs, wantAttrs) {
			t.Errorf("metrics[%d] dataPoint.attributes = %v, want %v", i, gotAttrs, wantAttrs)
		}
	}
}

func TestEncodeNoEventsKeepsTheResourceAndSendsNoMetrics(t *testing.T) {
	res := testResource()
	req := encodeToMap(t, nil, res)

	rm := elemAt(t, sliceAt(t, req, "resourceMetrics"), 0, "resourceMetrics")
	got := attrPairs(t, sliceAt(t, mapAt(t, rm, "resource"), "attributes"), "resource.attributes")
	want := map[string]string{
		"version":    res.Version,
		"os":         res.OS,
		"arch":       res.Arch,
		"install_id": res.InstallID,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("resource.attributes = %v, want %v", got, want)
	}

	sm := scopeMetricsOf(t, req)
	if metrics, ok := sm["metrics"]; ok {
		s, ok := metrics.([]any)
		if !ok {
			t.Fatalf("scopeMetrics.metrics is %T, want a JSON array", metrics)
		}
		if len(s) != 0 {
			t.Errorf("scopeMetrics.metrics has %d entries, want 0 for an empty batch", len(s))
		}
	}
}

// collectAttrKeys walks the whole decoded payload and returns every value of a
// "key" field, wherever it appears.
func collectAttrKeys(v any, out map[string]bool) {
	switch t := v.(type) {
	case map[string]any:
		if k, ok := t["key"].(string); ok {
			out[k] = true
		}
		for _, child := range t {
			collectAttrKeys(child, out)
		}
	case []any:
		for _, child := range t {
			collectAttrKeys(child, out)
		}
	}
}

func TestEncodeEmitsNoAttributeKeyOutsideTheClosedSet(t *testing.T) {
	events := append(testEvents(), telemetry.Event{
		Name:  "snapback.mount",
		Attrs: []telemetry.Attr{{Key: "outcome", Value: "error"}, {Key: "code", Value: "2"}},
		Time:  time.Unix(1700000120, 0).UTC(),
	})

	req := encodeToMap(t, events, testResource())

	got := map[string]bool{}
	collectAttrKeys(req, got)
	if len(got) == 0 {
		t.Fatalf("payload carries no attribute keys at all, want the resource and event keys")
	}
	for k := range got {
		if !closedAttrKeys[k] {
			t.Errorf("attribute key %q is outside the closed set %v", k, sortedSet(closedAttrKeys))
		}
	}
}

func sortedSet(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
