// agentic:shim

package service

// RealProbe returns the Probe wired to the running host.
func RealProbe() Probe {
	return Probe{
		PID1Comm: func() (string, error) { return "agentic-shim-init", nil },
		Exists:   func(path string) bool { return path == "/run/openrc" },
	}
}
