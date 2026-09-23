package setup

// consentText is the fixed telemetry notice setup prints above
// OptInQuestion. It names every event Snapback can send, states that
// telemetry is off by default and that the operator supplies their own
// collector, and links the privacy page.
const consentText = `Snapback can send anonymous usage counters: setup.completed,
daemon.started, mount.ready, doctor.failed, error.

Telemetry is off by default. You supply your own collector;
nothing is sent unless you configure one.

Details: https://snapback.run/privacy
`

// ConsentText returns the fixed telemetry notice setup prints above
// OptInQuestion. It names every event Snapback can send, says the operator
// must supply their own collector and links the privacy page.
func ConsentText() string {
	return consentText
}
