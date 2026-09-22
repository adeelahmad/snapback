package latency

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"
)

// measurementKeys are the measurements every Result must carry.
var measurementKeys = []string{
	"cold_listing",
	"warm_prewarmed_listing",
	"cold_first_file_read",
	"warm_listing_after_restart",
}

// Stats holds the samples and summary figures for one measurement, in milliseconds.
type Stats struct {
	SamplesMS []float64 `json:"samples_ms"`
	MedianMS  float64   `json:"median_ms"`
	MinMS     float64   `json:"min_ms"`
	MaxMS     float64   `json:"max_ms"`
}

// Result is the JSON record of one latency run.
type Result struct {
	Measurements  map[string]Stats `json:"measurements"`
	ResticVersion string           `json:"restic_version"`
	RcloneVersion string           `json:"rclone_version"`
	OS            string           `json:"os"`
	Arch          string           `json:"arch"`
	TimestampUTC  time.Time        `json:"timestamp_utc"`
	Remote        string           `json:"remote"`
	DataFileCount int              `json:"data_file_count"`
	DataFileBytes int64            `json:"data_file_bytes"`
	RemoteDeleted bool             `json:"remote_deleted"`
}

// Summarize converts samples to milliseconds and computes median, min and max.
func Summarize(samples []time.Duration) (Stats, error) {
	if len(samples) == 0 {
		return Stats{}, errors.New("latency: no samples to summarize")
	}
	ms := make([]float64, len(samples))
	for i, d := range samples {
		ms[i] = float64(d) / float64(time.Millisecond)
	}
	sorted := slices.Clone(ms)
	slices.Sort(sorted)
	n := len(sorted)
	median := sorted[n/2]
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return Stats{SamplesMS: ms, MedianMS: median, MinMS: sorted[0], MaxMS: sorted[n-1]}, nil
}

// Encode renders r as indented JSON after checking all four measurements exist.
func Encode(r Result) ([]byte, error) {
	for _, k := range measurementKeys {
		if _, ok := r.Measurements[k]; !ok {
			return nil, fmt.Errorf("latency: result missing measurement %q", k)
		}
	}
	r.TimestampUTC = r.TimestampUTC.UTC()
	return json.MarshalIndent(r, "", "  ")
}
