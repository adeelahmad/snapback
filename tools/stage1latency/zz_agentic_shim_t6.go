// agentic:shim
package main

import (
	"io"

	"github.com/adeelahmad/snapback/internal/compat/latency"
)

// newConfig builds the latency collaborators for remote; tests replace it with fakes.
var newConfig = func(_ string) latency.Config {
	return latency.Config{}
}

// run is a compile shim: it writes nothing and always returns 3.
func run(_ []string, _ func(string) string, _, _ io.Writer) int {
	return 3
}

func main() {}
