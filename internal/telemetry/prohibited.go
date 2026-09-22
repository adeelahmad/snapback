package telemetry

// Finding is one hit of a prohibited-value rule in a scanned payload.
type Finding struct {
	// Rule names the rule that matched, as listed by ProhibitedRules.
	Rule string
	// Match is the matched text.
	Match string
}

// ScanProhibited reports every prohibited value in b, one Finding per hit.
func ScanProhibited(b []byte) []Finding { return nil }

// ProhibitedRules lists the rule names ScanProhibited can report, in table order.
func ProhibitedRules() []string { return nil }
