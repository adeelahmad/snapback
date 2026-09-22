// Package resticfx is a disposable restic fixture for compatibility tests.
package resticfx

// PinnedResticVersion is the only restic version the fixture accepts.
const PinnedResticVersion = "" // SUB-AGENT-TODO: set to "0.19.0"

// ParseResticVersion extracts X.Y.Z from the first line of `restic version` output.
func ParseResticVersion(out string) (string, error) {
	panic("SUB-AGENT-TODO: parse first line 'restic X.Y.Z compiled with ...'; return X.Y.Z, error on any other shape")
}

// CheckPinnedVersion errors unless the parsed version equals PinnedResticVersion.
func CheckPinnedVersion(out string) error {
	panic("SUB-AGENT-TODO: ParseResticVersion(out); error naming found and wanted versions unless it equals PinnedResticVersion")
}
