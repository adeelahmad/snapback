// agentic:shim

// Package resticfx is a disposable restic fixture for compatibility tests.
package resticfx

// PinnedResticVersion is deliberately wrong in this shim.
const PinnedResticVersion = "0.0.0-shim"

// Guard is a compile shim for S2-03/T1.
type Guard struct {
	TempRoot string
	Home     string
}

// ParseResticVersion is a compile shim with a deliberately wrong body.
func ParseResticVersion(_ string) (string, error) {
	return "shim", nil
}

// CheckPinnedVersion is a compile shim with a deliberately wrong body.
func CheckPinnedVersion(_ string) error {
	return nil
}

// GuardRepo is a compile shim with a deliberately wrong body.
func GuardRepo(_ string, _ Guard) error {
	return nil
}
