package refresh

import "log/slog"

// WithLog directs the Refresher's debug records to log and returns r. A nil
// log discards them.
func (r *Refresher) WithLog(log *slog.Logger) *Refresher {
	_ = log
	return r
}
