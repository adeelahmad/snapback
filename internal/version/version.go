// Package version reports build metadata injected via -ldflags -X.
package version

import (
	"fmt"
	"runtime"
)

// Build metadata; package-level vars so -ldflags -X can override them.
var (
	Version = "dev"
	Commit  = "none"
	Target  = runtime.GOOS + "/" + runtime.GOARCH
)

// Format renders the version line for the given build metadata.
func Format(version, commit, target string) string {
	return fmt.Sprintf("snapback %s (commit %s, target %s)\n", version, commit, target)
}

// String renders the version line for this binary's build metadata.
func String() string {
	return Format(Version, Commit, Target)
}
