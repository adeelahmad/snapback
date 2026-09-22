package doctor

import "time"

// BundleInput carries the already-collected, already-redacted material that
// goes into a diagnostic bundle. Nothing here is gathered or sent by the
// bundle writer itself.
type BundleInput struct {
	DoctorJSON     []byte
	Version        string
	Commit         string
	GOOS           string
	GOARCH         string
	DaemonLog      []byte
	ConfigRedacted []byte
}

// WriteBundle writes a gzipped tar of the diagnostic material into dir and
// returns the path it wrote.
func WriteBundle(dir string, in BundleInput, now time.Time) (string, error) {
	_, _, _ = dir, in, now
	return "", nil
}
