package aliases

import "time"

// renderZone returns the zone names and dates are rendered in: opts.Loc when
// opts.Local is set and Loc is non-nil, otherwise UTC. Never time.Local.
func renderZone(opts Options) *time.Location {
	if opts.Local && opts.Loc != nil {
		return opts.Loc
	}
	return time.UTC
}

// baseName renders t in loc as the minute-precision alias name.
func baseName(t time.Time, loc *time.Location, local bool) string {
	if local {
		return t.In(loc).Format("2006-01-02_1504-0700")
	}
	return t.In(loc).Format("2006-01-02_1504Z")
}

// dateOf renders t in loc as a 2006-01-02 date.
func dateOf(t time.Time, loc *time.Location) string {
	return t.In(loc).Format("2006-01-02")
}
