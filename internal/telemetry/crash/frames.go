// Package crash extracts snapback-only stack frames from a
// runtime/debug.Stack dump for use in crash reports. It keeps only the
// package path and function name of frames belonging to this module —
// never a file path, a line number or argument values.
package crash

// Frame is one retained stack frame: its package path and function name.
// It carries no file path, no line number and no argument values.
type Frame struct {
	Module   string
	Function string
}

// Frames parses a runtime/debug.Stack-style dump and returns, in order, the
// frames whose package path starts with "github.com/adeelahmad/snapback/".
// Frames from the standard library, third-party modules and the
// "goroutine N [running]:" header are dropped. Empty or garbage input
// yields nil.
func Frames(stack []byte) []Frame {
	return nil
}
