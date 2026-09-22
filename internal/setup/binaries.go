package setup

import "path/filepath"

// applyBinaries records which of the restic and rclone binaries lookPath finds,
// mirroring the lookups the web Setup form does. A missing restic is a named
// reason; rclone is optional, so a missing rclone is silent. A nil lookPath
// behaves as if neither binary were on PATH.
func applyBinaries(r *Result, lookPath func(string) (string, error)) {
	if lookPath == nil {
		r.Reasons = append(r.Reasons, "restic: not found on PATH")
		return
	}
	if p, err := lookPath("restic"); err == nil {
		if !filepath.IsAbs(p) {
			if abs, err := filepath.Abs(p); err == nil {
				p = abs
			}
		}
		r.ResticPath = p
	} else {
		r.Reasons = append(r.Reasons, "restic: not found on PATH")
	}
	if _, err := lookPath("rclone"); err == nil {
		r.RcloneFound = true
	}
}
