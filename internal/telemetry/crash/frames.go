// Package crash extracts snapback-only stack frames from a
// runtime/debug.Stack dump for use in crash reports. It keeps only the
// package path and function name of frames belonging to this module —
// never a file path, a line number or argument values.
package crash

import "strings"

// snapbackPrefix is the module import-path prefix a frame's package path
// must have to be kept.
const snapbackPrefix = "github.com/adeelahmad/snapback/"

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
	var frames []Frame
	for _, line := range strings.Split(string(stack), "\n") {
		if line == "" || strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "goroutine ") {
			continue
		}
		frame, ok := parseCallLine(line)
		if !ok || !strings.HasPrefix(frame.Module, snapbackPrefix) {
			continue
		}
		frames = append(frames, frame)
	}
	return frames
}

// parseCallLine parses a single function-call line of a runtime/debug.Stack
// dump, e.g. "pkg/path.(*Recv).Method(args)", into a Frame with the
// argument list dropped. It reports false for lines that do not have the
// expected "<prefix>(<args>)" shape.
func parseCallLine(line string) (Frame, bool) {
	open := strings.LastIndex(line, "(")
	if open < 0 {
		return Frame{}, false
	}
	prefix := line[:open]

	pathEnd := strings.LastIndex(prefix, "/")
	rest := prefix[pathEnd+1:]

	dot := strings.Index(rest, ".")
	if dot < 0 {
		return Frame{}, false
	}
	packageName, function := rest[:dot], rest[dot+1:]
	if packageName == "" || function == "" {
		return Frame{}, false
	}

	module := packageName
	if pathEnd >= 0 {
		module = prefix[:pathEnd+1] + packageName
	}
	return Frame{Module: module, Function: function}, true
}
