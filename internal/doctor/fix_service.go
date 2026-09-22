package doctor

// daemonFixText returns the fix line for a daemon that is not running. Only a
// detected, supported service manager can install the login service, so every
// other platform is pointed at the foreground command instead.
func daemonFixText(goos string, managerSupported bool) string {
	if managerSupported {
		return "start it with snapback install service"
	}
	if goos == "darwin" {
		return "start it with snapback run (the login service is Linux-only for now)"
	}
	return "start it with snapback run"
}
