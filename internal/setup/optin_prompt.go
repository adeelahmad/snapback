package setup

import (
	"bufio"
	"io"
	"strings"
)

// OptInQuestion is the single line setup prints to ask about telemetry. The
// capital N marks the default: anything but an explicit yes leaves it off.
const OptInQuestion = "Send anonymous usage counters to a collector you configure? [y/N] "

// AskOptIn asks once whether to send anonymous usage counters. It asks only
// when interactive, and asked reports whether the question was printed. A bare
// Enter, a no and a closed input all keep the default, which is off.
func AskOptIn(in io.Reader, out io.Writer, interactive bool) (enabled bool, asked bool) {
	if !interactive {
		return false, false
	}
	_, _ = io.WriteString(out, ConsentText())
	_, _ = io.WriteString(out, OptInQuestion)
	line, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, true
	}
	return false, true
}
