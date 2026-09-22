package aliases

import "time"

// Options controls how snapshot aliases are rendered.
type Options struct {
	// Local renders names in Loc with a numeric offset instead of UTC.
	Local bool
	// Loc is the render zone when Local is true; nil means UTC.
	Loc *time.Location
	// Rsnapshot enables the hourly, daily, weekly and monthly views.
	Rsnapshot bool
	// Keep sets how many snapshots each rsnapshot view keeps.
	Keep Keep
}

// Keep is the per-kind keep count for rsnapshot views; zero disables a kind.
type Keep struct {
	Hourly, Daily, Weekly, Monthly int
}
