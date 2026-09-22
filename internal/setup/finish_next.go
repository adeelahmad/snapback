package setup

import (
	"path/filepath"
	"strings"
)

// snapPrefix starts the advice that must survive every other outcome: a
// repository holding nothing for the root needs a first snapshot before any
// history is there to browse.
const snapPrefix = "snapback snap "

// linkName is the entry Snapback places in every linked directory.
const linkName = ".snapshot"

// NextAfterSetup names the one command setup ends on. Advice to snap wins,
// because nothing can be browsed yet. Otherwise a history that is already
// served — by the login service setup installed on Linux, or by a daemon
// already running — points at the first linked root; anything else has to
// start the daemon by hand.
func NextAfterSetup(goos string, serviceInstalled, daemonRunning bool, linked []string, advice Advice) string {
	if strings.HasPrefix(advice.Next, snapPrefix) {
		return advice.Next
	}
	served := daemonRunning || (goos == "linux" && serviceInstalled)
	if served && len(linked) > 0 {
		return "ls " + filepath.Join(linked[0], linkName)
	}
	return nextRun
}
