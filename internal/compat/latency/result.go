package latency

import "time"

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
	panic("SUB-AGENT-TODO: T2 error on empty input; SamplesMS = each sample in float64 ms (input order); median of sorted copy (even count = mean of middle two), min, max")
}

// Encode renders r as indented JSON after checking all four measurements exist.
func Encode(r Result) ([]byte, error) {
	panic("SUB-AGENT-TODO: T2 error if Measurements lacks any of cold_listing, warm_prewarmed_listing, cold_first_file_read, warm_listing_after_restart; else json.MarshalIndent with timestamp in UTC RFC 3339")
}
