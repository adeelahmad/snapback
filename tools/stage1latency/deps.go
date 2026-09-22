package main

import "github.com/adeelahmad/snapback/internal/compat/latency"

// newConfig builds the real latency collaborators for remote; tests replace it with fakes.
var newConfig = func(remote string) latency.Config {
	panic("SUB-AGENT-TODO: wire resticfx runner, mounter and time.Now clock")
}
