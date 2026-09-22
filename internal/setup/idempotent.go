package setup

import "github.com/adeelahmad/snapback/internal/config"

// Report describes the configuration file a second setup run found.
type Report struct {
	Exists bool
	Same   bool
	Lines  []string
}

// Existing reports what the configuration file at path already holds compared
// with want. It never writes: a caller that gets Same skips the save.
func Existing(path string, want *config.Config) (Report, error) {
	return Report{}, nil
}
