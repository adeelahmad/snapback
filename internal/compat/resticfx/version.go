// Package resticfx is a disposable restic fixture for compatibility tests.
package resticfx

import (
	"fmt"
	"regexp"
	"strings"
)

// PinnedResticVersion is the only restic version the fixture accepts.
const PinnedResticVersion = "0.19.0"

var resticVersionRe = regexp.MustCompile(`^restic (\d+\.\d+\.\d+)(\s|$)`)

// ParseResticVersion extracts X.Y.Z from the first line of `restic version` output.
func ParseResticVersion(out string) (string, error) {
	first, _, _ := strings.Cut(out, "\n")
	m := resticVersionRe.FindStringSubmatch(strings.TrimSpace(first))
	if m == nil {
		return "", fmt.Errorf("unrecognized restic version output %q", first)
	}
	return m[1], nil
}

// CheckPinnedVersion errors unless the parsed version equals PinnedResticVersion.
func CheckPinnedVersion(out string) error {
	v, err := ParseResticVersion(out)
	if err != nil {
		return err
	}
	if v != PinnedResticVersion {
		return fmt.Errorf("restic version %s found, want %s", v, PinnedResticVersion)
	}
	return nil
}
