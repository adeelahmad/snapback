// agentic:shim

package main

// daemonProbe reports whether a daemon holds the state dir lock. The linker
// is meant to call it at most once per process.
var daemonProbe = func(stateDir string) bool { return false }
