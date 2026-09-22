package doctor

// platformSkipReason replaces the fix text of a check that the platform
// cannot satisfy, so the report says why instead of asking for an install.
const platformSkipReason = "not applicable on macOS in v0.1"

// darwinInapplicable names the checks whose prerequisites exist only on
// Linux: the two FUSE probes and the systemd service manager probe.
var darwinInapplicable = map[string]bool{
	"fuse_device":     true,
	"fusermount3":     true,
	"service_manager": true,
}

// applyPlatform reports results adjusted for goos. On darwin the checks that
// the platform cannot pass become skips carrying platformSkipReason, unless
// strict is set, in which case they keep failing. Other platforms are left
// alone. The input slice is never modified.
func applyPlatform(results []Check, goos string, strict bool) []Check {
	if goos != "darwin" || strict {
		return results
	}
	out := make([]Check, len(results))
	copy(out, results)
	for i, c := range out {
		if c.Status == statusFail && darwinInapplicable[c.Name] {
			out[i].Status = statusSkip
			out[i].Fix = platformSkipReason
		}
	}
	return out
}

// exitCode reports 0 when no check failed and 1 otherwise. Skipped, warning
// and unavailable checks do not change the exit code.
func exitCode(results []Check) int {
	for _, c := range results {
		if c.Status == statusFail {
			return 1
		}
	}
	return 0
}
