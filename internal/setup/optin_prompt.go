package setup

import "io"

// OptInQuestion is the single line setup prints to ask about telemetry. The
// capital N marks the default: anything but an explicit yes leaves it off.
const OptInQuestion = "Send anonymous usage counters to a collector you configure? [y/N] "

// AskOptIn asks once whether to send anonymous usage counters. It asks only
// when interactive, and asked reports whether the question was printed.
func AskOptIn(in io.Reader, out io.Writer, interactive bool) (enabled bool, asked bool) {
	return false, false
}
