// agentic:shim

package aliases

import "time"

// Options is a RED compile shim.
type Options struct {
	Local     bool
	Loc       *time.Location
	Rsnapshot bool
	Keep      Keep
}

// Keep is a RED compile shim.
type Keep struct {
	Hourly, Daily, Weekly, Monthly int
}

func renderZone(Options) *time.Location { return time.UTC }

func baseName(time.Time, *time.Location, bool) string { return "" }

func dateOf(time.Time, *time.Location) string { return "" }
