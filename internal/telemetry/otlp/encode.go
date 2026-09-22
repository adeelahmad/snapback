package otlp

import (
	"encoding/json"
	"strconv"

	"github.com/adeelahmad/snapback/internal/telemetry"
)

// scopeName is the instrumentation scope every Snapback metric is reported under.
const scopeName = "github.com/adeelahmad/snapback"

// deltaTemporality is AGGREGATION_TEMPORALITY_DELTA: each point reports what
// happened since the last export, so nothing has to be carried across sends.
const deltaTemporality = 2

// Resource identifies the sending install. Every field is a coarse, non-identifying
// string: the release version, the Go GOOS and GOARCH, and the random install id.
// There is no hostname, username or path here, and there never will be.
type Resource struct {
	// Version is the Snapback release version.
	Version string
	// OS is the GOOS value.
	OS string
	// Arch is the GOARCH value.
	Arch string
	// InstallID is the random install identifier (D6).
	InstallID string
}

// request is an OTLP ExportMetricsServiceRequest in the protobuf-JSON mapping.
type request struct {
	ResourceMetrics []resourceMetrics `json:"resourceMetrics"`
}

// resourceMetrics carries one resource and the metrics collected under it.
type resourceMetrics struct {
	Resource     resourceBody   `json:"resource"`
	ScopeMetrics []scopeMetrics `json:"scopeMetrics"`
}

// resourceBody holds the resource's attributes.
type resourceBody struct {
	Attributes []keyValue `json:"attributes"`
}

// scopeMetrics groups metrics under the instrumentation scope that produced them.
type scopeMetrics struct {
	Scope   scope    `json:"scope"`
	Metrics []metric `json:"metrics,omitempty"`
}

// scope names the instrumentation that produced the metrics.
type scope struct {
	Name string `json:"name"`
}

// metric is a single named metric; Snapback only emits sums.
type metric struct {
	Name string `json:"name"`
	Sum  sum    `json:"sum"`
}

// sum is a monotonic delta counter with its data points.
type sum struct {
	AggregationTemporality int         `json:"aggregationTemporality"`
	IsMonotonic            bool        `json:"isMonotonic"`
	DataPoints             []dataPoint `json:"dataPoints"`
}

// dataPoint is one counted occurrence. Both int64 fields are JSON strings, as the
// protobuf-JSON mapping requires.
type dataPoint struct {
	AsInt        string     `json:"asInt"`
	TimeUnixNano string     `json:"timeUnixNano"`
	Attributes   []keyValue `json:"attributes"`
}

// keyValue is an OTLP attribute. Snapback only ever sends string values.
type keyValue struct {
	Key   string     `json:"key"`
	Value stringBody `json:"value"`
}

// stringBody is the OTLP AnyValue restricted to its string case.
type stringBody struct {
	StringValue string `json:"stringValue"`
}

// attrsOf renders telemetry attributes as OTLP string-valued attributes.
func attrsOf(attrs []telemetry.Attr) []keyValue {
	out := make([]keyValue, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, keyValue{Key: a.Key, Value: stringBody{StringValue: a.Value}})
	}
	return out
}

// Encode renders events as an OTLP/HTTP ExportMetricsServiceRequest in the
// protobuf-JSON mapping: one resourceMetrics entry carrying res, one scopeMetrics
// scope, and one monotonic delta sum metric per event.
func Encode(events []telemetry.Event, res Resource) ([]byte, error) {
	metrics := make([]metric, 0, len(events))
	for _, ev := range events {
		metrics = append(metrics, metric{
			Name: ev.Name,
			Sum: sum{
				AggregationTemporality: deltaTemporality,
				IsMonotonic:            true,
				DataPoints: []dataPoint{{
					AsInt:        "1",
					TimeUnixNano: strconv.FormatInt(ev.Time.UnixNano(), 10),
					Attributes:   attrsOf(ev.Attrs),
				}},
			},
		})
	}

	req := request{ResourceMetrics: []resourceMetrics{{
		Resource: resourceBody{Attributes: []keyValue{
			{Key: "version", Value: stringBody{StringValue: res.Version}},
			{Key: "os", Value: stringBody{StringValue: res.OS}},
			{Key: "arch", Value: stringBody{StringValue: res.Arch}},
			{Key: "install_id", Value: stringBody{StringValue: res.InstallID}},
		}},
		ScopeMetrics: []scopeMetrics{{Scope: scope{Name: scopeName}, Metrics: metrics}},
	}}}

	return json.Marshal(req)
}
