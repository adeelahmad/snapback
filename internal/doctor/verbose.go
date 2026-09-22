package doctor

import (
	"fmt"
	"io"
)

// printChecksVerbose writes the report of printChecks and adds the probe the
// check ran and the raw value it read under each verdict. It only ever adds
// lines, so the report without --verbose stays byte-identical.
func printChecksVerbose(w io.Writer, checks []Check) {
	for _, c := range checks {
		printCheck(w, c)
		_, _ = fmt.Fprintf(w, "  probe: %s\n", c.Probe)
		_, _ = fmt.Fprintf(w, "  observed: %s\n", c.Observed)
	}
}

// withoutVerbose returns a copy of checks with the --verbose detail cleared,
// so the default report and JSON carry neither key. checks is not modified.
func withoutVerbose(checks []Check) []Check {
	out := make([]Check, len(checks))
	copy(out, checks)
	for i := range out {
		out[i].Probe = ""
		out[i].Observed = ""
	}
	return out
}
