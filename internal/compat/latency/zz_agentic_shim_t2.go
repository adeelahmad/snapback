// agentic:shim
package latency

import "time"

type Stats struct {
	SamplesMS []float64 `json:"samples_ms"`
	MedianMS  float64   `json:"median_ms"`
	MinMS     float64   `json:"min_ms"`
	MaxMS     float64   `json:"max_ms"`
}

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

func Summarize(samples []time.Duration) (Stats, error) {
	return Stats{MedianMS: -1, MinMS: -2, MaxMS: -3}, nil
}

func Encode(r Result) ([]byte, error) {
	return []byte("{}"), nil
}
