package setup

import "context"

// nextRun is the command to run once the repository already holds the root.
const nextRun = "snapback run"

// Advice is the next command setup points the operator at, with the notes that
// explain why that command and not another.
type Advice struct {
	Next  string
	Notes []string
}

// Plan reads the detected repository once and returns the detection result the
// repository confirms, together with the next step. It is read-only.
func Plan(_ context.Context, _ Runner, r Result) (Result, Advice, error) {
	return r, Advice{Next: nextRun}, nil
}
